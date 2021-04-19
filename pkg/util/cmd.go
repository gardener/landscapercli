package util

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// ExecCommandBlocking executes a command and wait for its completion.
func ExecCommandBlocking(command string) error {
	fmt.Printf("Executing: %s\n", command)

	arr := strings.Split(command, " ")

	if arr[0] == "helm" {
		helmPath := os.Getenv("HELM_EXECUTABLE")
		if helmPath != "" {
			arr[0] = helmPath
			fmt.Printf("Using helm binary: %s\n", arr[0])
		}
	}

	cmd := exec.Command(arr[0], arr[1:]...)
	cmd.Env = []string{"HELM_EXPERIMENTAL_OCI=1", "HOME=" + os.Getenv("HOME"), "PATH=" + os.Getenv("PATH")}
	out, err := cmd.CombinedOutput()
	outStr := string(out)

	if err != nil {
		return fmt.Errorf("failed with error: %s:\n%s\n", err, outStr)
	}
	fmt.Println("Executed sucessfully!")

	return nil
}

type CmdResult struct {
	Error  error
	Stdout string
	StdErr string
}

// ExecCommandNonBlocking executes a command without without blocking. Returns a Cmd that can be used to stop the command.
// When the command has stopped or failed, the result is written into the channel resultCh.
func ExecCommandNonBlocking(command string, resultCh chan<- CmdResult) (*exec.Cmd, error) {
	fmt.Printf("Executing: %s\n", command)

	arr := strings.Split(command, " ")

	if arr[0] == "helm" {
		helmPath := os.Getenv("HELM_EXECUTABLE")
		if helmPath != "" {
			arr[0] = helmPath
			fmt.Printf("Using helm binary: %s\n", arr[0])
		}
	}

	outbuf := bytes.Buffer{}
	errbuf := bytes.Buffer{}

	cmd := exec.Command(arr[0], arr[1:]...)
	cmd.Env = []string{"HELM_EXPERIMENTAL_OCI=1", "HOME=" + os.Getenv("HOME"), "PATH=" + os.Getenv("PATH")}
	cmd.Stderr = &outbuf
	cmd.Stdout = &errbuf

	err := cmd.Start()
	if err != nil {
		fmt.Printf("Failed with error: %s:\n", err)
		return nil, err
	}
	fmt.Println("Started sucessfully!")

	go func() {
		exitErr := cmd.Wait()
		res := CmdResult{
			Error:  exitErr,
			Stdout: outbuf.String(),
			StdErr: outbuf.String(),
		}
		resultCh <- res
		close(resultCh)
	}()

	return cmd, nil
}
