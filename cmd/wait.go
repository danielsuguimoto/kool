package cmd

import (
	"kool-dev/kool/cmd/shell"
	"fmt"
	"os"
	"sync"
	"time"
	"net/url"

	"github.com/spf13/cobra"
)

// WaitFlags holds the flags for the wait command
type WaitFlags struct {
	Timeout, Interval time.Duration
}

// waitCmd represents the wait command
var waitCmd = &cobra.Command{
	Use:   "wait",
	Short: "Wait for resources availability",
	Args:  cobra.MinimumNArgs(1),
	Run: runWait,
}

var waitFlags = &WaitFlags{60 * time.Second, time.Second}

func init() {
	rootCmd.AddCommand(waitCmd)

	waitCmd.Flags().DurationVar(&waitFlags.Timeout, "timeout", 60 * time.Second, "Timeout")
	waitCmd.Flags().DurationVar(&waitFlags.Interval, "interval", time.Second, "Check interval")
}

func runWait(cmd *cobra.Command, args []string) {
	var (
		connects chan int = make(chan int)
		waitGroup sync.WaitGroup
		urls []url.URL
	)

	for _, arg := range args {
		u, err := url.Parse(arg)

		if err != nil {
			execError("", err)
			os.Exit(1)
		}

		urls = append(urls, *u)
	}

	_, err := shell.Exec("docker", "network", "create", "-d", "bridge", "ping_network")

	if err != nil {
		execError("", err)
		os.Exit(1)
	}

	go func(urls []url.URL) {
		for _, u := range urls {
			fmt.Println("Waiting for:", u.String())
			waitGroup.Add(1)

			go func(u url.URL) {
				defer waitGroup.Done()

				serviceID, err := shell.Exec("docker-compose", "ps", "-q", u.Hostname())

				if err != nil {
					execError("", err)
					os.Exit(1)
				}

				containerName, err := shell.Exec("docker", "ps", "-f", "ID=" + serviceID, "--format", "{{.Names}}")

				if err != nil {
					execError("", err)
					os.Exit(1)
				}

				shell.Exec("docker", "network", "connect", "--alias=" + u.Hostname(), "ping_network", containerName)

				for {
					_, err := shell.Exec("docker", "run", "-it", "--rm", "--network=ping_network", "kooldev/bash", "nc", "-z", u.Hostname(), u.Port())

					if err == nil {
						fmt.Println("Connected to ", u.String())
						connects <- 1
						break
					}

					fmt.Println("Problem with dial:", err.Error())
					fmt.Println("Sleeping", waitFlags.Interval)
					time.Sleep(waitFlags.Interval)
				}

				shell.Exec("docker", "network", "disconnect", "ping_network", containerName)
			}(u)
		}

		waitGroup.Wait()
	}(urls)

	select {
	case <-connects:
		{
			shell.Exec("docker", "network", "rm", "ping_network")
			break
		}
	case <-time.After(waitFlags.Timeout):
		{
			fmt.Println("timeout waiting for resources")
			shell.Exec("docker", "network", "rm", "ping_network")
			break
		}
	}
}
