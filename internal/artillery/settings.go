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
	"io/ioutil"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const LoadTestFilename = "loadtest-cr.yaml"
const LabelPrefix = "loadtest"
const DefaultManifestDir = "artillery-manifests"
const DefaultScriptsDir = "artillery-scripts"
const cliSettingsFilename = ".artillerykuberc"

// CLISettings defines global settings used by the kubectl-artillery CLI.
type CLISettings struct {
	File      string                `yaml:"-" json:"-"`
	Analytics *TelemetryCLISettings `yaml:"kubectl-artillery,omitempty" json:"kubectl-artillery,omitempty"`
}

// TelemetryCLISettings telemetry specific settings.
type TelemetryCLISettings struct {
	FirstRun *bool `yaml:"telemetry-first-run-msg,omitempty" json:"telemetry-first-run-msg,omitempty"`
}

// GetOrCreateCLISettings retrieves the kubectl-artillery CLI settings.
// Creates a new settings file if one doesn't already exist.
func GetOrCreateCLISettings() (*CLISettings, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	settingsPath := filepath.Join(homeDir, cliSettingsFilename)
	if !DirOrFileExists(settingsPath) {
		on := true
		data, err := yaml.Marshal(&CLISettings{
			Analytics: &TelemetryCLISettings{
				FirstRun: &on,
			},
		})
		if err != nil {
			return nil, err
		}

		if err := ioutil.WriteFile(settingsPath, data, 0666); err != nil {
			return nil, err
		}
	}

	var out CLISettings
	data, err := ioutil.ReadFile(settingsPath)
	if err != nil {
		return nil, err
	}
	if err := yaml.Unmarshal(data, &out); err != nil {
		return nil, err
	}
	out.File = settingsPath

	return &out, nil
}

// GetFirstRun returns if this is the first run of kubectl-artillery CLI.
func (s *CLISettings) GetFirstRun() bool {
	return *s.Analytics.FirstRun
}

// SetFirstRun configures the first run of kubectl-artillery CLI.
func (s *CLISettings) SetFirstRun(b bool) *CLISettings {
	s.Analytics.FirstRun = &b
	return s
}

// Save writes the kubectl-artillery CLI settings to a file.
func (s *CLISettings) Save() error {
	data, err := yaml.Marshal(s)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(s.File, data, 0666)
}
