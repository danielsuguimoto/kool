package shell

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"kool-dev/kool/cmd/builder"
	"kool-dev/kool/environment"
	"os"
	"os/exec"
	"reflect"
	"strings"
	"testing"

	"github.com/gookit/color"
)

func readOutput(r io.Reader) (output string, err error) {
	var out []byte

	if out, err = ioutil.ReadAll(r); err != nil {
		return
	}

	output = strings.TrimSpace(string(out))
	return
}

func newTestingOutputShell() (s Shell, buf *bytes.Buffer) {
	s = NewShell()
	buf = bytes.NewBufferString("")
	s.SetOutStream(buf)

	return
}

func TestNewShell(t *testing.T) {
	s := NewShell()

	if _, ok := s.(*DefaultShell); !ok {
		t.Error("NewShell() did not return a *DefaultShell")
	}

	if s.InStream() != os.Stdin {
		t.Error("NewShell() should initialize input stream with os.Stdin")
	}

	if s.OutStream() != os.Stdout {
		t.Error("NewShell() should initialize output stream with os.Stdout")
	}

	if s.ErrStream() != os.Stderr {
		t.Error("NewShell() should initialize error stream with os.Stderr")
	}
}

func TestSetInStreamDefaultShell(t *testing.T) {
	s := NewShell()

	r := bytes.NewReader([]byte{})
	s.SetInStream(r)

	if s.InStream() != r {
		t.Error("failed calling SetInStream() on *DefaultShell")
	}
}

func TestSetOutStreamDefaultShell(t *testing.T) {
	s := NewShell()

	w := bytes.NewBufferString("")
	s.SetOutStream(w)

	if s.OutStream() != w {
		t.Error("failed calling SetOutStream() on *DefaultShell")
	}
}

func TestSetErrStreamDefaultShell(t *testing.T) {
	s := NewShell()

	w := bytes.NewBufferString("")
	s.SetErrStream(w)

	if s.ErrStream() != w {
		t.Error("failed calling SetErrStream() on *DefaultShell")
	}
}

func TestExecDefaultShell(t *testing.T) {
	s := NewShell()
	command := builder.NewCommand("echo", "x")

	output, err := s.Exec(command)

	if err != nil {
		t.Errorf("unexpected error calling Exec on *DefaultShell, err: %v", err)
	}

	output = strings.TrimSpace(output)

	if output != "x" {
		t.Errorf("Exec failed; expected output 'x', got '%s'", output)
	}
}

func TestExecDockerComposeDefaultShell(t *testing.T) {
	s := &DefaultShell{
		inStream:  os.Stdin,
		outStream: ioutil.Discard,
		errStream: os.Stderr,
		lookedUp:  make(map[string]error),
		env:       environment.NewFakeEnvStorage(),
	}

	s.env.Set("KOOL_NAME", "kool_test")

	command := &builder.FakeCommand{MockCmd: "docker-compose"}
	command.AppendArgs("ps")

	var (
		exeTest  string
		argsTest []string
	)

	originalExecCmdFn := execCmdFn
	execCmdFn = func(exe string, args ...string) *exec.Cmd {
		exeTest = exe
		argsTest = args
		return exec.Command("echo", "x1")
	}
	defer func() {
		execCmdFn = originalExecCmdFn
	}()

	_, _ = s.Exec(command, "extraArg")

	if exeTest != "docker-compose" {
		t.Errorf("expecting to run 'docker-compose', got '%s'", exeTest)
	}

	expectedArgs := []string{"ps", "extraArg"}
	if !reflect.DeepEqual(argsTest, expectedArgs) {
		t.Errorf("expecting args '%v', got '%v'", expectedArgs, argsTest)
	}
}

func TestExec(t *testing.T) {
	output, err := Exec("echo", "x")

	if err != nil {
		t.Errorf("unexpected error calling Exec on *DefaultShell, err: %v", err)
	}

	output = strings.TrimSpace(output)

	if output != "x" {
		t.Errorf("Exec failed; expected output 'x', got '%s'", output)
	}
}

func TestInteractiveDefaultShell(t *testing.T) {
	s := NewShell()
	command := builder.NewCommand("echo", "x")

	r, w, _ := os.Pipe()
	s.SetOutStream(w)

	err := s.Interactive(command)

	w.Close()

	if err != nil {
		t.Errorf("Interactive failed on *DefaultShell; expected no errors 'x', got '%v'", err)
	}

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)

	output := strings.TrimSpace(buf.String())

	if output != "x" {
		t.Errorf("Interactive failed on *DefaultShell; expected output 'x', got '%s'", output)
	}
}

func TestInteractiveDockerComposeDefaultShell(t *testing.T) {
	s := &DefaultShell{
		inStream:  os.Stdin,
		outStream: ioutil.Discard,
		errStream: os.Stderr,
		lookedUp:  make(map[string]error),
		env:       environment.NewFakeEnvStorage(),
	}

	s.env.Set("KOOL_NAME", "kool_test")

	command := &builder.FakeCommand{MockCmd: "docker-compose"}
	command.AppendArgs("ps")

	var (
		exeTest  string
		argsTest []string
	)

	originalExecCmdFn := execCmdFn
	execCmdFn = func(exe string, args ...string) *exec.Cmd {
		exeTest = exe
		argsTest = args
		return exec.Command("echo", "x1")
	}

	originalExecLookPath := execLookPathFn
	execLookPathFn = func(exe string) (string, error) {
		return "", nil
	}
	defer func() {
		execCmdFn = originalExecCmdFn
		execLookPathFn = originalExecLookPath
	}()

	err := s.Interactive(command, "extraArg")

	if err != nil {
		t.Errorf("Interactive failed on *DefaultShell; expected no errors 'x', got '%v'", err)
	}

	if exeTest != "docker-compose" {
		t.Errorf("expecting to run 'docker-compose', got '%s'", exeTest)
	}

	expectedArgs := []string{"ps", "extraArg"}
	if !reflect.DeepEqual(argsTest, expectedArgs) {
		t.Errorf("expecting args '%v', got '%v'", expectedArgs, argsTest)
	}
}

func TestInteractive(t *testing.T) {
	r, w, _ := os.Pipe()

	originalStdout := os.Stdout
	os.Stdout = w
	defer func() {
		os.Stdout = originalStdout
	}()

	err := Interactive("echo", "x")

	w.Close()

	if err != nil {
		t.Errorf("Interactive failed on *DefaultShell; expected no errors 'x', got '%v'", err)
	}

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)

	output := strings.TrimSpace(buf.String())

	if output != "x" {
		t.Errorf("Interactive failed on *DefaultShell; expected output 'x', got '%s'", output)
	}
}

func TestInteractiveLookPathErrorDefaultShell(t *testing.T) {
	s := NewShell()
	s.SetOutStream(ioutil.Discard)

	originalExecLookPath := execLookPathFn
	execLookPathFn = func(exe string) (string, error) {
		return "", errors.New("error")
	}
	defer func() {
		execLookPathFn = originalExecLookPath
	}()

	command := builder.NewCommand("echo", "x")

	err := s.Interactive(command)

	if err == nil {
		t.Errorf("expecting error %v, got none", ErrLookPath)
	} else if err != ErrLookPath {
		t.Errorf("expecting error %v, got %v", ErrLookPath, err)
	}
}

func TestPrintlnShell(t *testing.T) {
	o, b := newTestingOutputShell()

	expected := "testing text"
	o.Println(expected)

	output, err := readOutput(b)

	if err != nil {
		t.Fatal(err)
	}

	if output != expected {
		t.Errorf("expecting output '%s', got '%s'", expected, output)
	}
}

func TestPrintfShell(t *testing.T) {
	o, b := newTestingOutputShell()

	expected := "testing text"
	o.Printf("testing %s", "text")

	output, err := readOutput(b)

	if err != nil {
		t.Fatal(err)
	}

	if output != expected {
		t.Errorf("expecting output '%s', got '%s'", expected, output)
	}
}

func TestErrorShell(t *testing.T) {
	o, b := newTestingOutputShell()

	o.Error(errors.New("testing error"))

	output, err := readOutput(b)

	if err != nil {
		t.Fatal(err)
	}

	expected := color.New(color.BgRed, color.FgWhite).Sprint("error: testing error")

	if output != expected {
		t.Errorf("expecting output '%s', got '%s'", expected, output)
	}
}

func TestWarningShell(t *testing.T) {
	o, b := newTestingOutputShell()

	o.Warning("testing warning")

	output, err := readOutput(b)

	if err != nil {
		t.Fatal(err)
	}

	expected := color.New(color.Yellow).Sprint("testing warning")

	if output != expected {
		t.Errorf("expecting output '%s', got '%s'", expected, output)
	}
}

func TestSuccessShell(t *testing.T) {
	o, b := newTestingOutputShell()

	o.Success("testing success")

	output, err := readOutput(b)

	if err != nil {
		t.Fatal(err)
	}

	expected := color.New(color.Green).Sprint("testing success")

	if output != expected {
		t.Errorf("expecting output '%s', got '%s'", expected, output)
	}
}

func TestRecursiveInteractiveCommand(t *testing.T) {
	s := NewShell()
	command := builder.NewCommand("kool", "-v")

	var (
		calledRecursive     = false
		calledRecursiveArgs []string
	)

	oldRecursiveCall := RecursiveCall
	defer func() {
		RecursiveCall = oldRecursiveCall
	}()

	// set published RecursiveCall handler
	RecursiveCall = func(args []string) error {
		calledRecursive = true
		calledRecursiveArgs = args
		return nil
	}

	err := s.Interactive(command)

	if err != nil {
		t.Errorf("unexpected error calling recursive kool: %s", err.Error())
	}

	if !calledRecursive || len(calledRecursiveArgs) != 1 || calledRecursiveArgs[0] != "-v" {
		t.Errorf("unexpected recursive call - args: %v", calledRecursiveArgs)
	}
}
