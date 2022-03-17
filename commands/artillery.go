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
	"io"

	"github.com/artilleryio/artillery-operator/internal/artillery"
	"github.com/artilleryio/artillery-operator/internal/telemetry"
	"github.com/posthog/posthog-go"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

const cliName = "kubectl artillery"

func NewCmdArtillery(
	workingDir string,
	io genericclioptions.IOStreams,
	tClient posthog.Client,
	tCfg telemetry.Config,
) *cobra.Command {

	cmd := &cobra.Command{
		Short:        "Use artillery.io operator helpers",
		Use:          "artillery",
		SilenceUsage: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return highlightTelemetryIfRequired(io.Out)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 0 {
				return fmt.Errorf("%q is not a %[1]s command\nSee '%[1]s --help'", args[0], cliName)
			}
			cmd.HelpFunc()(cmd, args)
			return nil
		},
	}

	cmd.AddCommand(newCmdGenerate(workingDir, io, cliName, tClient, tCfg))

	return cmd
}

func highlightTelemetryIfRequired(out io.Writer) error {
	settings, err := artillery.GetOrCreateCLISettings()
	if err != nil {
		return err
	}

	if !settings.GetFirstRun() {
		return nil
	}

	_, _ = out.Write([]byte("Telemetry is on. Learn more: https://artillery.io/docs/resources/core/telemetry.html\n"))

	return settings.SetFirstRun(false).Save()
}
