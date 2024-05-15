package plugin

import (
	"fmt"
	"log"
	"log/slog"
	"os"
	"regexp"
	"slices"
	"strings"

	"github.com/nyambati/drone-pr-checker/internal/config"
	"github.com/nyambati/drone-pr-checker/internal/github"
)

type PullRequestChecker struct {
	steps    []Step
	errors   int
	settings config.Settings
	github   github.GitHubInterface
}

func (prc *PullRequestChecker) checkPRTitlePrefixes() *PullRequestChecker {
	if prc.settings.Prefixes == "" {
		prc.steps = append(prc.steps, Step{status: Skip, message: PrefixSkipMsg, id: PrefixStepID})
		return prc
	}

	prefixes := strings.Split(prc.settings.Prefixes, ",")

	for _, prefix := range prefixes {
		if strings.HasPrefix(strings.ToLower(prc.settings.Title), strings.ToLower(prefix)) {
			prc.steps = append(prc.steps, Step{status: Success, message: PrefixSuccesMsg, id: PrefixStepID})
			return prc
		}
	}

	prc.steps = append(
		prc.steps,
		Step{
			status:  Err,
			id:      PrefixStepID,
			message: fmt.Sprintf(PrefixErrMsg, prc.settings.Prefixes),
		},
	)
	prc.errors++
	return prc
}

func (prc *PullRequestChecker) checkPRTitleRegexep() *PullRequestChecker {

	if prc.settings.Regexp == "" {
		prc.steps = append(prc.steps, Step{status: Skip, message: RegexpSkipMsg, id: RegexpStepID})
		return prc
	}

	// run regex against pull request title
	regex := regexp.MustCompile(prc.settings.Regexp)

	if !regex.MatchString(prc.settings.Title) {
		prc.steps = append(prc.steps, Step{status: Err, message: RegexpErrMsg, id: RegexpStepID})
		prc.errors++
		return prc
	}

	prc.steps = append(prc.steps, Step{status: Success, message: RegexpSuccesMsg, id: RegexpStepID})
	return prc
}

func (prc *PullRequestChecker) checkPRLabels() *PullRequestChecker {

	if prc.settings.SkipOnLabels == "" {
		prc.steps = append(prc.steps, Step{status: Skip, message: LabelsSkipMsg, id: LabelsStepID})
		return prc
	}

	labelsToIgnore := strings.Split(prc.settings.SkipOnLabels, ",")

	pr, err := prc.github.GetPullRequest(
		prc.settings.Owner,
		prc.settings.Repo,
		prc.settings.PullRequest,
	)

	if err != nil {
		switch prc.settings.IgnoreGitHubError {
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
		labels = append(labels, label.GetName())
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

func (prc *PullRequestChecker) checkPRChecklist() *PullRequestChecker {

	if !prc.settings.Checklist {
		prc.steps = append(prc.steps, Step{status: Skip, message: ChecklistSkipMsg, id: ChecklistStepID})
		return prc
	}

	pr, err := prc.github.GetPullRequest(
		prc.settings.Owner,
		prc.settings.Repo,
		prc.settings.PullRequest,
	)

	if err != nil {
		switch prc.settings.IgnoreGitHubError {
		case true:
			prc.steps = append(prc.steps, Step{status: Skip, message: err.Error(), id: ChecklistStepID})
			return prc
		default:
			prc.steps = append(prc.steps, Step{status: Err, message: err.Error(), id: ChecklistStepID})
			prc.errors++
			return prc
		}
	}

	re := regexp.MustCompile(
		fmt.Sprintf(
			`(?s)%s.*?((?:(?:- \[[ x]\] .+?)(?:\n|$))+)`,
			prc.settings.ChecklistTitle,
		),
	)

	// Find the checklist section
	checklistSection := re.FindStringSubmatch(pr.GetBody())

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
	checker := prc.checkPRLabels().
		checkPRTitlePrefixes().
		checkPRTitleRegexep().
		checkPRChecklist()

	for _, step := range checker.steps {
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

func New(settings config.Settings, github github.GitHubInterface) PullRequestChecker {
	return PullRequestChecker{
		steps:    []Step{},
		errors:   0,
		github:   github,
		settings: settings,
	}
}
