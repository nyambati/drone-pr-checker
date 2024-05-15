package github

import (
	"context"

	"github.com/google/go-github/v61/github"
)

type GitHub struct {
	client *github.Client
}

func (g *GitHub) GetPullRequest(owner string, repo string, number int) (*github.PullRequest, error) {
	pr, _, err := g.client.PullRequests.Get(context.Background(), owner, repo, number)
	return pr, err
}

func New(token string) GitHubInterface {
	return &GitHub{
		client: github.NewClient(nil).WithAuthToken(token),
	}
}
