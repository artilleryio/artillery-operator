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

	"github.com/artilleryio/artillery-operator/internal/telemetry"
	"github.com/posthog/posthog-go"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

func newCmdGenerate(
	workingDir string,
	io genericclioptions.IOStreams,
	cliName string,
	tClient posthog.Client,
	tCfg telemetry.Config,
) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "generate [command]",
		Short:   fmt.Sprintf("The '%s generate' command invokes a specific generator to generate\na loadtest or test scripts.", cliName),
		Aliases: []string{"gen"},
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	cmd.AddCommand(newCmdLoadTest(workingDir, io, cliName, tClient, tCfg))

	return cmd
}

func formatCmdExample(doc, cliName string) string {
	return fmt.Sprintf(doc, cliName)
}
