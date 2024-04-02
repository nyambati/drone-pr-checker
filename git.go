package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

type Github struct {
	Client    *http.Client
	Repo      string
	PR        string
	RepoOwner string
	BaseURL   string
}

type PR struct {
	Labels []string `json:"labels"`
}

func (g Github) getPRLabels() ([]string, error) {
	request := http.Request{
		Method: http.MethodGet,
		URL: &url.URL{
			Scheme: "https",
			Host:   "api.github.com",
			Path:   fmt.Sprintf("/repos/%s/%s/pulls/%s/", g.RepoOwner, g.Repo, g.PR),
		},
		Header: http.Header{
			"Accept":               []string{"application/vnd.github.v3+json"},
			"Authorization":        []string{"Bearer " + getEnvVar("GITHUB_TOKEN", "")},
			"X-GitHub-Api-Version": []string{"2022-11-28"},
		},
	}

	resp, err := g.Client.Do(&request)

	if err != nil {
		return nil, err
	}

	pr := PR{}

	if err := json.NewDecoder(resp.Body).Decode(&pr); err != nil {
		return nil, err
	}

	return pr.Labels, nil
}
