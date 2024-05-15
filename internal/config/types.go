package config

type Config struct {
	Settings Settings
	Github   GitHub
}

type GitHub struct {
	Token string `validate:"required"`
}

type Settings struct {
	Prefixes          string `validate:"required"`
	Regexp            string `validate:"required"`
	SkipOnLabels      string `validate:"required"`
	IgnoreGitHubError bool   `validate:"required"`
	Checklist         bool   `validate:"required"`
	Title             string `validate:"required"`
	ChecklistTitle    string `validate:"required"`
	Repo              string `validate:"required"`
	Owner             string `validate:"required"`
	PullRequest       int    `validate:"required"`
}
