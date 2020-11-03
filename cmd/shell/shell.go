package shell

import (
	"fmt"
	"io"
	"kool-dev/kool/environment"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
)

// DefaultShell holds data for shell handling
type DefaultShell struct {
	in       io.ReadCloser
	out      io.WriteCloser
	err      io.WriteCloser
	lookedUp map[string]bool
}

// Shell holds logic for shell handling
type Shell interface {
	In() io.ReadCloser
	Out() io.WriteCloser
	Err() io.WriteCloser
	Exec(string, ...string) (string, error)
	Interactive(string, ...string) error
}

// NewShell creates a new shell
func NewShell() *DefaultShell {
	return &DefaultShell{
		in:  os.Stdin,
		out: os.Stdout,
		err: os.Stderr,
	}
}

// In shell input
func (s *DefaultShell) In() io.ReadCloser {
	return s.in
}

// Out shell output
func (s *DefaultShell) Out() io.WriteCloser {
	return s.out
}

// Err shell error output
func (s *DefaultShell) Err() io.WriteCloser {
	return s.err
}

// Exec will execute the given command silently and return the combined
// error/standard output, and an error if any.
func (s *DefaultShell) Exec(exe string, args ...string) (outStr string, err error) {
	var (
		cmd *exec.Cmd
		out []byte
	)

	if exe == "docker-compose" {
		args = append(dockerComposeDefaultArgs(), args...)
	}

	cmd = exec.Command(exe, args...)
	cmd.Env = os.Environ()
	cmd.Stdin = s.In()

	out, err = cmd.CombinedOutput()
	outStr = strings.TrimSpace(string(out))
	return
}

// Interactive runs the given command proxying current Stdin/Stdout/Stderr
// which makes it interactive for running even something like `bash`.
func (s *DefaultShell) Interactive(exe string, args ...string) (err error) {
	var (
		cmd            *exec.Cmd
		parsedRedirect *DefaultParsedRedirect
		outputWriter   OutputWriter
	)

	outputWriter = NewOutputWriter()

	if exe == "docker-compose" {
		args = append(dockerComposeDefaultArgs(), args...)
	}

	if environment.NewEnvStorage().IsTrue("KOOL_VERBOSE") {
		fmt.Println("$", exe, strings.Join(args, " "))
	}

	// soon should refactor this onto a struct with methods
	// so we can remove this too long list of returned values.
	if parsedRedirect, err = parseRedirects(s, args); err != nil {
		return
	}

	defer parsedRedirect.Close()

	cmd = parsedRedirect.CreateCommand(exe)

	if err = s.lookPath(exe); err != nil {
		outputWriter.Error(fmt.Errorf("failed to run %s error: %v", cmd.String(), err))
		os.Exit(2)
	}

	err = cmd.Start()

	if err != nil {
		return
	}

	waitCh := make(chan error, 1)
	go func() {
		waitCh <- cmd.Wait()
		close(waitCh)
	}()
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan)

	// You need a for loop to handle multiple signals
	for {
		select {
		case err = <-waitCh:
			// Subprocess exited. Get the return code, if we can
			var waitStatus syscall.WaitStatus
			if exitError, ok := err.(*exec.ExitError); ok {
				waitStatus = exitError.Sys().(syscall.WaitStatus)
				os.Exit(waitStatus.ExitStatus())
			}
			if err != nil {
				log.Fatal(err)
			}
			return
		case sig := <-sigChan:
			if err := cmd.Process.Signal(sig); err != nil {
				// check if it is something we should care about
				if err.Error() != "os: process already finished" {
					outputWriter.Error(fmt.Errorf("error sending signal to child process %v %v", sig, err))
				}
			}
		}
	}
}

func (s *DefaultShell) lookPath(exe string) (err error) {
	if s.lookedUp == nil {
		s.lookedUp = make(map[string]bool)
	}

	if exe != "kool" && !s.lookedUp[exe] && !strings.HasPrefix(exe, "./") && !strings.HasPrefix(exe, "/") {
		// non-kool and non-absolute/relative path... let's look it up
		_, err = exec.LookPath(exe)

		s.lookedUp[exe] = true
	}
	return
}

func dockerComposeDefaultArgs() []string {
	return []string{"-p", environment.NewEnvStorage().Get("KOOL_NAME")}
}
