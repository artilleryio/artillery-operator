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
	"path"
	"path/filepath"
	"strings"

	"github.com/artilleryio/artillery-operator/api/v1alpha1"
	"github.com/artilleryio/artillery-operator/internal/artillery"
	"github.com/spf13/cobra"
	k8sValidation "k8s.io/apimachinery/pkg/api/validation"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

const generateExample = `- $ %[1]s generate <load-test-name> --script path/to/test-script
- $ %[1]s generate <load-test-name> -s path/to/test-script
- $ %[1]s generate <load-test-name> -s path/to/test-script [--env ]
- $ %[1]s generate <load-test-name> -s path/to/test-script [-e ]
- $ %[1]s generate <load-test-name> -s path/to/test-script [--out ]
- $ %[1]s generate <load-test-name> -s path/to/test-script [-o ]
- $ %[1]s generate <load-test-name> -s path/to/test-script [--out ] [--count ]
- $ %[1]s generate <load-test-name> -s path/to/test-script [--out ] [-c ]`

func newCmdGenerate(workingDir string, io genericclioptions.IOStreams, cliName string) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "generate [OPTIONS]",
		Aliases: []string{"gen"},
		Short:   "Generates load test manifests and wires dependencies in a kustomization.yaml file",
		Example: formatCmdExample(generateExample, cliName),
		RunE:    makeRunGenerate(workingDir, io),
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

func makeRunGenerate(workingDir string, io genericclioptions.IOStreams) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if err := validate(args); err != nil {
			return err
		}

		testScriptPath, err := cmd.Flags().GetString("script")
		if err != nil {
			return err
		}

		if err := validateScript(testScriptPath); err != nil {
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

		targetDir, err := artillery.GetOrCreateTargetDir(workingDir, outPath)
		if err != nil {
			return err
		}

		loadTest := v1alpha1.NewLoadTest(loadTestName, configMapName, env, count)
		kustomization := artillery.NewKustomization(artillery.LoadTestFilename, configMapName, testScriptPath, artillery.LabelPrefix)

		msg, err := artillery.Generatables{
			{
				Path:      path.Join(targetDir, artillery.LoadTestFilename),
				Marshaler: loadTest,
			},
			{
				Path:      path.Join(targetDir, "kustomization.yaml"),
				Marshaler: kustomization,
			},
		}.Write(2)
		if err != nil {
			return err
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

	loadTestName := args[0]
	invalids := k8sValidation.NameIsDNSSubdomain(loadTestName, false)
	if len(invalids) > 0 {
		return fmt.Errorf("load test name %s must be a valid DNS subdomain name, \n%s", loadTestName, strings.Join(invalids, "\n- "))
	}

	return nil
}

func validateScript(s string) error {
	absPath, err := filepath.Abs(s)
	if err != nil {
		return err
	}

	if !artillery.DirOrFileExists(absPath) {
		return fmt.Errorf("cannot find script file %s ", s)
	}

	return nil
}

func formatCmdExample(doc, cliName string) string {
	return fmt.Sprintf(doc, cliName)
}
