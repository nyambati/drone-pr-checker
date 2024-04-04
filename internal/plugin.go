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
	prefixes          string
	regexp            string
	skipOnLabels      string
	ignoreGitHubError bool
	checklist         bool
	title             string
	checklistTitle    string
}

type PullRequestChecker struct {
	steps    []Step
	errors   int
	settings Settings
	github   GitHub
}

type Plugin interface {
	Report()
	CheckPrTitlePrefixes() *PullRequestChecker
	CheckPRChecklist() *PullRequestChecker
	CheckPRLabels() *PullRequestChecker
	CheckPRTitleRegexep() *PullRequestChecker
}

func (prc *PullRequestChecker) CheckPRTitlePrefixes() *PullRequestChecker {
	if prc.settings.prefixes == "" {
		prc.steps = append(prc.steps, Step{status: Skip, message: PrefixSkipMsg, id: PrefixStepID})
		return prc
	}

	prefixes := strings.Split(prc.settings.prefixes, ",")

	for _, prefix := range prefixes {
		if strings.HasPrefix(strings.ToLower(prc.settings.title), strings.ToLower(prefix)) {
			prc.steps = append(prc.steps, Step{status: Success, message: PrefixSuccesMsg, id: PrefixStepID})
			return prc
		}
	}

	prc.steps = append(
		prc.steps,
		Step{
			status:  Err,
			id:      PrefixStepID,
			message: fmt.Sprintf(PrefixErrMsg, prc.settings.prefixes),
		},
	)
	prc.errors++
	return prc
}

func (prc *PullRequestChecker) CheckPRTitleRegexep() *PullRequestChecker {

	if prc.settings.regexp == "" {
		prc.steps = append(prc.steps, Step{status: Skip, message: RegexpSkipMsg, id: RegexpStepID})
		return prc
	}

	// run regex against pull request title
	regex := regexp.MustCompile(prc.settings.regexp)

	if !regex.MatchString(prc.settings.title) {
		prc.steps = append(prc.steps, Step{status: Err, message: RegexpErrMsg, id: RegexpStepID})
		prc.errors++
		return prc
	}

	prc.steps = append(prc.steps, Step{status: Success, message: RegexpSuccesMsg, id: RegexpStepID})
	return prc
}

func (prc *PullRequestChecker) CheckPRLabels() *PullRequestChecker {

	if prc.settings.skipOnLabels == "" {
		prc.steps = append(prc.steps, Step{status: Skip, message: LabelsSkipMsg, id: LabelsStepID})
		return prc
	}

	labelsToIgnore := strings.Split(prc.settings.skipOnLabels, ",")

	pr, err := prc.github.GetPullRequest()

	if err != nil {
		switch prc.settings.ignoreGitHubError {
		case true:
			prc.steps = append(prc.steps, Step{status: Skip, message: err.Error(), id: LabelsStepID})
			return prc
		default:
			prc.steps = append(prc.steps, Step{status: Err, message: err.Error(), id: LabelsStepID})
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
					message: LabelsSkipMsg,
					id:      LabelsStepID,
					exit:    true,
				},
			)
			return prc
		}
	}

	prc.steps = append(prc.steps, Step{status: Success, message: LabelsSuccesMsg, id: LabelsStepID})
	return prc
}

func (prc *PullRequestChecker) CheckPRChecklist() *PullRequestChecker {

	if !prc.settings.checklist {
		prc.steps = append(prc.steps, Step{status: Skip, message: ChecklistSkipMsg, id: ChecklistStepID})
		return prc
	}

	pr, err := prc.github.GetPullRequest()

	if err != nil {
		switch prc.settings.ignoreGitHubError {
		case true:
			prc.steps = append(prc.steps, Step{status: Skip, message: err.Error(), id: ChecklistStepID})
			return prc
		default:
			prc.steps = append(prc.steps, Step{status: Err, message: err.Error(), id: ChecklistStepID})
			prc.errors++
			return prc
		}
	}

	re := regexp.MustCompile(fmt.Sprintf(`(?s)%s.*?((?:(?:- \[[ x]\] .+?)(?:\n|$))+)`, prc.settings.checklistTitle))

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
					message: fmt.Sprintf(ChecklistErrMsg, len(checklistItems)),
					id:      ChecklistStepID,
				},
			)
			prc.errors++
			return prc
		}

	}

	prc.steps = append(prc.steps, Step{status: Success, message: ChecklistSuccesMsg, id: ChecklistStepID})
	return prc
}

func (prc *PullRequestChecker) Report() {
	for _, step := range prc.steps {
		switch step.status {
		case Err:
			fmt.Println("❌", slog.String("step", step.id), slog.String("message", strings.ToLower(step.message)))
		case Success:
			fmt.Println("✅", slog.String("step", step.id), slog.String("message", strings.ToLower(step.message)))
		case Skip:
			fmt.Println("🦘", slog.String("step", step.id), slog.String("message", strings.ToLower(step.message)))
			// Exit gracefully when exit is detected. Comes from the labels check.
			if step.exit {
				os.Exit(0)
			}
		}
	}

	if condition := prc.errors > 0; condition {
		log.Fatalf("Found %d errors", prc.errors)
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
	ignoreGitHubError := GetEnvVar("PLUGIN_IGNORE_ON_GITHUB_ERRORS", "false") == "true"
	checklist := GetEnvVar("PLUGIN_CHECKLIST", "false") == "true"
	regex := GetEnvVar("PLUGIN_TITLE_REGEXP", "")
	pullRequestTitle := GetEnvVar("DRONE_PULL_REQUEST_TITLE", "")
	checkListTitle := GetEnvVar("PLUGIN_CHECKLIST_TITLE", "## Checklist")

	return Settings{
		prefixes:          prefixes,
		regexp:            regex,
		skipOnLabels:      labels,
		checklist:         checklist,
		title:             pullRequestTitle,
		checklistTitle:    checkListTitle,
		ignoreGitHubError: ignoreGitHubError,
	}
}

func NewPlugin(url *url.URL, token string) PullRequestChecker {
	return PullRequestChecker{
		steps:    []Step{},
		errors:   0,
		github:   NewGithub(url, token),
		settings: readPRCheckerSettings(),
	}
}
