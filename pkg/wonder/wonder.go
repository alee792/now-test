package wonder

import (
	"context"
	"fmt"
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
	Users   map[int64]*User    // Key: ID, Value: User
	Commits map[string]*Commit // Key: SHA, Value: Commit
	Since   time.Time
	Until   time.Time
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

// ProcessRepo aggregates a repositories user stats.
func (s *Server) ProcessRepo(ctx context.Context, owner, repoName string) (*Repo, error) {
	gitRepo, _, err := s.RepoService.Get(ctx, owner, repoName)
	if err != nil {
		return nil, errors.Wrap(err, "get repo failed")
	}
	repo := &Repo{
		Repository: gitRepo,
		Users:      make(map[int64]*User),
		Commits:    make(map[string]*Commit),
		Since: time.Now().Truncate(24 * time.Hour).
			Add(-time.Duration(s.Config.SinceDays)),
	}

	group, ctx := errgroup.WithContext(ctx)

	// Obtain commits for specified repo.
	cmtC := make(chan *Commit)
	group.Go(func() error {
		err = s.getCommits(ctx, owner, repoName, cmtC)
		if err != nil {
			return errors.Wrap(err, "getCommits failed")
		}
		return nil
	})

	// Populate Users and Commits maps.
	// Aggregate user statistics.
	group.Go(func() error {
		for cmt := range cmtC {
			u := cmt.user()
			if u == nil {
				fmt.Printf("%+v", cmt.commit)
				continue
			}
			id := u.user.GetID()
			stored, ok := repo.Users[id]
			if ok {
				u = stored
			}
			// Not checking for existing commit because SHAs should be unique.
			repo.Commits[cmt.commit.GetSHA()] = cmt
			u.aggregateStats(cmt)
			repo.Users[id] = u
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
	defer close(cmtC)
	sem := semaphore.NewWeighted(s.MaxConcurrency)
	group, ctx := errgroup.WithContext(ctx)

	opt := &github.CommitsListOptions{
		ListOptions: github.ListOptions{
			PerPage: 500,
		},
		Since: time.Now().Truncate(24 * time.Hour).
			Add(-time.Duration(s.Config.SinceDays) * 24 * time.Hour),
	}

	// Bulk commit queries do not return all stats,
	// so we must query each commit individually.

	// Get commit SHAs.
	shaC := make(chan string)
	var shaWG sync.WaitGroup

	cc, resp, err := s.ListCommits(ctx, owner, repoName, opt)
	if err != nil {
		return errors.Wrap(err, "ListCommits failed")
	}

	shaWG.Add(1)
	go func() {
		for _, c := range cc {
			shaC <- c.GetSHA()
		}
		defer shaWG.Done()
	}()
	group.Go(func() error {
		defer close(shaC)
		for i := 0; i < resp.LastPage; i++ {
			sem.Acquire(ctx, 1)
			shaWG.Add(1)
			group.Go(func() error {
				defer sem.Release(1)
				defer shaWG.Done()
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
		shaWG.Wait()
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
		ca := c.commit.Commit.GetAuthor()
		return &User{
			Name:  ca.GetName(),
			Login: ca.GetLogin(),
		}
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
