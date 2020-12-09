package checker

import (
	"kool-dev/kool/cmd/builder"
	"kool-dev/kool/cmd/shell"
)

// Checker defines the check kool dependencies method
type Checker interface {
	Check() error
}

// DefaultChecker holds commands to be checked.
type DefaultChecker struct {
	dockerCmd        builder.Command
	dockerComposeCmd builder.Command
	shell            shell.Shell
}

// NewChecker initializes checker
func NewChecker(s shell.Shell) *DefaultChecker {
	var dockerInfoCmd, dockerComposePsCmd *builder.DefaultCommand

	dockerInfoCmd = builder.NewCommand("docker", "info")
	dockerComposePsCmd = builder.NewCommand("docker-compose", "ps")

	return &DefaultChecker{dockerInfoCmd, dockerComposePsCmd, s}
}

// Check checks kool dependencies
func (c *DefaultChecker) Check() error {
	if err := c.shell.LookPath(c.dockerCmd); err != nil {
		return ErrDockerNotFound
	}

	if err := c.shell.LookPath(c.dockerComposeCmd); err != nil {
		return ErrDockerComposeNotFound
	}

	if _, err := c.shell.Exec(c.dockerCmd); err != nil {
		return ErrDockerNotRunning
	}

	return nil
}
