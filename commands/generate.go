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
	"errors"
	"fmt"

	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

const generateExample = `- $ %[1]s generate <load-test-name> --script path/to/test-script
- $ %[1]s generate <load-test-name> -s path/to/test-script
- $ %[1]s generate <load-test-name> -s path/to/test-script [--out ]
- $ %[1]s generate <load-test-name> -s path/to/test-script [-o ]`

func newCmdGenerate(io genericclioptions.IOStreams, cliName string) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "generate [OPTIONS]",
		Aliases: []string{"gen"},
		Short:   "Generates load test manifests and wires dependencies in a kustomization.yaml file",
		Example: formatCmdExample(generateExample, cliName),
		RunE:    makeRunGenerate(io),
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	flags.StringP(
		"script",
		"s",
		"",
		"Specify path to artillery test-script file",
	)

	flags.StringP(
		"out",
		"o",
		"",
		"Optional. Specify output path to write load test manifests and kustomization.yaml",
	)

	if err := cmd.MarkFlagRequired("script"); err != nil {
		return nil
	}

	return cmd
}

func makeRunGenerate(io genericclioptions.IOStreams) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if err := validate(args); err != nil {
			return err
		}

		testScriptPath, err := cmd.Flags().GetString("script")
		if err != nil {
			return err
		}

		outPath, err := cmd.Flags().GetString("out")
		if err != nil {
			return err
		}

		msg := fmt.Sprintf("  >> load test name: [%s]\n  >> test-script path: [%s]", args[0], testScriptPath)
		if len(outPath) > 0 {
			msg = fmt.Sprintf("%s\n  >> test-script path: [%s]", msg, outPath)
		}

		_, _ = io.Out.Write([]byte(msg))
		_, _ = io.Out.Write([]byte("\n"))
		return nil
	}
}

func validate(args []string) error {
	if len(args) == 0 {
		return errors.New("missing load test name")
	}
	if len(args) > 1 {
		return errors.New("unknown arguments detected")
	}
	return nil
}

func formatCmdExample(doc, cliName string) string {
	return fmt.Sprintf(doc, cliName)
}
