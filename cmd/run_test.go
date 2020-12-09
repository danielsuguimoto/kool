package cmd

import (
	"errors"
	"fmt"
	"kool-dev/kool/cmd/builder"
	"kool-dev/kool/cmd/parser"
	"kool-dev/kool/cmd/shell"
	"kool-dev/kool/environment"
	"strings"
	"testing"
)

func newFakeKoolRun(mockParsedCommands []builder.Command, mockParseError error) *KoolRun {
	return &KoolRun{
		*newFakeKoolService(),
		&parser.FakeParser{MockParsedCommands: mockParsedCommands, MockParseError: mockParseError},
		environment.NewFakeEnvStorage(),
		[]builder.Command{},
		&shell.FakePromptInput{},
	}
}

func TestNewKoolRun(t *testing.T) {
	k := NewKoolRun()

	if _, ok := k.DefaultKoolService.shell.(*shell.DefaultShell); !ok {
		t.Errorf("unexpected shell.Shell on default KoolRun instance")
	}

	if _, ok := k.DefaultKoolService.exiter.(*shell.DefaultExiter); !ok {
		t.Errorf("unexpected shell.Exiter on default KoolRun instance")
	}

	if _, ok := k.DefaultKoolService.term.(*shell.DefaultTerminalChecker); !ok {
		t.Errorf("unexpected shell.TerminalChecker on default KoolRun instance")
	}

	if _, ok := k.parser.(*parser.DefaultParser); !ok {
		t.Errorf("unexpected parser.Parser on default KoolRun instance")
	}
}

func TestNewRunCommand(t *testing.T) {
	fakeParsedCommands := []builder.Command{&builder.FakeCommand{MockCmd: "cmd1"}}

	f := newFakeKoolRun(fakeParsedCommands, nil)
	cmd := NewRunCommand(f)

	cmd.SetArgs([]string{"script"})

	if err := cmd.Execute(); err != nil {
		t.Errorf("unexpected error executing run command; error: %v", err)
	}

	if !f.parser.(*parser.FakeParser).CalledAddLookupPath {
		t.Errorf("did not call AddLookupPath")
	}

	targetFiles := f.parser.(*parser.FakeParser).TargetFiles

	if len(targetFiles) != 2 {
		t.Errorf("did not call AddLookupPath twice (global and local)")
	}

	if !f.parser.(*parser.FakeParser).CalledParse {
		t.Errorf("did not call Parse")
	}

	if len(f.commands) != 1 {
		t.Errorf("did not parse the commands")
	}

	for _, command := range f.commands {
		if command.(*builder.FakeCommand).CalledAppendArgs {
			t.Errorf("unexpected AppendArgs call by parsed command")
		}

		if val, ok := f.shell.(*shell.FakeShell).CalledInteractive[command.Cmd()]; !ok || !val {
			t.Errorf("parsed command did not call Interactive")
		}
	}
}

func TestNewRunCommandMultipleScriptsWarning(t *testing.T) {
	f := newFakeKoolRun([]builder.Command{}, parser.ErrMultipleDefinedScript)
	cmd := NewRunCommand(f)

	cmd.SetArgs([]string{"script"})

	if err := cmd.Execute(); err != nil {
		t.Errorf("unexpected error executing run command; error: %v", err)
	}

	if !f.shell.(*shell.FakeShell).CalledWarning {
		t.Errorf("did not call Warning for multiple scripts")
	}

	expectedWarning := "Attention: the script was found in more than one kool.yml file"

	if gotWarning := fmt.Sprint(f.shell.(*shell.FakeShell).WarningOutput...); gotWarning != expectedWarning {
		t.Errorf("expecting warning '%s', got '%s'", expectedWarning, gotWarning)
	}
}

func TestNewRunCommandParseError(t *testing.T) {
	f := newFakeKoolRun([]builder.Command{}, errors.New("parse error"))
	cmd := NewRunCommand(f)

	cmd.SetArgs([]string{"script"})

	if err := cmd.Execute(); err != nil {
		t.Errorf("unexpected error executing run command; error: %v", err)
	}

	if !f.shell.(*shell.FakeShell).CalledError {
		t.Error("did not call Error for parse error")
	}

	expectedError := "parse error"

	if gotError := f.shell.(*shell.FakeShell).Err.Error(); gotError != expectedError {
		t.Errorf("expecting error '%s', got '%s'", expectedError, gotError)
	}

	if !f.exiter.(*shell.FakeExiter).Exited() {
		t.Error("got an parse error, but command did not exit")
	}
}

func TestNewRunCommandExtraArgsError(t *testing.T) {
	fakeParsedCommands := []builder.Command{&builder.FakeCommand{}, &builder.FakeCommand{}}
	f := newFakeKoolRun(fakeParsedCommands, nil)
	cmd := NewRunCommand(f)

	cmd.SetArgs([]string{"script", "extraArg"})

	if err := cmd.Execute(); err != nil {
		t.Errorf("unexpected error executing run command; error: %v", err)
	}

	if !f.shell.(*shell.FakeShell).CalledError {
		t.Error("did not call Error for extra arguments")
	}

	expectedError := ErrExtraArguments.Error()

	if gotError := f.shell.(*shell.FakeShell).Err.Error(); gotError != expectedError {
		t.Errorf("expecting error '%s', got '%s'", expectedError, gotError)
	}

	if !f.exiter.(*shell.FakeExiter).Exited() {
		t.Error("got an extra arguments error, but command did not exit")
	}
}

func TestNewRunCommandErrorInteractive(t *testing.T) {
	f := newFakeKoolRun([]builder.Command{&builder.FakeCommand{MockInteractiveError: errors.New("interactive error")}}, nil)
	cmd := NewRunCommand(f)

	cmd.SetArgs([]string{"script"})

	if err := cmd.Execute(); err != nil {
		t.Errorf("unexpected error executing run command; error: %v", err)
	}

	if !f.shell.(*shell.FakeShell).CalledError {
		t.Error("did not call Error for parsed command failure")
	}

	expectedError := "interactive error"

	if gotError := f.shell.(*shell.FakeShell).Err.Error(); gotError != expectedError {
		t.Errorf("expecting error '%s', got '%s'", expectedError, gotError)
	}

	if !f.exiter.(*shell.FakeExiter).Exited() {
		t.Error("got an error executing parsed command, but command did not exit")
	}
}

func TestNewRunCommandScriptNotFound(t *testing.T) {
	f := newFakeKoolRun([]builder.Command{}, nil)
	cmd := NewRunCommand(f)

	cmd.SetArgs([]string{"script"})

	if err := cmd.Execute(); err != nil {
		t.Errorf("unexpected error executing run command; error: %v", err)
	}

	if !f.shell.(*shell.FakeShell).CalledError {
		t.Error("did not call Error for not found script error")
	}

	expectedError := ErrKoolScriptNotFound.Error()

	if gotError := f.shell.(*shell.FakeShell).Err.Error(); gotError != expectedError {
		t.Errorf("expecting error '%s', got '%s'", expectedError, gotError)
	}

	if !f.exiter.(*shell.FakeExiter).Exited() {
		t.Error("got an not found script error, but command did not exit")
	}
}

func TestNewRunCommandWithArguments(t *testing.T) {
	fakeParsedCommands := []builder.Command{&builder.FakeCommand{}}
	f := newFakeKoolRun(fakeParsedCommands, nil)
	cmd := NewRunCommand(f)

	cmd.SetArgs([]string{"script", "arg1", "arg2"})

	if err := cmd.Execute(); err != nil {
		t.Errorf("unexpected error executing run command; error: %v", err)
	}

	if !f.commands[0].(*builder.FakeCommand).CalledAppendArgs {
		t.Error("did not call AppendArgs for parsed command")
	}

	fakeCommandArgs := f.commands[0].(*builder.FakeCommand).ArgsAppend

	if len(fakeCommandArgs) != 2 || fakeCommandArgs[0] != "arg1" || fakeCommandArgs[1] != "arg2" {
		t.Error("did not call AppendArgs properly for parsed command")
	}
}

func TestNewRunCommandUsageTemplate(t *testing.T) {
	f := newFakeKoolRun([]builder.Command{}, nil)
	f.parser.(*parser.FakeParser).MockScripts = []string{"testing_script"}
	cmd := NewRunCommand(f)
	SetRunUsageFunc(f, cmd)

	cmd.SetArgs([]string{"--help"})

	if err := cmd.Execute(); err != nil {
		t.Errorf("unexpected error executing run command; error: %v", err)
	}

	if !f.shell.(*shell.FakeShell).CalledPrintln {
		t.Error("did not call Println for command usage")
	}

	usage := f.shell.(*shell.FakeShell).OutLines[0]

	if !strings.Contains(usage, "testing_script") {
		t.Error("did not find testing_script as available script on usage text")
	}
}

func TestNewRunCommandFailingUsageTemplate(t *testing.T) {
	f := newFakeKoolRun([]builder.Command{}, nil)
	f.parser.(*parser.FakeParser).MockScripts = []string{"testing_script"}
	f.parser.(*parser.FakeParser).MockParseAvailableScriptsError = errors.New("error parse avaliable scripts")
	f.envStorage.(*environment.FakeEnvStorage).Envs["KOOL_VERBOSE"] = "1"

	cmd := NewRunCommand(f)
	SetRunUsageFunc(f, cmd)

	cmd.SetArgs([]string{"--help"})

	if err := cmd.Execute(); err != nil {
		t.Errorf("unexpected error executing run command; error: %v", err)
	}

	output := f.shell.(*shell.FakeShell).OutLines[0]

	if strings.Contains(output, "testing_script") {
		t.Error("should not find testing_script as available script on usage text due to error on parsing scripts")
	}

	if !f.shell.(*shell.FakeShell).CalledPrintln {
		t.Error("did not call Println to output error on getting available scripts when KOOL_VERBOSE is true")
	}

	expected := "$ got an error trying to add available scripts to command usage template; error: error parse avaliable scripts"

	if expected != output {
		t.Errorf("expecting message '%s', got '%s'", expected, output)
	}
}

func TestNewRunCommandCompletion(t *testing.T) {
	var scripts []string
	f := newFakeKoolRun([]builder.Command{}, nil)
	f.parser.(*parser.FakeParser).MockScripts = []string{"testing_script"}
	cmd := NewRunCommand(f)

	scripts, _ = cmd.ValidArgsFunction(cmd, []string{}, "")

	if len(scripts) != 1 || scripts[0] != "testing_script" {
		t.Errorf("expecting suggestions [testing_script], got %v", scripts)
	}

	scripts, _ = cmd.ValidArgsFunction(cmd, []string{}, "tes")

	if len(scripts) != 1 || scripts[0] != "testing_script" {
		t.Errorf("expecting suggestions [testing_script], got %v", scripts)
	}

	scripts, _ = cmd.ValidArgsFunction(cmd, []string{}, "invalid")

	if len(scripts) != 0 {
		t.Errorf("expecting no suggestion, got %v", scripts)
	}

	scripts, _ = cmd.ValidArgsFunction(cmd, []string{"testing_script"}, "")

	if scripts != nil {
		t.Errorf("expecting no suggestion, got %v", scripts)
	}
}

func TestNewRunCommandFailingCompletion(t *testing.T) {
	var scripts []string
	f := newFakeKoolRun([]builder.Command{}, nil)
	f.parser.(*parser.FakeParser).MockScripts = []string{"testing_script"}
	f.parser.(*parser.FakeParser).MockParseAvailableScriptsError = errors.New("parsing error")
	cmd := NewRunCommand(f)

	scripts, _ = cmd.ValidArgsFunction(cmd, []string{}, "")

	if scripts != nil {
		t.Errorf("expecting no suggestion, got %v", scripts)
	}
}

func TestNewRunCommandSetVariableArgument(t *testing.T) {
	f := newFakeKoolRun([]builder.Command{&builder.FakeCommand{}}, nil)
	f.parser.(*parser.FakeParser).MockVariables = []string{"foo"}

	cmd := NewRunCommand(f)
	cmd.SetArgs([]string{"script", "--foo=bar"})

	if err := cmd.Execute(); err != nil {
		t.Errorf("unexpected error executing run command; error: %v", err)
	}

	if !f.parser.(*parser.FakeParser).CalledLookUpVariables {
		t.Error("did not call LookUpVariables on parser.Parser")
	}

	if fooVal, hasFoo := f.envStorage.(*environment.FakeEnvStorage).Envs["foo"]; !hasFoo || fooVal != "bar" {
		t.Error("failed to set the variable 'foo'")
	}
}

func TestNewRunCommandAskForVariableValue(t *testing.T) {
	f := newFakeKoolRun([]builder.Command{&builder.FakeCommand{}}, nil)
	f.parser.(*parser.FakeParser).MockVariables = []string{"foo"}
	f.promptInput.(*shell.FakePromptInput).MockAnswer = map[string]string{
		"There is no value for variable 'foo'. Please, type one:": "bar",
	}

	cmd := NewRunCommand(f)
	cmd.SetArgs([]string{"script"})

	if err := cmd.Execute(); err != nil {
		t.Errorf("unexpected error executing run command; error: %v", err)
	}

	if !f.promptInput.(*shell.FakePromptInput).CalledAsk {
		t.Error("did not call Ask on shell.PromptInput")
	}

	if fooVal, hasFoo := f.envStorage.(*environment.FakeEnvStorage).Envs["foo"]; !hasFoo || fooVal != "bar" {
		t.Error("failed to set the variable 'foo'")
	}
}

func TestNewRunCommandErrorAskForVariableValue(t *testing.T) {
	f := newFakeKoolRun([]builder.Command{&builder.FakeCommand{}}, nil)
	f.parser.(*parser.FakeParser).MockVariables = []string{"foo"}
	f.promptInput.(*shell.FakePromptInput).MockError = map[string]error{
		"There is no value for variable 'foo'. Please, type one:": errors.New("error prompt input"),
	}

	cmd := NewRunCommand(f)
	cmd.SetArgs([]string{"script"})

	if err := cmd.Execute(); err != nil {
		t.Errorf("unexpected error executing run command; error: %v", err)
	}

	if !f.shell.(*shell.FakeShell).CalledError {
		t.Error("did not call Error for prompt input error")
	}

	expectedError := "error prompt input"

	if gotError := f.shell.(*shell.FakeShell).Err.Error(); gotError != expectedError {
		t.Errorf("expecting error '%s', got '%s'", expectedError, gotError)
	}

	if !f.exiter.(*shell.FakeExiter).Exited() {
		t.Error("got a prompt input error, but command did not exit")
	}
}

func TestNewRunCommandNonTtyAskForVariableValue(t *testing.T) {
	f := newFakeKoolRun([]builder.Command{&builder.FakeCommand{}}, nil)
	f.term.(*shell.FakeTerminalChecker).MockIsTerminal = false

	cmd := NewRunCommand(f)
	cmd.SetArgs([]string{"script"})

	if err := cmd.Execute(); err != nil {
		t.Errorf("unexpected error executing run command; error: %v", err)
	}

	if f.parser.(*parser.FakeParser).CalledLookUpVariables {
		t.Error("should not call LookUpVariables on parser.Parser on Non-Tty environment")
	}
}
