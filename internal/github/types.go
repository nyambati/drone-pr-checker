package github

import "github.com/google/go-github/v61/github"

type GitHubInterface interface {
	GetPullRequest(owner string, repo string, number int) (*github.PullRequest, error)
}
