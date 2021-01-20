package util

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func GetValueFromNestedMap(data map[string]interface{}, valuePath string) (interface{}, error) {
	var val interface{}
	var ok bool

	keys := strings.Split(valuePath, ".")
	for index, key := range keys {
		if index == len(keys)-1 {
			val, ok = data[key]
			if !ok {
				return nil, fmt.Errorf("Cannot get value for path %s", valuePath)
			}
		} else {
			tmp := data[key]
			data, ok = tmp.(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("Cannot get value for path %s", valuePath)
			}
		}
	}

	return val, nil
}

func ExecCommandBlocking(command string) error {
	fmt.Printf("Executing: %s\n", command)

	arr := strings.Split(command, " ")

	cmd := exec.Command(arr[0], arr[1:]...)
	cmd.Env = []string{"HELM_EXPERIMENTAL_OCI=1", "HOME=" + os.Getenv("HOME"), "PATH=" + os.Getenv("PATH")}
	out, err := cmd.CombinedOutput()

	if err != nil {
		fmt.Printf("Failed with error: %s:\n%s\n", err, string(out))
		return err
	}
	fmt.Println("Executed sucessfully!")

	return nil
}

func ExecCommandNonBlocking(command string) (*exec.Cmd, error) {
	fmt.Printf("Executing: %s\n", command)

	arr := strings.Split(command, " ")

	cmd := exec.Command(arr[0], arr[1:]...)
	cmd.Env = []string{"HELM_EXPERIMENTAL_OCI=1", "HOME=" + os.Getenv("HOME"), "PATH=" + os.Getenv("PATH")}
	err := cmd.Start()

	if err != nil {
		fmt.Printf("Failed with error: %s:\n", err)
		return nil, err
	}
	fmt.Println("Executed sucessfully!")

	return cmd, nil
}
