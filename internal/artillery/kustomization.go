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
	"encoding/json"
	"fmt"
	"path"
	"path/filepath"

	"sigs.k8s.io/kustomize/api/types"
)

type Kustomization struct {
	*types.Kustomization
}

func NewKustomization(configMapName string, testScriptPath string, labelPrefix string) *Kustomization {
	k := &Kustomization{
		Kustomization: &types.Kustomization{},
	}

	if !filepath.IsAbs(testScriptPath) {
		testScriptPath = path.Join("..", testScriptPath)
	}

	k.WithTypeMeta(types.TypeMeta{
		Kind:       "Kustomization",
		APIVersion: "kustomize.config.k8s.io/v1beta1",
	}).WithConfigMapGenerator([]types.ConfigMapArgs{
		{
			GeneratorArgs: types.GeneratorArgs{
				Name: configMapName,
				KvPairSources: types.KvPairSources{
					FileSources: []string{testScriptPath},
				},
			},
		},
	}).WithGeneratorOptions(&types.GeneratorOptions{
		DisableNameSuffixHash: true,
		Labels: map[string]string{
			"artillery.io/component": fmt.Sprintf("%s-config", labelPrefix),
			"artillery.io/part-of":   labelPrefix,
		},
	})

	return k
}

func (k *Kustomization) WithTypeMeta(meta types.TypeMeta) *Kustomization {
	k.TypeMeta = meta
	return k
}

func (k *Kustomization) WithConfigMapGenerator(c []types.ConfigMapArgs) *Kustomization {
	k.ConfigMapGenerator = c
	return k
}

func (k *Kustomization) WithGeneratorOptions(c *types.GeneratorOptions) *Kustomization {
	k.GeneratorOptions = c
	return k
}

func (k *Kustomization) MarshalWithIndent(indent int) ([]byte, error) {
	j, err := json.Marshal(k)
	if err != nil {
		return nil, err
	}

	y, err := JsonToYaml(j, indent)
	if err != nil {

		return nil, err
	}

	return y, nil
}
