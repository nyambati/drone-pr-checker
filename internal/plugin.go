package internal

import (
	"fmt"
	"log"
	"log/slog"
	"net/url"
	"os"
	"regexp"
	"slices"
	"strings"
)

type State int

const (
	Success State = iota
	Err
	Skip
)

var BaseURL = "https://api.github.com"

type Step struct {
	status  State
	message string
	id      string
	exit    bool
}

type Settings struct {
	titlePrefixes       string
	titleRegexep        string
	skipOnLabels        string
	ignoreOnGitHubError bool
	checklist           bool
	pullRequestTitle    string
	checkListTitle      string
}

type PullRequestChecker struct {
	steps    []Step
	errors   int
	settings Settings
	github   GitHub
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
	prc.errors++
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
		prc.errors++
		return prc
	}

	prc.steps = append(prc.steps, Step{status: Success, message: "Regular expression check passed", id: "regexep"})
	return prc
}

func (prc *PullRequestChecker) CheckPRLabels() *PullRequestChecker {

	if prc.settings.skipOnLabels == "" {
		prc.steps = append(prc.steps, Step{status: Skip, message: "No labels to check", id: "labels"})
		return prc
	}

	labelsToIgnore := strings.Split(prc.settings.skipOnLabels, ",")

	pr, err := prc.github.GetPullRequest()

	if err != nil {
		switch prc.settings.ignoreOnGitHubError {
		case true:
			prc.steps = append(prc.steps, Step{status: Skip, message: err.Error(), id: "labels"})
			return prc
		default:
			prc.steps = append(prc.steps, Step{status: Err, message: err.Error(), id: "labels"})
			prc.errors++
			return prc
		}
	}

	labels := []string{}

	for _, label := range pr.Labels {
		labels = append(labels, label.Name)
	}

	for _, label := range labelsToIgnore {
		if slices.Contains(labels, label) {
			prc.steps = append(
				prc.steps,
				Step{
					status:  Skip,
					message: "Skipping, detected skip label",
					id:      "labels",
					exit:    true,
				},
			)
			return prc
		}
	}

	prc.steps = append(prc.steps, Step{status: Success, message: "No skip labels detected", id: "labels"})
	return prc
}

func (prc *PullRequestChecker) CheckPRChecklist() *PullRequestChecker {

	if !prc.settings.checklist {
		prc.steps = append(prc.steps, Step{status: Skip, message: "Checklist checks disabled", id: "checklist"})
		return prc
	}

	pr, err := prc.github.GetPullRequest()

	if err != nil {
		switch prc.settings.ignoreOnGitHubError {
		case true:
			prc.steps = append(prc.steps, Step{status: Skip, message: err.Error(), id: "checklist"})
			return prc
		default:
			prc.steps = append(prc.steps, Step{status: Err, message: err.Error(), id: "checklist"})
			prc.errors++
			return prc
		}
	}

	re := regexp.MustCompile(fmt.Sprintf(`(?s)%s.*?((?:(?:- \[[ x]\] .+?)(?:\n|$))+)`, prc.settings.checkListTitle))

	// Find the checklist section
	checklistSection := re.FindStringSubmatch(pr.Body)

	if len(checklistSection) > 1 {
		// Extract matched items
		checklistItemsRe := regexp.MustCompile(`- \[[ ]\] (.+)`)
		checklistItems := checklistItemsRe.FindAllStringSubmatch(checklistSection[1], -1)
		if len(checklistItems) > 1 {
			prc.steps = append(
				prc.steps,
				Step{
					status:  Err,
					message: fmt.Sprintf("Found %d unchecked checklist items", len(checklistItems)),
					id:      "checklist",
				},
			)
			prc.errors++
			return prc
		}

	}

	prc.steps = append(prc.steps, Step{status: Success, message: "Found 0 unchecked checklist items", id: "checklist"})
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

func GetEnvVar(name, defaultValue string) string {
	if value, ok := os.LookupEnv(name); ok {
		return value
	}
	return defaultValue
}

func readPRCheckerSettings() Settings {
	prefixes := GetEnvVar("PLUGIN_PREFIXES", "")
	labels := GetEnvVar("PLUGIN_SKIP_ON_LABELS", "")
	ignoreOnGitHubError := GetEnvVar("PLUGIN_IGNORE_ON_GITHUB_ERRORS", "false") == "true"
	checklist := GetEnvVar("PLUGIN_CHECKLIST", "false") == "true"
	regex := GetEnvVar("PLUGIN_TITLE_REGEXEP", "")
	pullRequestTitle := GetEnvVar("DRONE_PULL_REQUEST_TITLE", "")
	checkListTitle := GetEnvVar("PLUGIN_CHECKLIST_TITLE", "## Checklist")

	return Settings{
		titlePrefixes:       prefixes,
		titleRegexep:        regex,
		skipOnLabels:        labels,
		checklist:           checklist,
		pullRequestTitle:    pullRequestTitle,
		checkListTitle:      checkListTitle,
		ignoreOnGitHubError: ignoreOnGitHubError,
	}
}

func NewPullRequestChecker(url *url.URL, token string) PullRequestChecker {
	return PullRequestChecker{
		steps:    []Step{},
		errors:   0,
		github:   NewGithub(url, token),
		settings: readPRCheckerSettings(),
	}
}
