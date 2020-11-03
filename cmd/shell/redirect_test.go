package shell

import (
	"os"
	"path/filepath"
	"testing"
)

type fakeReaderWriterCloser struct {
	calledClose bool
}

func (f *fakeReaderWriterCloser) Close() error {
	f.calledClose = true
	return nil
}

func (f *fakeReaderWriterCloser) Read(p []byte) (n int, err error) {
	return
}

func (f *fakeReaderWriterCloser) Write(p []byte) (n int, err error) {
	return
}

func TestParseRedirectParseNoRedirects(t *testing.T) {
	// test no redirects
	p, err := parseRedirects(NewShell(), []string{"foo", "bar"})

	if err != nil {
		t.Errorf("unexpected error parsing redirects")
	}

	if p.closeStdin || p.closeStdout {
		t.Errorf("bad parse - should not close in/outs")
	}

	// test input redirect
	input := filepath.Join(t.TempDir(), "input")
	file, _ := os.Create(input)
	file.Close()
	p, err = parseRedirects(NewShell(), []string{"foo", "<", input})

	if err != nil {
		t.Errorf("unexpected error parsing redirects")
	}

	if !p.closeStdin || p.closeStdout {
		t.Errorf("bad parse - should close in; not out")
	}

	p.Close()

	// test output
	output := filepath.Join(t.TempDir(), "output")
	file, _ = os.Create(output)
	file.Close()
	p, err = parseRedirects(NewShell(), []string{"foo", ">", output})

	if err != nil {
		t.Errorf("unexpected error parsing redirects")
	}

	if p.closeStdin || !p.closeStdout {
		t.Errorf("bad parse - should close in; not out")
	}

	p.Close()
}

func TestParsedRedirectCreateCommand(t *testing.T) {
	p := &DefaultParsedRedirect{
		in:          &fakeReaderWriterCloser{},
		out:         &fakeReaderWriterCloser{},
		args:        []string{"arg1", "arg2"},
		closeStdin:  false,
		closeStdout: false,
	}

	exe := "foo"
	cmd := p.CreateCommand(exe)

	if cmd == nil {
		t.Errorf("failed to create command")
		return
	}

	if cmd.Args == nil || cmd.Args[0] != exe || cmd.Args[1] != "arg1" || cmd.Args[2] != "arg2" {
		t.Errorf("bad command/arguments for created Commands")
	}
}

func TestParsedRedirectCloses(t *testing.T) {
	p := &DefaultParsedRedirect{
		in:          &fakeReaderWriterCloser{},
		out:         &fakeReaderWriterCloser{},
		closeStdin:  false,
		closeStdout: false,
	}

	// calls close - should not clode in/out
	p.Close()

	if p.in.(*fakeReaderWriterCloser).calledClose {
		t.Errorf("did not expect to call close on Stdin")
	}
	if p.out.(*fakeReaderWriterCloser).calledClose {
		t.Errorf("did not expect to call close on Stdout")
	}

	p.closeStdin = true
	p.closeStdout = true

	// calls close - should close in/out
	p.Close()

	if !p.in.(*fakeReaderWriterCloser).calledClose {
		t.Errorf("did not get expected call to close on Stdin")
	}
	if !p.out.(*fakeReaderWriterCloser).calledClose {
		t.Errorf("did not get expected call to close on Stdout")
	}
}
