package internal

import (
	"encoding/json"
	"net/http"
	"net/url"
)

type GitHub interface {
	getPRLabels() ([]string, error)
}

type Client struct {
	Http  *http.Client
	URL   *url.URL
	token string
}

type Label struct {
	Name string `json:"name"`
}

type PR struct {
	Labels []Label `json:"labels"`
}

func (c *Client) getPRLabels() ([]string, error) {
	request, err := http.NewRequest("GET", c.URL.String(), nil)
	request.Header.Set("Authorization", "Bearer "+c.token)
	request.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	request.Header.Set("Accept", "application/vnd.github+json")

	if err != nil {
		return nil, err
	}

	resp, err := c.Http.Do(request)

	if err != nil {
		return nil, err
	}

	pr := PR{}

	if err := json.NewDecoder(resp.Body).Decode(&pr); err != nil {
		return nil, err
	}

	labels := []string{}
	for _, label := range pr.Labels {
		labels = append(labels, label.Name)
	}

	return labels, nil
}

func NewGithub(url *url.URL, token string) GitHub {
	return &Client{
		Http:  &http.Client{},
		URL:   url,
		token: token,
	}
}
