package cmd

import (
	"errors"
	"fmt"
	"io/ioutil"
	"kool-dev/kool/cmd/shell"
	"kool-dev/kool/cmd/updater"
	"testing"
)

func newFakeKoolSelfUpdate(currentVersion string, latestVersion string, err error) *KoolSelfUpdate {
	selfUpdate := &KoolSelfUpdate{
		*newFakeKoolService(),
		&updater.FakeUpdater{
			MockCurrentVersion: currentVersion,
			MockLatestVersion:  latestVersion,
			MockError:          err,
		},
	}

	selfUpdate.shell.(*shell.FakeShell).MockOutStream = ioutil.Discard
	return selfUpdate
}

func TestNewKoolSelfUpdate(t *testing.T) {
	k := NewKoolSelfUpdate()

	if _, ok := k.DefaultKoolService.shell.(*shell.DefaultShell); !ok {
		t.Errorf("unexpected shell.Shell on KoolSelfUpdate KoolRun instance")
	}

	if _, ok := k.DefaultKoolService.exiter.(*shell.DefaultExiter); !ok {
		t.Errorf("unexpected shell.Exiter on default KoolSelfUpdate instance")
	}

	if _, ok := k.DefaultKoolService.term.(*shell.DefaultTerminalChecker); !ok {
		t.Errorf("unexpected shell.TerminalChecker on KoolSelfUpdate KoolRun instance")
	}

	if _, ok := k.updater.(*updater.DefaultUpdater); !ok {
		t.Errorf("unexpected updater.Updater on default KoolSelfUpdate instance")
	}
}

func TestNewSelfUpdateCommand(t *testing.T) {
	f := newFakeKoolSelfUpdate("0.0.0", "1.0.0", nil)
	cmd := NewSelfUpdateCommand(f)

	if err := cmd.Execute(); err != nil {
		t.Errorf("unexpected error executing self-update command; error: %v", err)
	}

	if !f.updater.(*updater.FakeUpdater).CalledGetCurrentVersion {
		t.Errorf("did not call GetCurrentVersion")
	}

	if !f.updater.(*updater.FakeUpdater).CalledUpdate {
		t.Errorf("did not call Update")
	}

	if !f.shell.(*shell.FakeShell).CalledSuccess {
		t.Errorf("did not call Success for updating successfully")
	}

	expected := "Successfully updated to version 1.0.0"

	if output := fmt.Sprint(f.shell.(*shell.FakeShell).SuccessOutput...); output != expected {
		t.Errorf("expecting success message '%s', got '%s'", expected, output)
	}
}

func TestNewSelfUpdateUpToDateCommand(t *testing.T) {
	f := newFakeKoolSelfUpdate("1.0.0", "1.0.0", nil)
	cmd := NewSelfUpdateCommand(f)

	if err := cmd.Execute(); err != nil {
		t.Errorf("unexpected error executing self-update command; error: %v", err)
	}

	if !f.shell.(*shell.FakeShell).CalledWarning {
		t.Errorf("did not call Warning for already having latest version")
	}

	if f.shell.(*shell.FakeShell).CalledSuccess {
		t.Errorf("unexpected update successful message when already having latest version")
	}

	expected := "You already have the latest version 1.0.0"

	if output := fmt.Sprint(f.shell.(*shell.FakeShell).WarningOutput...); output != expected {
		t.Errorf("expecting warning message '%s', got '%s'", expected, output)
	}
}

func TestNewSelfUpdateErrorCommand(t *testing.T) {
	f := newFakeKoolSelfUpdate("1.0.0", "1.0.0", errors.New("error"))
	cmd := NewSelfUpdateCommand(f)

	if err := cmd.Execute(); err != nil {
		t.Errorf("unexpected error executing self-update command; error: %v", err)
	}

	if f.shell.(*shell.FakeShell).CalledWarning {
		t.Errorf("unexpected warning message for failed update")
	}

	if f.shell.(*shell.FakeShell).CalledSuccess {
		t.Errorf("unexpected update successful message for failed update")
	}

	if !f.shell.(*shell.FakeShell).CalledError {
		t.Errorf("did not call Error for failed update")
	}

	expected := "kool self-update failed: error"

	if output := f.shell.(*shell.FakeShell).Err.Error(); output != expected {
		t.Errorf("expecting error message '%s', got '%s'", expected, output)
	}

	if !f.exiter.(*shell.FakeExiter).Exited() {
		t.Errorf("did not exited after failing update")
	}
}
