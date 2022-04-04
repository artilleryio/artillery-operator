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

	"github.com/artilleryio/artillery-operator/internal/kube"
	yaml3 "gopkg.in/yaml.v3"
)

type TestScript struct {
	Config    Config     `json:"config" yaml:"config"`
	Scenarios []Scenario `json:"scenarios" yaml:"scenarios"`
}

type Config struct {
	Target       string                 `json:"target" yaml:"target"`
	Phases       []Phase                `json:"phases,omitempty" yaml:"phases,omitempty"`
	Environments map[string]Environment `json:"environments,omitempty" yaml:"environments,omitempty"`
}

type Phase struct {
	Duration     int `json:"duration,omitempty" yaml:"duration,omitempty"`
	ArrivalCount int `json:"arrivalCount,omitempty" yaml:"arrivalCount,omitempty"`
	ArrivalRate  int `json:"arrivalRate,omitempty" yaml:"arrivalRate,omitempty"`
}

type Environment struct {
	Phases  []Phase                `json:"phases" yaml:"phases"`
	Target  string                 `json:"target,omitempty" yaml:"target,omitempty"`
	Plugins map[string]interface{} `json:"plugins" yaml:"plugins"`
}

type Scenario struct {
	Flows []Flow `json:"flow,omitempty" yaml:"flow,omitempty"`
}

type Flow struct {
	GetFlow GetFlow `json:"get,omitempty" yaml:"get,omitempty"`
}

type GetFlow struct {
	Url    string       `json:"url,omitempty" yaml:"url,omitempty"`
	Expect []StatusCode `json:"expect,omitempty" yaml:"expect,omitempty"`
}

type StatusCode struct {
	Code int `json:"statusCode,omitempty" yaml:"statusCode,omitempty"`
}

func NewTestScript(probes kube.ServiceProbes) *TestScript {
	var flows []Flow
	for _, probe := range probes {
		target := probe.Url
		for _, get := range probe.HTTPGets {
			target.Path = get.Path
			flow := Flow{
				GetFlow: GetFlow{
					Url: fmt.Sprintf("%s", target.String()),
					Expect: []StatusCode{
						{
							Code: 200,
						},
					},
				},
			}
			flows = append(flows, flow)
		}
	}

	return &TestScript{
		Config: Config{
			Target: probes[0].Url.String(),

			Environments: map[string]Environment{
				"functional": {
					Phases: []Phase{
						{
							Duration:     1,
							ArrivalCount: 1,
						},
					},
					Plugins: map[string]interface{}{
						"expect": make(map[string]string),
					},
				},
			},
		},
		Scenarios: []Scenario{
			{
				Flows: flows,
			},
		},
	}
}

func (t *TestScript) MarshalWithIndent(indent int) ([]byte, error) {
	var out bytes.Buffer
	encoder := yaml3.NewEncoder(&out)
	defer func(encoder *yaml3.Encoder) {
		err := encoder.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(encoder)
	encoder.SetIndent(indent)

	if err := encoder.Encode(t); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}
