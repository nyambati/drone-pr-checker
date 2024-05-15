package config

import (
	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
)

var (
	prefixes          = "plugin_prefixes"
	regexp            = "plugin_regexp"
	skipOnLabels      = "plugin_skip_on_labels"
	ignoreGitHubError = "plugin_ignore_github_error"
	checklist         = "plugin_checklist"
	checklistTitle    = "plugin_checklist_title"
	title             = "drone_pull_request_title"
	githubToken       = "github_token"
	repo              = "drone_repo_name"
	owner             = "drone_repo_owner"
	pullRequest       = "drone_pull_request"
)

var envVars = []string{
	prefixes,
	regexp,
	skipOnLabels,
	ignoreGitHubError,
	checklist,
	checklistTitle,
	title,
	githubToken,
	repo,
	owner,
	pullRequest,
}

func New() (*Config, error) {
	v := viper.New()
	v.SetDefault(checklistTitle, "## Checklist")
	v.SetDefault(ignoreGitHubError, true)
	v.SetDefault(checklist, false)

	for _, envVar := range envVars {
		if err := v.BindEnv(envVar); err != nil {
			return nil, err
		}
	}

	cfg := &Config{
		Settings: Settings{
			Prefixes:          v.GetString(prefixes),
			Regexp:            v.GetString(regexp),
			SkipOnLabels:      v.GetString(skipOnLabels),
			IgnoreGitHubError: v.GetBool(ignoreGitHubError),
			Title:             v.GetString(title),
			ChecklistTitle:    v.GetString(checklistTitle),
			Repo:              v.GetString(repo),
			Owner:             v.GetString(owner),
			PullRequest:       v.GetInt(pullRequest),
			Checklist:         v.GetBool(checklist),
		},
		Github: GitHub{Token: v.GetString(githubToken)},
	}

	return cfg.validate()
}

func (config *Config) validate() (*Config, error) {
	validate := validator.New(validator.WithRequiredStructEnabled())
	if err := validate.Struct(config); err != nil {
		return nil, err
	}
	return config, nil
}
