package internal

const (
	PrefixStepID       = "prefix"
	PrefixSkipMsg      = "No prefixes to check"
	PrefixErrMsg       = "PR title does not have any required prefix (%s)"
	PrefixSuccesMsg    = "Prefixes check passed"
	LabelsStepID       = "labels"
	LabelsSkipMsg      = "No labels to check"
	LabelsErrMsg       = "PR does not have any required labels"
	LabelsSuccesMsg    = "Labels check passed"
	RegexpStepID       = "regexp"
	RegexpSkipMsg      = "No regexep to check"
	RegexpErrMsg       = "PR title does not match specified regular expression"
	RegexpSuccesMsg    = "Regular expression check passed"
	ChecklistStepID    = "checklist"
	ChecklistSkipMsg   = "Checklist checks disabled"
	ChecklistErrMsg    = "Found %d unchecked checklist items"
	ChecklistSuccesMsg = "Checklist check passed"
)
