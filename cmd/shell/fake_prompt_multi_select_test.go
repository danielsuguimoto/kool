package shell

import (
	"errors"
	"testing"
)

func TestFakePromptMultiSelect(t *testing.T) {
	f := &FakePromptMultiSelect{}
	f.MockAnswers = make(map[string][]string)
	f.MockAnswers["question"] = []string{"answer1", "answer2"}

	answers, err := f.Ask("question", []string{"option1", "option2", "option3"})

	if err != nil {
		t.Errorf("unexpected error on Ask: %v", err)
	}

	if len(answers) != 2 || answers[0] != "answer1" || answers[1] != "answer2"{
		t.Errorf("expecting answers '[answer1 answer2]', got %v", answers)
	}

	f.MockError = make(map[string]error)
	f.MockError["question"] = errors.New("error")

	_, err = f.Ask("question", []string{"option"})

	if err == nil {
		t.Errorf("should throw an error on Ask")
	}
}
