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

package v1alpha1

import (
	"encoding/json"

	"github.com/artilleryio/artillery-operator/internal/artillery"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewLoadTest(loadTestName, configMapName, env string, count int) *LoadTest {
	kind := "LoadTest"
	apiVersion := "loadtest.artillery.io/v1alpha1"

	meta := metav1.TypeMeta{
		Kind:       kind,
		APIVersion: apiVersion,
	}
	objectMeta := metav1.ObjectMeta{
		Name: loadTestName,
	}
	testScript := TestScript{
		Config: Config{
			ConfigMap: configMapName,
		},
	}
	spec := LoadTestSpec{
		Count:       count,
		Environment: env,
		TestScript:  testScript,
	}

	return &LoadTest{
		TypeMeta:   meta,
		ObjectMeta: objectMeta,
		Spec:       spec,
	}
}

func (lt *LoadTest) MarshalWithIndent(indent int) ([]byte, error) {
	j, err := lt.json()
	if err != nil {
		return nil, err
	}

	y, err := artillery.JsonToYaml(j, indent)
	if err != nil {

		return nil, err
	}

	return y, nil
}

func (lt *LoadTest) json() ([]byte, error) {
	j, err := json.Marshal(lt)
	if err != nil {
		return nil, err
	}

	var temp map[string]interface{}
	if err := json.Unmarshal(j, &temp); err != nil {
		return nil, err
	}
	delete(temp, "status")

	return json.Marshal(temp)
}
