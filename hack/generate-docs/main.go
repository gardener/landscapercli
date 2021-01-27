// SPDX-FileCopyrightText: 2020 SAP SE or an SAP affiliate company and Gardener contributors.
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra/doc"

	"github.com/gardener/landscapercli/cmd"
)

func main() {
	fmt.Println("> Generate Docs for LandscaperCli")

	if len(os.Args) != 2 { // expect 2 as the first one is the programm name
		fmt.Printf("Expected exactly one argument, but got %d", len(os.Args) - 1)
		os.Exit(1)
	}
	outputDir := os.Args[1]

	// clear generated
	check(os.RemoveAll(outputDir))
	check(os.MkdirAll(outputDir, os.ModePerm))

	landscaperCliCommand := cmd.NewLandscaperCliCommand(context.TODO())
	landscaperCliCommand.DisableAutoGenTag = true
	check(doc.GenMarkdownTree(landscaperCliCommand, outputDir))
	fmt.Printf("Successfully written docs to %s", outputDir)
}

func check(err error) {
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}