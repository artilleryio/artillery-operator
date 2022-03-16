/*
 * Copyright (c) 2022.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0.
 *
 * If a copy of the MPL was not distributed with
 * this file, You can obtain one at
 *
 *   http://mozilla.org/MPL/2.0/
 */

package commands

import (
	"fmt"

	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

const cliName = "kubectl artillery"

func NewCmdArtillery(workingDir string, io genericclioptions.IOStreams) *cobra.Command {

	cmd := &cobra.Command{
		Short:        "Use artillery.io operator helpers",
		Use:          "artillery",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 0 {
				return fmt.Errorf("%q is not a %[1]s command\nSee '%[1]s --help'", args[0], cliName)
			}
			cmd.HelpFunc()(cmd, args)
			return nil
		},
	}

	cmd.AddCommand(newCmdGenerate(workingDir, io, cliName))

	return cmd
}