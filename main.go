package main

import (
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"regexp"
	"strings"
)

type State int

const (
	Err State = iota
	Success
	Skip
)

type Step struct {
	status  State
	message string
	id      string
}

type Settings struct {
	titlePrefixes    string
	titleRegexep     string
	ignoreOnLabels   string
	checklist        bool
	pullRequestTitle string
	checkListTitle   string
}

type PullRequestChecker struct {
	steps    []Step
	errors   int
	settings Settings
	github   *Github
}

func (prc *PullRequestChecker) ReadPRCheckerSettings() *PullRequestChecker {
	prefixes := getEnvVar("PLUGIN_PREFIXES", "")
	labels := getEnvVar("PLUGIN_IGNORE_ON_LABELS", "")
	checklist := getEnvVar("PLUGIN_CHECKLIST", "false") == "true"
	regex := getEnvVar("PLUGIN_TITLE_REGEXEP", "")
	pullRequestTitle := getEnvVar("PULL_REQUEST_TITLE", "")
	checkListTitle := getEnvVar("PLUGIN_CHECKLIST_TITLE", "## Checklist")

	prc.settings = Settings{
		titlePrefixes:    prefixes,
		titleRegexep:     regex,
		ignoreOnLabels:   labels,
		checklist:        checklist,
		pullRequestTitle: pullRequestTitle,
		checkListTitle:   checkListTitle,
	}
	return prc
}

func (prc *PullRequestChecker) CheckPRTitlePrefixes() *PullRequestChecker {
	if prc.settings.titlePrefixes == "" {
		prc.steps = append(prc.steps, Step{status: Skip, message: "No prefixes to check", id: "prefix"})
		return prc
	}

	prefixes := strings.Split(prc.settings.titlePrefixes, ",")

	for _, prefix := range prefixes {
		if strings.HasPrefix(prc.settings.pullRequestTitle, prefix) {
			prc.steps = append(prc.steps, Step{status: Success, message: "PR title has required prefix", id: "prefix"})
			return prc
		}
	}

	prc.steps = append(
		prc.steps,
		Step{
			status:  Err,
			id:      "prefix",
			message: fmt.Sprintf("PR title does not have any required prefix %s", prc.settings.titlePrefixes),
		},
	)
	return prc
}

func (prc *PullRequestChecker) CheckPRTitleRegexep() *PullRequestChecker {

	if prc.settings.titleRegexep == "" {
		prc.steps = append(prc.steps, Step{status: Skip, message: "No regexep to check", id: "regexep"})
		return prc
	}

	// run regex against pull request title
	regex := regexp.MustCompile(prc.settings.titleRegexep)

	if !regex.MatchString(prc.settings.pullRequestTitle) {
		prc.steps = append(
			prc.steps,
			Step{
				status:  Err,
				message: "PR title does not match specified regular expression",
				id:      "regexep",
			},
		)
		return prc
	}

	prc.steps = append(prc.steps, Step{status: Success, message: "Regular expression check passed", id: "regexep"})
	return prc
}

func (prc *PullRequestChecker) CheckPRLabels() *PullRequestChecker {

	if prc.settings.ignoreOnLabels == "" {
		prc.steps = append(prc.steps, Step{status: Skip, message: "No labels to check", id: "labels"})
		return prc
	}

	labelsToIgnore := strings.Split(prc.settings.ignoreOnLabels, ",")
	labels, err := prc.github.getPRLabels()

	if err != nil {
		step := Step{status: Err, message: err.Error(), id: "labels"}
		fmt.Println("❌", slog.String("step", step.id), slog.String("message", strings.ToLower(step.message)))
		os.Exit(1)
	}

	for _, label := range labelsToIgnore {
		for _, l := range labels {
			if l == label {
				step := Step{status: Success, message: "Skipping checks, ignore label detected", id: "labels"}
				fmt.Println("🦘", slog.String("step", step.id), slog.String("message", strings.ToLower(step.message)))
				os.Exit(0)
			}
		}
	}
	return prc
}

func (prc *PullRequestChecker) CheckPRChecklist() *PullRequestChecker {

	if !prc.settings.checklist {
		prc.steps = append(prc.steps, Step{status: Skip, message: "No checklist to check", id: "checklist"})
		return prc
	}

	readme, err := os.ReadFile("README.md")

	if err != nil {
		prc.steps = append(prc.steps, Step{status: Err, message: err.Error(), id: "checklist"})
		return prc
	}

	re := regexp.MustCompile(fmt.Sprintf(`(?s)%s.*?((?:(?:- \[[ x]\] .+?)(?:\n|$))+)`, prc.settings.checkListTitle))

	// Find the checklist section
	checklistSection := re.FindStringSubmatch(string(readme))

	if len(checklistSection) > 1 {
		// Extract matched items
		checklistItemsRe := regexp.MustCompile(`- \[[ x]\] (.+)`)
		checklistItems := checklistItemsRe.FindAllStringSubmatch(checklistSection[1], -1)
		prc.steps = append(
			prc.steps,
			Step{
				status:  Err,
				message: fmt.Sprintf("Found %d unchecked checklist items", len(checklistItems)),
				id:      "checklist",
			},
		)
		return prc
	}

	prc.steps = append(prc.steps, Step{status: Success, message: "all checklist items are checked", id: "checklist"})
	return prc
}

func (prc *PullRequestChecker) Report() {
	for _, step := range prc.steps {
		switch step.status {
		case Err:
			fmt.Println("❌", slog.String("step", step.id), slog.String("message", strings.ToLower(step.message)))
			prc.errors++
		case Success:
			fmt.Println("✅", slog.String("step", step.id), slog.String("message", strings.ToLower(step.message)))
		case Skip:
			fmt.Println("🦘", slog.String("step", step.id), slog.String("message", strings.ToLower(step.message)))
		}
	}
	if condition := prc.errors > 0; condition {
		log.Fatal(fmt.Sprintf("Found %d errors", prc.errors))
	}
}

func main() {
	prc := PullRequestChecker{
		steps:  []Step{},
		errors: 0,
		github: &Github{
			Client:    &http.Client{},
			Repo:      getEnvVar("REPO", ""),
			PR:        getEnvVar("PR", ""),
			RepoOwner: getEnvVar("REPO_OWNER", ""),
		},
	}

	prc.ReadPRCheckerSettings().
		CheckPRLabels().
		CheckPRTitlePrefixes().
		CheckPRTitleRegexep().
		CheckPRChecklist().Report()
}

func getEnvVar(name, defaultValue string) string {
	if value, ok := os.LookupEnv(name); ok {
		return value
	}
	return defaultValue
}
