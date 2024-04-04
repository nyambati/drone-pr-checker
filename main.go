package main

import (
	"fmt"
	"net/url"

	"github.com/nyambati/drone-pr-checker/internal"
)

func main() {

	repo := internal.GetEnvVar("DRONE_REPO", "")
	pull_request := internal.GetEnvVar("DRONE_PULL_REQUEST", "")
	token := internal.GetEnvVar("GITHUB_TOKEN", "")

	url := &url.URL{
		Scheme: "https",
		Host:   "api.github.com",
		Path:   fmt.Sprintf("/repos/%s/pulls/%s", repo, pull_request),
	}

	prc := internal.NewPlugin(url, token)

	prc.CheckPRLabels().
		CheckPRTitlePrefixes().
		CheckPRTitleRegexep().
		CheckPRChecklist().
		Report()

}
