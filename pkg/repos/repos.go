package repos

import (
	"context"
	"net/http"
	"time"

	"github.com/google/go-github/v45/github"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/samber/lo"
	"golang.org/x/oauth2"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

var (
	repoLabelNames = []string{"owner", "repo"}
	issuesGauge    = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gh_repo_open_issues",
			Help: "Number of open issues",
		},
		repoLabelNames,
	)
	prsGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gh_repo_open_pull_requests",
			Help: "Number of open pull requests",
		},
		repoLabelNames,
	)
	starsGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gh_repo_stars",
			Help: "Number of stars",
		},
		repoLabelNames,
	)
	forksGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gh_repo_forks",
			Help: "Number of forks",
		},
		repoLabelNames,
	)
	subscribersGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gh_repo_subscribers",
			Help: "Number of subscribers",
		},
		repoLabelNames,
	)
)

type Repos struct {
	client *github.Client
}

type RepoMetrics struct {
	OpenIssues  int
	OpenPRs     int
	Stars       int
	Forks       int
	Subscribers int
}

func init() {
	metrics.Registry.MustRegister(prsGauge, issuesGauge, starsGauge, forksGauge, subscribersGauge)
}

func New(ctx context.Context, ghToken []byte) *Repos {
	var client *github.Client
	if ghToken != nil {
		ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: string(ghToken)})
		client = github.NewClient(oauth2.NewClient(ctx, ts))
	} else {
		client = github.NewClient(&http.Client{Timeout: 15 * time.Second})
	}
	return &Repos{client: client}
}

func (r *Repos) EmitMetrics(ctx context.Context, owner string, repoName string) error {
	ghRepo, err := r.GetMetrics(ctx, owner, repoName)
	if err != nil {
		return err
	}
	repoLabels := prometheus.Labels{"owner": owner, "repo": repoName}
	prsGauge.With(repoLabels).Set(float64(ghRepo.OpenPRs))
	issuesGauge.With(repoLabels).Set(float64(ghRepo.OpenIssues))
	starsGauge.With(repoLabels).Set(float64(ghRepo.Stars))
	forksGauge.With(repoLabels).Set(float64(ghRepo.Forks))
	subscribersGauge.With(repoLabels).Set(float64(ghRepo.Subscribers))
	return nil
}

func (r *Repos) GetMetrics(ctx context.Context, owner string, repoName string) (*RepoMetrics, error) {
	ghRepo, _, err := r.client.Repositories.Get(ctx, owner, repoName)
	if err != nil {
		return nil, err
	}
	prs, err := r.getPRs(ctx, owner, repoName)
	if err != nil {
		return nil, err
	}
	issues, err := r.getIssues(ctx, owner, repoName)
	if err != nil {
		return nil, err
	}
	return &RepoMetrics{
		OpenIssues:  len(issues),
		OpenPRs:     len(prs),
		Stars:       ghRepo.GetStargazersCount(),
		Forks:       ghRepo.GetForksCount(),
		Subscribers: ghRepo.GetSubscribersCount(),
	}, nil
}

func (r *Repos) getPRs(ctx context.Context, owner string, repoName string) ([]*github.PullRequest, error) {
	var allPRs []*github.PullRequest
	opt := &github.PullRequestListOptions{
		State:       "open",
		ListOptions: github.ListOptions{PerPage: 30},
	}
	// get all pages of results
	for {
		prs, resp, err := r.client.PullRequests.List(ctx, owner, repoName, opt)
		if err != nil {
			return nil, err
		}
		allPRs = append(allPRs, prs...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}
	return allPRs, nil
}

func (r *Repos) getIssues(ctx context.Context, owner string, repoName string) ([]*github.Issue, error) {
	var allIssues []*github.Issue
	opt := &github.IssueListByRepoOptions{
		State:       "open",
		ListOptions: github.ListOptions{PerPage: 30},
	}
	// get all pages of results
	for {
		issues, resp, err := r.client.Issues.ListByRepo(ctx, owner, repoName, opt)
		if err != nil {
			return nil, err
		}

		allIssues = append(allIssues, lo.Filter(issues, func(i *github.Issue, _ int) bool { return !i.IsPullRequest() })...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}
	return allIssues, nil
}
