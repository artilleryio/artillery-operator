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
	"path/filepath"
	"strings"

	"github.com/artilleryio/artillery-operator/api/v1alpha1"
	"github.com/artilleryio/artillery-operator/internal/artillery"
	"github.com/artilleryio/artillery-operator/internal/telemetry"
	"github.com/posthog/posthog-go"
	"github.com/spf13/cobra"
	k8sValidation "k8s.io/apimachinery/pkg/api/validation"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

const loadtestExample = `- $ %[1]s generate loadtest <name> --script path/to/test-script
- $ %[1]s generate loadtest <name> -s path/to/test-script
- $ %[1]s generate loadtest <name> -s path/to/test-script [--env ] [--out ] [--count ]`

func newCmdLoadTest(
	workingDir string,
	io genericclioptions.IOStreams,
	cliName string,
	tClient posthog.Client,
	tCfg telemetry.Config,
) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "loadtest [OPTIONS]",
		Aliases: []string{"lt"},
		Short:   "Generates load test manifests and wires dependencies in a kustomization.yaml file",
		Example: formatCmdExample(loadtestExample, cliName),
		RunE:    makeRunLoadTest(workingDir, io),
		PostRunE: func(cmd *cobra.Command, args []string) error {
			testScriptPath, _ := cmd.Flags().GetString("script")
			env, _ := cmd.Flags().GetString("env")
			outPath, _ := cmd.Flags().GetString("out")
			count, _ := cmd.Flags().GetInt("count")

			logger := artillery.NewIOLogger(io.Out, io.ErrOut)
			telemetry.TelemeterGenerateManifests(args[0], testScriptPath, env, outPath, count, tClient, tCfg, logger)
			return nil
		},
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
		"env",
		"e",
		"dev",
		"Optional. Specify the load test environment - defaults to dev",
	)

	flags.StringP(
		"out",
		"o",
		"",
		"Optional. Specify output path to write load test manifests and kustomization.yaml",
	)

	flags.IntP(
		"count",
		"c",
		1,
		"Optional. Specify number of load test workers",
	)

	if err := cmd.MarkFlagRequired("script"); err != nil {
		return nil
	}

	return cmd
}

func makeRunLoadTest(workingDir string, io genericclioptions.IOStreams) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if err := validateLoadTest(args); err != nil {
			return err
		}

		testScriptPath, err := cmd.Flags().GetString("script")
		if err != nil {
			return err
		}

		if err := validateLoadTestScript(testScriptPath); err != nil {
			return err
		}

		env, err := cmd.Flags().GetString("env")
		if err != nil {
			return err
		}

		outPath, err := cmd.Flags().GetString("out")
		if err != nil {
			return err
		}

		count, err := cmd.Flags().GetInt("count")
		if err != nil {
			return err
		}

		loadTestName := args[0]
		configMapName := fmt.Sprintf("%s-test-script", loadTestName)

		targetDir, err := artillery.MkdirAllTargetOrDefault(workingDir, outPath, artillery.DefaultManifestDir)
		if err != nil {
			return err
		}

		if err := artillery.CopyFileTo(targetDir, testScriptPath); err != nil {
			return err
		}

		loadTest := v1alpha1.NewLoadTest(loadTestName, configMapName, env, count)
		kustomization := artillery.NewKustomization(artillery.LoadTestFilename, configMapName, testScriptPath, artillery.LabelPrefix)

		msg, err := artillery.Generatables{
			{
				Path:      filepath.Join(targetDir, artillery.LoadTestFilename),
				Marshaler: loadTest,
			},
			{
				Path:      filepath.Join(targetDir, "kustomization.yaml"),
				Marshaler: kustomization,
			},
		}.Generate(2)
		if err != nil {
			return err
		}

		_, _ = io.Out.Write([]byte(msg))
		_, _ = io.Out.Write([]byte("\n"))
		return nil
	}
}

func validateLoadTest(args []string) error {
	if len(args) == 0 {
		return errors.New("missing load test name")
	}
	if len(args) > 1 {
		return errors.New("unknown arguments detected")
	}

	loadTestName := args[0]
	invalids := k8sValidation.NameIsDNSSubdomain(loadTestName, false)
	if len(invalids) > 0 {
		return fmt.Errorf("load test name %s must be a valid DNS subdomain name, \n%s", loadTestName, strings.Join(invalids, "\n- "))
	}

	return nil
}

func validateLoadTestScript(s string) error {
	absPath, err := filepath.Abs(s)
	if err != nil {
		return err
	}

	if !artillery.DirOrFileExists(absPath) {
		return fmt.Errorf("cannot find script file %s ", s)
	}

	return nil
}
