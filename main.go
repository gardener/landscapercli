// SPDX-FileCopyrightText: 2020 SAP SE or an SAP affiliate company and Gardener contributors.
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"fmt"
	"os"

	"github.com/gardener/landscapercli/cmd"
)

func main() {
	ctx := context.Background()
	defer ctx.Done()

	landscaperCliCmd := cmd.NewLandscaperCliCommand(ctx)

	if err := landscaperCliCmd.Execute(); err != nil {
		fmt.Print(err)
		os.Exit(1)
	}
}
