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

func NewCmdArtillery(io genericclioptions.IOStreams) *cobra.Command {
	cmd := &cobra.Command{
		Short:        "Use artillery.io operator helpers",
		Use:          "artillery",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 0 {
				return fmt.Errorf("%q is not a %[1]s command\nSee '%[1]s --help'", args[0], "kubectl artillery")
			}
			cmd.HelpFunc()(cmd, args)
			_, _ = io.Out.Write([]byte("\n"))
			_, _ = io.Out.Write([]byte("kubectl artillery says hello"))
			return nil
		},
	}
	return cmd
}
