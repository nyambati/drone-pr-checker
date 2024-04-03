package internal

import (
	"encoding/json"
	"net/http"
	"net/url"
)

type GitHub interface {
	GetPullRequest() (PullRequest, error)
}

type Client struct {
	Http  *http.Client
	URL   *url.URL
	token string
}

type Label struct {
	Name string `json:"name"`
}

type PullRequest struct {
	Labels []Label `json:"labels"`
	Body   string  `json:"body"`
}

func (c *Client) GetPullRequest() (PullRequest, error) {
	request, err := http.NewRequest("GET", c.URL.String(), nil)
	request.Header.Set("Authorization", "Bearer "+c.token)
	request.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	request.Header.Set("Accept", "application/vnd.github+json")

	pr := PullRequest{}

	if err != nil {
		return pr, err
	}

	resp, err := c.Http.Do(request)

	if err != nil {
		return pr, err
	}

	if err := json.NewDecoder(resp.Body).Decode(&pr); err != nil {
		return pr, err
	}

	return pr, nil

}

func NewGithub(url *url.URL, token string) GitHub {
	return &Client{
		Http:  &http.Client{},
		URL:   url,
		token: token,
	}
}
