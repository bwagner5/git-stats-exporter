package repos

import (
	"context"
	"net/http"
	"time"

	"github.com/google/go-github/v45/github"
	"github.com/prometheus/client_golang/prometheus"
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
	Stars       int
	Forks       int
	Subscribers int
}

func init() {
	metrics.Registry.MustRegister(issuesGauge, starsGauge, forksGauge, subscribersGauge)
}

func New() *Repos {
	return &Repos{client: github.NewClient(&http.Client{Timeout: 15 * time.Second})}
}

func (r *Repos) EmitMetrics(ctx context.Context, owner string, repoName string) error {
	ghRepo, err := r.GetMetrics(ctx, owner, repoName)
	if err != nil {
		return err
	}
	repoLabels := prometheus.Labels{"owner": owner, "repo": repoName}
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
	return &RepoMetrics{
		OpenIssues:  ghRepo.GetOpenIssuesCount(),
		Stars:       ghRepo.GetStargazersCount(),
		Forks:       ghRepo.GetForksCount(),
		Subscribers: ghRepo.GetSubscribersCount(),
	}, nil
}
