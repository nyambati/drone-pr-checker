package internal

import (
	"fmt"
	"net/url"
	"reflect"
	"testing"

	"github.com/h2non/gock"
)

var pullRequestTitle = "feat: add a new feature"

func TestPullRequestChecker_CheckPRTitlePrefixes(t *testing.T) {
	type fields struct {
		settings Settings
	}
	tests := []struct {
		name   string
		fields fields
		want   func(settings Settings) *PullRequestChecker
	}{
		{
			name: "CheckPRTitlePrefixesEmptyString",
			fields: fields{
				settings: Settings{
					prefixes: "",
					title:    pullRequestTitle,
				},
			},
			want: func(settings Settings) *PullRequestChecker {
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
				settings: Settings{
					prefixes: "feat:",
					title:    pullRequestTitle,
				},
			},
			want: func(settings Settings) *PullRequestChecker {
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
				settings: Settings{
					prefixes: "chore:",
					title:    pullRequestTitle,
				},
			},
			want: func(settings Settings) *PullRequestChecker {
				return &PullRequestChecker{
					settings: settings,
					steps: []Step{{
						status:  Err,
						message: fmt.Sprintf(PrefixErrMsg, settings.prefixes),
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
			if got := prc.CheckPRTitlePrefixes(); !reflect.DeepEqual(got, tt.want(tt.fields.settings)) {
				t.Errorf("PullRequestChecker.CheckPRTitlePrefixes() = %v, want %v", got, tt.want(tt.fields.settings))
			}
		})
	}
}

func TestPullRequestChecker_CheckPRTitleRegexep(t *testing.T) {

	type fields struct {
		settings Settings
	}
	tests := []struct {
		name   string
		fields fields
		want   func(settings Settings) *PullRequestChecker
	}{
		{
			name: "CheckPRTitleRegexEpEmptyString",
			fields: fields{
				settings: Settings{
					regexp: "",
					title:  "feat: add a new feature",
				},
			},
			want: func(settings Settings) *PullRequestChecker {
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
				settings: Settings{
					regexp: `^feat:.*$`,
					title:  "feat: add a new feature",
				},
			},
			want: func(settings Settings) *PullRequestChecker {
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
				settings: Settings{
					regexp: `^chore:.*$`,
					title:  "feat: add a new feature",
				},
			},
			want: func(settings Settings) *PullRequestChecker {
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
			if got := prc.CheckPRTitleRegexep(); !reflect.DeepEqual(got, tt.want(tt.fields.settings)) {
				t.Errorf("PullRequestChecker.CheckPRTitleRegexep() = %v, want %v", got, tt.want(tt.fields.settings))
			}
		})
	}
}

func TestPullRequestChecker_CheckPRLabels(t *testing.T) {
	gurl := &url.URL{
		Scheme: "http",
		Host:   "localhost",
		Path:   "/repos/sample/pulls/1",
	}

	errUrl := &url.URL{
		Scheme: "http",
		Host:   "localhost",
		Path:   "/repos/sample/pulls/2",
	}

	type fields struct {
		settings Settings
		github   GitHub
	}
	tests := []struct {
		name   string
		fields fields
		want   func(settings Settings, github GitHub) *PullRequestChecker
	}{
		{
			name: "CheckPRLabelsEmptyString",
			fields: fields{
				settings: Settings{},
				github:   NewGithub(gurl, "token"),
			},
			want: func(settings Settings, github GitHub) *PullRequestChecker {
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
				settings: Settings{
					skipOnLabels: "label1",
				},
				github: NewGithub(gurl, "token"),
			},
			want: func(settings Settings, github GitHub) *PullRequestChecker {
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
				settings: Settings{
					skipOnLabels: "label3,label4",
				},
				github: NewGithub(gurl, "token"),
			},
			want: func(settings Settings, github GitHub) *PullRequestChecker {
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
				settings: Settings{
					skipOnLabels: "label3,label4",
				},
				github: NewGithub(errUrl, "token"),
			},
			want: func(settings Settings, github GitHub) *PullRequestChecker {
				return &PullRequestChecker{
					settings: settings,
					github:   github,
					steps: []Step{
						{
							status:  Err,
							message: "Get \"http://localhost/repos/sample/pulls/2\": gock: cannot match any request",
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
				settings: Settings{
					skipOnLabels:      "label3,label4",
					ignoreGitHubError: true,
				},
				github: NewGithub(errUrl, "token"),
			},
			want: func(settings Settings, github GitHub) *PullRequestChecker {
				return &PullRequestChecker{
					settings: settings,
					github:   github,
					steps: []Step{
						{
							status:  Skip,
							message: "Get \"http://localhost/repos/sample/pulls/2\": gock: cannot match any request",
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
			defer gock.Off()

			gock.New("http://localhost").
				MatchHeader("Authorization", "Bearer token").
				MatchHeader("Accept", "application/vnd.github+json").
				MatchHeader("X-GitHub-Api-Version", "2022-11-28").
				Get("/repos/sample/pulls/1").
				Reply(200).
				JSON(map[string]interface{}{"labels": []Label{{Name: "label1"}, {Name: "label2"}}})

			prc := &PullRequestChecker{
				settings: tt.fields.settings,
				github:   tt.fields.github,
			}
			if got := prc.CheckPRLabels(); !reflect.DeepEqual(got, tt.want(tt.fields.settings, tt.fields.github)) {
				t.Errorf("PullRequestChecker.CheckPRLabels() = %v, want %v", got, tt.want(tt.fields.settings, tt.fields.github))
			}
		})
	}
}

func TestPullRequestChecker_CheckPRChecklist(t *testing.T) {

	gurl := &url.URL{
		Scheme: "http",
		Host:   "localhost",
		Path:   "/repos/sample/pulls/1",
	}

	errUrl := &url.URL{
		Scheme: "http",
		Host:   "localhost",
		Path:   "/repos/sample/pulls/2",
	}

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
		settings Settings
		github   GitHub
		prBody   []byte
	}
	tests := []struct {
		name   string
		fields fields
		want   func(settings Settings, github GitHub) *PullRequestChecker
	}{
		{
			name: "CheckPRChecklistDisabled",
			fields: fields{
				settings: Settings{},
				github:   NewGithub(gurl, "token"),
				prBody:   prBodyUnchecked,
			},
			want: func(settings Settings, github GitHub) *PullRequestChecker {
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
				github: NewGithub(gurl, "token"),
				prBody: prBodyUnchecked,
				settings: Settings{
					checklist: true,
				},
			},
			want: func(settings Settings, github GitHub) *PullRequestChecker {
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
				github: NewGithub(gurl, "token"),
				prBody: prBodyChecked,
				settings: Settings{
					checklist: true,
				},
			},
			want: func(settings Settings, github GitHub) *PullRequestChecker {
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
				settings: Settings{
					checklist:         true,
					ignoreGitHubError: true,
				},
				github: NewGithub(errUrl, "token"),
				prBody: prBodyUnchecked,
			},
			want: func(settings Settings, github GitHub) *PullRequestChecker {
				return &PullRequestChecker{
					settings: settings,
					github:   github,
					steps: []Step{
						{
							status:  Skip,
							message: "Get \"http://localhost/repos/sample/pulls/2\": gock: cannot match any request",
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
				settings: Settings{
					checklist: true,
				},
				github: NewGithub(errUrl, "token"),
				prBody: prBodyUnchecked,
			},
			want: func(settings Settings, github GitHub) *PullRequestChecker {
				return &PullRequestChecker{
					settings: settings,
					github:   github,
					steps: []Step{{
						status:  Err,
						message: "Get \"http://localhost/repos/sample/pulls/2\": gock: cannot match any request",
						id:      ChecklistStepID,
					}},
					errors: 1,
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer gock.Off()

			data := PullRequest{
				Labels: []Label{{Name: "label 1"}, {Name: "label 2"}},
				Body:   string(tt.fields.prBody),
			}

			gock.New("http://localhost").
				MatchHeader("Authorization", "Bearer token").
				MatchHeader("Accept", "application/vnd.github+json").
				MatchHeader("X-GitHub-Api-Version", "2022-11-28").
				Get("/repos/sample/pulls/1").
				Reply(200).
				JSON(data)

			prc := &PullRequestChecker{
				settings: tt.fields.settings,
				github:   tt.fields.github,
			}

			if got := prc.CheckPRChecklist(); !reflect.DeepEqual(got, tt.want(tt.fields.settings, tt.fields.github)) {
				t.Errorf("PullRequestChecker.CheckPRChecklist() = %v, want %v", got, tt.want(tt.fields.settings, tt.fields.github))
			}
		})
	}
}
