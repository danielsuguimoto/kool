package shell

// FakePromptMultiSelect holds data for fake prompt multi select behavior
type FakePromptMultiSelect struct {
	CalledAsk  bool
	MockAnswers map[string][]string
	MockError  map[string]error
}

// Ask fake behavior for prompting a select question
func (f *FakePromptMultiSelect) Ask(question string, options []string) (answers []string, err error) {
	f.CalledAsk = true
	answers = f.MockAnswers[question]
	err = f.MockError[question]
	return
}
