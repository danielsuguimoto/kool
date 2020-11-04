package shell

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

func TestNewPromptMultiSelect(t *testing.T) {
	p := NewPromptMultiSelect()

	if _, ok := p.(*DefaultPromptMultiSelect); !ok {
		t.Errorf("unexpected PromptMultiSelect on NewPromptMultiSelect")
	}
}

func TestAskPromptMultiSelect(t *testing.T) {
	oldStdout := os.Stdout
	oldStdin := os.Stdin

	r, w, _ := os.Pipe()
	rStdin, wStdin, _ := os.Pipe()

	os.Stdin = rStdin
	os.Stdout = w

	wStdin.WriteString("\n")

	p := NewPromptMultiSelect()

	_, _ = p.Ask("testing_question", []string{"testing_option1", "testing_option2"})

	w.Close()
	wStdin.Close()

	out, err := ioutil.ReadAll(r)
	os.Stdout = oldStdout
	os.Stdin = oldStdin

	if err != nil {
		t.Fatal(err)
	}

	output := string(out)

	if !strings.Contains(output, "testing_question") || !strings.Contains(output, "testing_option1") || !strings.Contains(output, "testing_option2") {
		t.Error("failed to render the question and its options")
	}
}
