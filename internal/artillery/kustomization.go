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

package artillery

import (
	"bytes"
	"fmt"
	"log"
	"path/filepath"

	yaml3 "gopkg.in/yaml.v3"
	"sigs.k8s.io/kustomize/api/types"
)

// Kustomization wrapper to enable marshaling a Kustomization to a file
type Kustomization struct {
	*types.Kustomization
}

// NewKustomization returns a configured Kustomization wrapper for a LoadTest
func NewKustomization(loadtest, configMap, testScript, labelPrefix string) *Kustomization {
	testScript = filepath.Base(testScript)

	k := &Kustomization{
		Kustomization: &types.Kustomization{
			TypeMeta: types.TypeMeta{
				Kind:       types.KustomizationKind,
				APIVersion: types.KustomizationVersion,
			},
			ConfigMapGenerator: []types.ConfigMapArgs{
				{
					GeneratorArgs: types.GeneratorArgs{
						Name: configMap,
						KvPairSources: types.KvPairSources{
							FileSources: []string{testScript},
						},
					},
				},
			},
			GeneratorOptions: &types.GeneratorOptions{
				DisableNameSuffixHash: true,
				Labels: map[string]string{
					"artillery.io/component": fmt.Sprintf("%s-config", labelPrefix),
					"artillery.io/part-of":   labelPrefix,
				},
			},
			Resources: []string{loadtest},
		},
	}

	return k
}

// MarshalWithIndent marshals a Kustomization using a specified indentation.
func (k *Kustomization) MarshalWithIndent(indent int) ([]byte, error) {
	var out bytes.Buffer
	encoder := yaml3.NewEncoder(&out)
	defer func(encoder *yaml3.Encoder) {
		err := encoder.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(encoder)
	encoder.SetIndent(indent)

	if err := encoder.Encode(k.Kustomization); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}
