package shell

import (
	"github.com/AlecAivazis/survey/v2"
	"github.com/AlecAivazis/survey/v2/terminal"
)

// ErrPromptMultiSelectInterrupted error throwed on signal interrupt
var ErrPromptMultiSelectInterrupted error = terminal.InterruptErr

// PromptMultiSelect contract that holds logic for prompt a multi select question
type PromptMultiSelect interface {
	Ask(string, []string) ([]string, error)
}

// DefaultPromptMultiSelect holds data for prompting a multi select question
type DefaultPromptMultiSelect struct{}

// NewPromptMultiSelect creates a new prompt multi select
func NewPromptMultiSelect() PromptMultiSelect {
	return &DefaultPromptMultiSelect{}
}

// Ask prompt to the user a select question
func (p *DefaultPromptMultiSelect) Ask(question string, options []string) (answers []string, err error) {
	prompt := &survey.MultiSelect{
		Message: question,
		Options: options,
	}
	err = survey.AskOne(prompt, &answers)
	return
}
