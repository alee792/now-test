package wonder

import (
	"context"
	"sync"
	"time"

	"github.com/google/go-github/github"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
)

// Config for Server.
type Config struct {
	OAuth2Token    string `envconfig:"GITHUB_OA2"`
	MaxConcurrency int64  `envconfig:"MAX_CONCURRENCY" default:"20"`
	SinceDays      int    `envconfig:"SINCE_DAYS" default:"45"`
}

// RepoService to handle repo interactions.
type RepoService interface {
	Get(ctx context.Context, owner, repo string) (*github.Repository, *github.Response, error)
	ListCommits(ctx context.Context, owner, repo string, opt *github.CommitsListOptions) ([]*github.RepositoryCommit, *github.Response, error)
	GetCommit(ctx context.Context, owner, repo, sha string) (*github.RepositoryCommit, *github.Response, error)
}

// Server used to access and process GitHub repositories.
type Server struct {
	RepoService
	Logger *zap.SugaredLogger
	Config
}

// Repo holds data relevant to a Git repository.
type Repo struct {
	*github.Repository
	Users   *sync.Map // Key: ID, Value: User
	Commits *sync.Map // Key: SHA, Value: Commit
}

// User shows user's stats.
type User struct {
	Name      string
	Login     string
	Email     string
	Commits   int
	Additions int
	Deletions int
	Total     int
	user      *github.User
}

// Commit holds data relevant to a Git commit.
type Commit struct {
	commit *github.RepositoryCommit
}

// NewClient provides a Wonder Server with sensible defaults.
func NewClient(cfg Config) *Server {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: cfg.OAuth2Token},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)
	return &Server{
		RepoService: client.Repositories,
		Logger:      zap.NewExample().Sugar(),
		Config:      cfg,
	}
}

func (s *Server) ProcessRepo(ctx context.Context, owner, repoName string) (*Repo, error) {
	gitRepo, _, err := s.RepoService.Get(ctx, owner, repoName)
	if err != nil {
		return nil, errors.Wrap(err, "get repo failed")
	}
	repo := &Repo{
		Repository: gitRepo,
		Users:      &sync.Map{},
		Commits:    &sync.Map{},
	}

	group, ctx := errgroup.WithContext(ctx)

	cmtC := make(chan *Commit)
	group.Go(func() error {
		err = s.getCommits(ctx, owner, repoName, cmtC)
		if err != nil {
			return errors.Wrap(err, "getCommits failed")
		}
		return nil
	})
	group.Go(func() error {
		for cmt := range cmtC {
			u := cmt.user()
			stored, ok := repo.Users.LoadOrStore(u.user.GetID(), u)
			if ok {
				u, ok = stored.(*User)
				if !ok {
					return errors.New("*User type assertion failed")
				}
			}
			// Not loading because SHAs should be unique.
			repo.Commits.Store(cmt.commit.GetSHA(), cmt)
			u.aggregateStats(cmt)
		}
		return nil
	})

	err = group.Wait()
	if err != nil {
		return nil, err
	}

	return repo, nil
}

func (s *Server) getCommits(ctx context.Context, owner, repoName string, cmtC chan<- *Commit) error {
	sem := semaphore.NewWeighted(s.MaxConcurrency)
	group, ctx := errgroup.WithContext(ctx)

	opt := &github.CommitsListOptions{
		ListOptions: github.ListOptions{
			PerPage: 500,
		},
		Since: time.Now().Truncate(24 * time.Hour).
			Add(-time.Duration(s.Config.SinceDays)),
	}

	// Bulk commit queries do not return all stats,
	// so we must query each commit individually.

	// Get commit SHAs.
	shaC := make(chan string)
	cc, resp, err := s.ListCommits(ctx, owner, repoName, opt)
	if err != nil {
		return errors.Wrap(err, "ListCommits failed")
	}
	for _, c := range cc {
		shaC <- c.GetSHA()
	}
	group.Go(func() error {
		for i := 0; i < resp.LastPage; i++ {
			sem.Acquire(ctx, 1)
			group.Go(func() error {
				defer sem.Release(1)
				cc, _, err := s.ListCommits(ctx, owner, repoName, opt)
				if err != nil {
					return errors.Wrap(err, "ListCommits failed")
				}
				for _, c := range cc {
					shaC <- c.GetSHA()
				}
				return nil
			})
		}
		return nil
	})

	// Get individual commits.
	group.Go(func() error {
		for sha := range shaC {
			sem.Acquire(ctx, 1)
			group.Go(func() error {
				defer sem.Release(1)
				fullCmt, _, err := s.GetCommit(ctx, owner, repoName, sha)
				if err != nil {
					return errors.Wrap(err, "GetCommit failed")
				}
				cmtC <- &Commit{commit: fullCmt}
				return nil
			})
		}
		return nil
	})

	return group.Wait()
}

func (c *Commit) user() *User {
	a := c.commit.GetAuthor()
	if a == nil {
		return nil
	}
	id := a.GetID()
	if id == 0 {
		return nil
	}
	return &User{
		user:  a,
		Name:  a.GetName(),
		Login: a.GetLogin(),
	}
}

func (u *User) aggregateStats(cmt *Commit) {
	stats := cmt.commit.GetStats()
	u.Additions += stats.GetAdditions()
	u.Deletions += stats.GetDeletions()
	u.Total += stats.GetTotal()
	u.Commits++
}
