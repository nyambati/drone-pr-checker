package internal

import (
	"net/http"
	"net/url"
	"reflect"
	"testing"

	"github.com/h2non/gock"
	"github.com/stretchr/testify/assert"
)

func TestNewGithub(t *testing.T) {
	url := &url.URL{
		Scheme: "http",
		Host:   "localhost",
		Path:   "/repos/sample/pulls/1", // Adjusted Path
	}
	github := NewGithub(url, "token")
	assert.Equal(t, github, &Client{Http: &http.Client{}, URL: url, token: "token"})
}

func TestClient_getPRLabels(t *testing.T) {
	defer gock.Off() // Flush pending mocks after test execution
	gock.New("http://localhost").
		MatchHeader("Authorization", "Bearer token").
		MatchHeader("Accept", "application/vnd.github+json").
		MatchHeader("X-GitHub-Api-Version", "2022-11-28").
		Get("/repos/sample/pulls/1").
		Reply(200).
		JSON(map[string]interface{}{"labels": []Label{{Name: "label1"}, {Name: "label2"}}})

	github := NewGithub(
		&url.URL{
			Scheme: "http",
			Host:   "localhost",
			Path:   "/repos/sample/pulls/1", // Adjusted Path
		},
		"token",
	)

	labels, err := github.getPRLabels()

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expectedLabels := []string{"label1", "label2"}
	if !reflect.DeepEqual(labels, expectedLabels) {
		t.Errorf("Expected labels %v, but got %v", expectedLabels, labels)
	}
}
