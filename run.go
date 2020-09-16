package main

import (
	"fmt"
	"os"
	"os/exec"
)

func Run(record *CommandRecord) error {
	fmt.Fprintln(os.Stderr, "running command '"+record.ID+"': "+record.Command)

	cmd := exec.Command("bash", "-c", record.Command)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return err
	}
	err := cmd.Wait()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			// Copy the exit error code.
			os.Exit(exitErr.ExitCode())
		}
	}
	return err
}
