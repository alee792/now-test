package wonder

import (
	"context"
	"fmt"

	"github.com/google/go-github/v18/github"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
)

// Config for Client.
type Config struct {
	OAuth2Token    string `envconfig:"GITHUB_OA2"`
	MaxConcurrency int    `envconfig:"MAX_CONCURRENCY" default:"20"`
}

// repoer to handle repo interactions.
type repoer interface {
	ListCommits(ctx context.Context, owner, repo string, opt *github.CommitsListOptions) ([]*github.RepositoryCommit, *github.Response, error)
	GetCommit(ctx context.Context, owner, repo, sha string) (*github.RepositoryCommit, *github.Response, error)
}

// Client used to access and process GitHub repositories.
type Client struct {
	repoer
	Logger *zap.SugaredLogger
	Config
}

// Repo holds data relevant to a Git repository.
type Repo struct {
	*github.Repository
	Users map[int64]*User // Key: ID, Value: User
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

// NewClient provides a Wonder Client with sensible defaults.
func NewClient(cfg Config) *Client {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: cfg.OAuth2Token},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)
	return &Client{
		repoer: client.Repositories,
		Logger: zap.NewExample().Sugar(),
		Config: cfg,
	}
}

// GetCommits from a repo.
func (c *Client) GetCommits(ctx context.Context, owner, repo string) (map[string]Commit, error) {
	opt := &github.CommitsListOptions{
		ListOptions: github.ListOptions{
			PerPage: 500,
		},
	}
	var repoCommits = make(map[string]Commit)
	for {
		commits, resp, err := c.ListCommits(ctx, owner, repo, opt)
		if err != nil {
			return nil, errors.Wrap(err, "ListCommits failed")
		}
		// Bulk queries do not return all stats, so we must query each commit.
		for _, cmt := range commits {
			rcmt, _, err := c.GetCommit(ctx, owner, repo, cmt.GetSHA())
			if err != nil {
				return nil, errors.Wrap(err, "GetCommit failed")
			}
			sha := rcmt.GetSHA()
			if _, ok := repoCommits[sha]; ok {
				return nil, fmt.Errorf("duplicate sha: %s", sha)
			}
			repoCommits[sha] = Commit{
				commit: rcmt,
			}
		}
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}
	return repoCommits, nil
}

// GetUsers from commits.
func (c *Client) GetUsers(ctx context.Context, commits map[string]Commit) (map[int64]*User, error) {
	var users = make(map[int64]*User)
	for _, cmt := range commits {
		a := cmt.commit.GetAuthor()
		if a == nil {
			continue
		}
		id := a.GetID()
		if id == 0 {
			continue
		}
		u, ok := users[id]
		if !ok {
			u = &User{
				user:  a,
				Name:  a.GetName(),
				Login: a.GetLogin(),
			}
			users[id] = u
		}
		u.aggregateStats(cmt)
	}
	return users, nil
}

func (u *User) aggregateStats(cmt Commit) {
	stats := cmt.commit.GetStats()
	u.Additions += stats.GetAdditions()
	u.Deletions += stats.GetDeletions()
	u.Total += stats.GetTotal()
	u.Commits++
}
