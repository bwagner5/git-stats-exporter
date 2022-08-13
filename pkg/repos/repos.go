package repos

import (
	"context"
	"fmt"
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
	issuesGauge    = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "gh_repo_open_issues",
		Help: "Number of open issues",
	},
		repoLabelNames,
	)
	issuesHist = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "gh_repo_issues_duration",
		Help:    "Duration from issues open to close in minutes",
		Buckets: []float64{60, 1_440, 10_080, 20_160, 40_320, 80_640, 161_280},
	}, repoLabelNames,
	)
	prsGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "gh_repo_open_pull_requests",
		Help: "Number of open pull requests",
	},
		repoLabelNames,
	)
	starsGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "gh_repo_stars",
		Help: "Number of stars",
	},
		repoLabelNames,
	)
	forksGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "gh_repo_forks",
		Help: "Number of forks",
	},
		repoLabelNames,
	)
	subscribersGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
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
	OpenIssues     int
	IssueDurations []time.Duration
	OpenPRs        int
	Stars          int
	Forks          int
	Subscribers    int
}

func init() {
	metrics.Registry.MustRegister(prsGauge, issuesGauge, issuesHist, starsGauge, forksGauge, subscribersGauge)
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

func (r *Repos) EmitMetrics(ctx context.Context, owner string, repoName string, lastUpdated time.Time) error {
	ghRepo, err := r.GetMetrics(ctx, owner, repoName, lastUpdated)
	if err != nil {
		return err
	}
	repoLabels := prometheus.Labels{"owner": owner, "repo": repoName}
	prsGauge.With(repoLabels).Set(float64(ghRepo.OpenPRs))
	issuesGauge.With(repoLabels).Set(float64(ghRepo.OpenIssues))
	for _, issueDuration := range ghRepo.IssueDurations {
		issuesHist.With(repoLabels).Observe(issueDuration.Minutes())
	}
	starsGauge.With(repoLabels).Set(float64(ghRepo.Stars))
	forksGauge.With(repoLabels).Set(float64(ghRepo.Forks))
	subscribersGauge.With(repoLabels).Set(float64(ghRepo.Subscribers))
	fmt.Printf("%+v", ghRepo)
	return nil
}

func (r *Repos) GetMetrics(ctx context.Context, owner string, repoName string, lastUpdated time.Time) (*RepoMetrics, error) {
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
	openIssues := lo.Filter(issues, func(issue *github.Issue, _ int) bool { return issue.GetState() == "open" })
	closedIssuesSinceLastUpdate := lo.Filter(issues, func(issue *github.Issue, _ int) bool {
		return issue.GetState() == "closed" && issue.GetClosedAt().UnixMilli() > lastUpdated.UnixMilli()
	})
	var issueDurations []time.Duration
	for _, issue := range closedIssuesSinceLastUpdate {
		issueDurations = append(issueDurations, issue.ClosedAt.Sub(*issue.CreatedAt))
	}
	return &RepoMetrics{
		OpenIssues:     len(openIssues),
		IssueDurations: issueDurations,
		OpenPRs:        len(prs),
		Stars:          ghRepo.GetStargazersCount(),
		Forks:          ghRepo.GetForksCount(),
		Subscribers:    ghRepo.GetSubscribersCount(),
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
		State:       "all",
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
