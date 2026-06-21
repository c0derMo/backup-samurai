package handlers

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"slices"
)

func TestExitCodes(err error, allowedExitCodes []int) bool {
	exit_error := &exec.ExitError{}
	if errors.As(err, &exit_error) {
		if slices.Contains(allowedExitCodes, exit_error.ProcessState.ExitCode()) {
			return true
		}
	}
	return false
}

func FormatExitError(err error, message string, commandOutput []byte) error {
	exit_error := &exec.ExitError{}
	if errors.As(err, &exit_error) {
		lastLine := ExtractLastLine(commandOutput)
		return fmt.Errorf("%s: %s", message, lastLine)
	}
	return err
}

func ExtractLastLine(output []byte) string {
	toCheck := output
	if output[len(output)-1] == '\n' {
		toCheck = output[:len(output)-1]
	}
	lastLinebreak := bytes.LastIndex(toCheck, []byte{'\n'})
	if lastLinebreak < 0 {
		return string(toCheck)
	} else {
		return string(toCheck[lastLinebreak:])
	}
}
