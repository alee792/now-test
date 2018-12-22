package integration

import (
	"context"
	"os"
	"testing"

	"github.com/alee792/wonder/pkg/wonder"
)

func Test_GetCommits(t *testing.T) {
	c := wonder.NewClient(wonder.Config{
		OAuth2Token: os.Getenv("GITHUB_OA2"),
	})
	ctx := context.Background()
	cc, err := c.GetCommits(ctx, "alee792", "wonder")
	if err != nil {
		c.Logger.Error(err)
	}
	c.Logger.Infof("Commits: %+v", cc)
}

func Test_GetUsers(t *testing.T) {
	c := wonder.NewClient(wonder.Config{
		OAuth2Token: os.Getenv("GITHUB_OA2"),
	})
	ctx := context.Background()
	cc, err := c.GetCommits(ctx, "alee792", "wonder")
	if err != nil {
		t.Error(err)
	}
	uu, err := c.GetUsers(ctx, cc)
	if err != nil {
		t.Error(err)
	}
	for _, u := range uu {
		c.Logger.Infof("Commits: %+v", u)
	}
}
