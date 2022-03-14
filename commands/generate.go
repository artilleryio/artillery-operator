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
	"bytes"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/artilleryio/artillery-operator/api/v1alpha1"
	"github.com/spf13/cobra"
	yaml3 "gopkg.in/yaml.v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

const generateExample = `- $ %[1]s generate <load-test-name> --script path/to/test-script
- $ %[1]s generate <load-test-name> -s path/to/test-script
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
		//TODO: validate configMapName is a valid K8s name

		msg := fmt.Sprintf("  >> load test name: [%s]\n  >> test-script path: [%s]", loadTestName, testScriptPath)
		if len(outPath) > 0 {
			msg = fmt.Sprintf("%s\n  >> test-script path: [%s]", msg, outPath)
		}

		kind := "LoadTest"
		apiVersion := "loadtest.artillery.io/v1alpha1"

		meta := metav1.TypeMeta{
			Kind:       kind,
			APIVersion: apiVersion,
		}
		objectMeta := metav1.ObjectMeta{
			Name: loadTestName,
		}
		testScript := v1alpha1.TestScript{
			Config: v1alpha1.Config{
				ConfigMap: configMapName,
			},
		}
		spec := v1alpha1.LoadTestSpec{
			Count:       count,
			Environment: "CHANGE_ME",
			TestScript:  testScript,
		}

		loadTest := &v1alpha1.LoadTest{
			TypeMeta:   meta,
			ObjectMeta: objectMeta,
			Spec:       spec,
		}

		out, err := getOrCreateManifestsPath(workingDir, outPath)
		if err != nil {
			return err
		}

		manifestPath := path.Join(out, "loadtest-cr.yaml")
		written, err := writeTo(manifestPath, loadTest, 2)
		if err != nil {
			return err
		}

		if written > 0 {
			msg = fmt.Sprintf("%s\n  >> loadtest manifests created at: [%s]", msg, out)
		}

		_, _ = io.Out.Write([]byte(msg))
		_, _ = io.Out.Write([]byte("\n"))
		return nil
	}
}

func getOrCreateManifestsPath(workingDir string, outPath string) (string, error) {
	result := outPath

	if len(result) == 0 {
		result = path.Join(workingDir, "artillery-manifests")
	}

	return result, os.MkdirAll(result, 0700)
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

func writeTo(filePath string, v *v1alpha1.LoadTest, indent int) (n int64, err error) {
	data, err := marshalWithIndent(v, indent)
	if err != nil {
		return int64(0), err
	}

	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return int64(0), err
	}

	file, err := os.Create(absPath)
	if err != nil {
		return int64(0), err
	}

	written, err := file.Write(data)
	if err != nil {
		return int64(0), err
	}

	return int64(written), file.Close()
}

func marshalWithIndent(v *v1alpha1.LoadTest, indent int) ([]byte, error) {
	j, err := v.GenerateJson()
	if err != nil {
		return nil, err
	}

	y, err := jsonToYaml(j, indent)
	if err != nil {

		return nil, err
	}

	return y, nil
}

func jsonToYaml(j []byte, spaces int) ([]byte, error) {
	// Convert the JSON to an object.
	var jsonObj interface{}
	// We are using yaml.Unmarshal here (instead of json.Unmarshal) because the
	// Go JSON library doesn't try to pick the right number type (int, float,
	// etc.) when unmarshling to interface{}, it just picks float64
	// universally. go-yaml does go through the effort of picking the right
	// number type, so we can preserve number type throughout this process.
	err := yaml3.Unmarshal(j, &jsonObj)
	if err != nil {
		return nil, err
	}

	var b bytes.Buffer
	encoder := yaml3.NewEncoder(&b)
	encoder.SetIndent(spaces)
	if err := encoder.Encode(jsonObj); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}
