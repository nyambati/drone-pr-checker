package plugin

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/google/go-github/v61/github"
	"github.com/nyambati/drone-pr-checker/internal/config"
	g "github.com/nyambati/drone-pr-checker/internal/github"
)

var pullRequestTitle = "feat: add a new feature"

type TestGithubClient struct {
	body   *string
	labels []*github.Label
	err    error
}

func (t *TestGithubClient) GetPullRequest(owner string, repo string, number int) (*github.PullRequest, error) {
	if t.err != nil {
		return nil, t.err
	}
	return &github.PullRequest{
		Body:   t.body,
		Labels: t.labels,
	}, nil
}

func TestPullRequestChecker_CheckPRTitlePrefixes(t *testing.T) {
	type fields struct {
		settings config.Settings
	}
	tests := []struct {
		name   string
		fields fields
		want   func(settings config.Settings) *PullRequestChecker
	}{
		{
			name: "CheckPRTitlePrefixesEmptyString",
			fields: fields{
				settings: config.Settings{
					Prefixes: "",
					Title:    pullRequestTitle,
				},
			},
			want: func(settings config.Settings) *PullRequestChecker {
				return &PullRequestChecker{
					settings: settings,
					steps:    []Step{{status: Skip, message: PrefixSkipMsg, id: PrefixStepID}},
					errors:   0,
				}
			},
		},
		{
			name: "CheckPRTitlePrefixesValidString",
			fields: fields{
				settings: config.Settings{
					Prefixes: "feat:",
					Title:    pullRequestTitle,
				},
			},
			want: func(settings config.Settings) *PullRequestChecker {
				return &PullRequestChecker{
					settings: settings,
					steps:    []Step{{status: Success, message: PrefixSuccesMsg, id: PrefixStepID}},
					errors:   0,
				}
			},
		},
		{
			name: "CheckPRTitlePrefixesInvalidString",
			fields: fields{
				settings: config.Settings{
					Prefixes: "chore:",
					Title:    pullRequestTitle,
				},
			},
			want: func(settings config.Settings) *PullRequestChecker {
				return &PullRequestChecker{
					settings: settings,
					steps: []Step{{
						status:  Err,
						message: fmt.Sprintf(PrefixErrMsg, settings.Prefixes),
						id:      PrefixStepID,
					}},
					errors: 1,
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prc := &PullRequestChecker{
				settings: tt.fields.settings,
			}
			if got := prc.checkPRTitlePrefixes(); !reflect.DeepEqual(got, tt.want(tt.fields.settings)) {
				t.Errorf("PullRequestChecker.CheckPRTitlePrefixes() = %v, want %v", got, tt.want(tt.fields.settings))
			}
		})
	}
}

func TestPullRequestChecker_CheckPRTitleRegexep(t *testing.T) {

	type fields struct {
		settings config.Settings
	}
	tests := []struct {
		name   string
		fields fields
		want   func(settings config.Settings) *PullRequestChecker
	}{
		{
			name: "CheckPRTitleRegexEpEmptyString",
			fields: fields{
				settings: config.Settings{
					Regexp: "",
					Title:  "feat: add a new feature",
				},
			},
			want: func(settings config.Settings) *PullRequestChecker {
				return &PullRequestChecker{
					settings: settings,
					steps:    []Step{{status: Skip, message: RegexpSkipMsg, id: RegexpStepID}},
					errors:   0,
				}
			},
		},
		{
			name: "CheckPRTitleRegexEpValidRegex",
			fields: fields{
				settings: config.Settings{
					Regexp: `^feat:.*$`,
					Title:  "feat: add a new feature",
				},
			},
			want: func(settings config.Settings) *PullRequestChecker {
				return &PullRequestChecker{
					settings: settings,
					steps:    []Step{{status: Success, message: RegexpSuccesMsg, id: RegexpStepID}},
					errors:   0,
				}
			},
		},
		{
			name: "CheckPRTitleRegexEpInvalidRegex",
			fields: fields{
				settings: config.Settings{
					Regexp: `^chore:.*$`,
					Title:  "feat: add a new feature",
				},
			},
			want: func(settings config.Settings) *PullRequestChecker {
				return &PullRequestChecker{
					settings: settings,
					steps:    []Step{{status: Err, message: RegexpErrMsg, id: RegexpStepID}},
					errors:   1,
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prc := &PullRequestChecker{
				settings: tt.fields.settings,
			}
			if got := prc.checkPRTitleRegexep(); !reflect.DeepEqual(got, tt.want(tt.fields.settings)) {
				t.Errorf("PullRequestChecker.CheckPRTitleRegexep() = %v, want %v", got, tt.want(tt.fields.settings))
			}
		})
	}
}

func TestPullRequestChecker_CheckPRLabels(t *testing.T) {
	type fields struct {
		settings config.Settings
		github   g.GitHubInterface
	}
	tests := []struct {
		name   string
		fields fields
		want   func(settings config.Settings, github g.GitHubInterface) *PullRequestChecker
	}{
		{
			name: "CheckPRLabelsEmptyString",
			fields: fields{
				settings: config.Settings{},
				github:   &TestGithubClient{},
			},
			want: func(settings config.Settings, github g.GitHubInterface) *PullRequestChecker {
				return &PullRequestChecker{
					settings: settings,
					github:   github,
					steps:    []Step{{status: Skip, message: LabelsSkipMsg, id: LabelsStepID}},
					errors:   0,
				}
			},
		},
		{
			name: "CheckPRLabelsMatchLabels",
			fields: fields{
				settings: config.Settings{SkipOnLabels: "label1"},
				github: &TestGithubClient{
					labels: []*github.Label{
						{
							Name: github.String("label1"),
						},
					},
				},
			},
			want: func(settings config.Settings, github g.GitHubInterface) *PullRequestChecker {
				return &PullRequestChecker{
					settings: settings,
					github:   github,
					steps:    []Step{{status: Skip, message: LabelsSkipMsg, id: LabelsStepID, exit: true}},
					errors:   0,
				}
			},
		},
		{
			name: "CheckPRLabelsNoMatchLabels",
			fields: fields{
				settings: config.Settings{SkipOnLabels: "label3,label4"},
				github:   &TestGithubClient{},
			},
			want: func(settings config.Settings, github g.GitHubInterface) *PullRequestChecker {
				return &PullRequestChecker{
					settings: settings,
					github:   github,
					steps:    []Step{{status: Success, message: LabelsSuccesMsg, id: LabelsStepID}},
					errors:   0,
				}
			},
		},
		{
			name: "CheckPRLabelsSkipOnGithubError",
			fields: fields{
				settings: config.Settings{SkipOnLabels: "label3,label4"},
				github: &TestGithubClient{
					err: errors.New("Error"),
				},
			},
			want: func(settings config.Settings, github g.GitHubInterface) *PullRequestChecker {
				return &PullRequestChecker{
					settings: settings,
					github:   github,
					steps: []Step{
						{
							status:  Err,
							message: "Error",
							id:      LabelsStepID,
						},
					},
					errors: 1,
				}
			},
		},
		{
			name: "CheckPRLabelsInvalidSkipOnGithubError",
			fields: fields{
				settings: config.Settings{
					SkipOnLabels:      "label3,label4",
					IgnoreGitHubError: true,
				},
				github: &TestGithubClient{
					err: errors.New("Error"),
				},
			},
			want: func(settings config.Settings, github g.GitHubInterface) *PullRequestChecker {
				return &PullRequestChecker{
					settings: settings,
					github:   github,
					steps: []Step{
						{
							status:  Skip,
							message: "Error",
							id:      LabelsStepID,
						},
					},
					errors: 0,
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prc := &PullRequestChecker{
				settings: tt.fields.settings,
				github:   tt.fields.github,
			}
			if got := prc.checkPRLabels(); !reflect.DeepEqual(got, tt.want(tt.fields.settings, tt.fields.github)) {
				t.Errorf("PullRequestChecker.CheckPRLabels() = %v, want %v", got, tt.want(tt.fields.settings, tt.fields.github))
			}
		})
	}
}

func TestPullRequestChecker_CheckPRChecklist(t *testing.T) {

	prBodyUnchecked := []byte(`
## Checklist
- [ ] Completed code review
- [ ] Ran unit tests
- [ ] Completed e2e tests
		`,
	)

	prBodyChecked := []byte(`
## Checklist
- [x] Completed code review
- [x] Ran unit tests
- [x] Completed e2e tests

		`,
	)
	type fields struct {
		settings config.Settings
		github   g.GitHubInterface
	}
	tests := []struct {
		name   string
		fields fields
		want   func(settings config.Settings, github g.GitHubInterface) *PullRequestChecker
	}{
		{
			name: "CheckPRChecklistDisabled",
			fields: fields{
				settings: config.Settings{},
				github: &TestGithubClient{
					body: github.String(string(prBodyUnchecked)),
				},
			},
			want: func(settings config.Settings, github g.GitHubInterface) *PullRequestChecker {
				return &PullRequestChecker{
					settings: settings,
					github:   github,
					steps:    []Step{{status: Skip, message: ChecklistSkipMsg, id: ChecklistStepID}},
					errors:   0,
				}
			},
		},
		{
			name: "CheckPRChecklistUnchecked",
			fields: fields{
				settings: config.Settings{Checklist: true},
				github: &TestGithubClient{
					body: github.String(string(prBodyUnchecked)),
				},
			},
			want: func(settings config.Settings, github g.GitHubInterface) *PullRequestChecker {
				return &PullRequestChecker{
					settings: settings,
					github:   github,
					steps: []Step{
						{
							status:  Err,
							message: fmt.Sprintf(ChecklistErrMsg, 3),
							id:      ChecklistStepID,
						},
					},
					errors: 1,
				}
			},
		},
		{
			name: "CheckPRChecklistChecked",
			fields: fields{
				settings: config.Settings{Checklist: true},
				github: &TestGithubClient{
					body: github.String(string(prBodyChecked)),
				},
			},
			want: func(settings config.Settings, github g.GitHubInterface) *PullRequestChecker {
				return &PullRequestChecker{
					settings: settings,
					github:   github,
					steps: []Step{
						{
							status:  Success,
							message: ChecklistSuccesMsg,
							id:      ChecklistStepID,
						},
					},
					errors: 0,
				}
			},
		},
		{
			name: "CheckPRChecklistInvalidSkipOnGithubError",
			fields: fields{
				settings: config.Settings{
					Checklist:         true,
					IgnoreGitHubError: true,
				},

				github: &TestGithubClient{err: errors.New("Error")},
			},
			want: func(settings config.Settings, github g.GitHubInterface) *PullRequestChecker {
				return &PullRequestChecker{
					settings: settings,
					github:   github,
					steps: []Step{
						{
							status:  Skip,
							message: "Error",
							id:      "checklist",
						},
					},
					errors: 0,
				}
			},
		},
		{
			name: "CheckPRChecklistGitHubError",
			fields: fields{
				settings: config.Settings{
					Checklist: true,
				},
				github: &TestGithubClient{err: errors.New("Error")},
			},
			want: func(settings config.Settings, github g.GitHubInterface) *PullRequestChecker {
				return &PullRequestChecker{
					settings: settings,
					github:   github,
					steps: []Step{{
						status:  Err,
						message: "Error",
						id:      ChecklistStepID,
					}},
					errors: 1,
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prc := &PullRequestChecker{
				settings: tt.fields.settings,
				github:   tt.fields.github,
			}
			if got := prc.checkPRChecklist(); !reflect.DeepEqual(got, tt.want(tt.fields.settings, tt.fields.github)) {
				t.Errorf("PullRequestChecker.CheckPRChecklist() = %v, want %v", got, tt.want(tt.fields.settings, tt.fields.github))
			}
		})
	}
}
