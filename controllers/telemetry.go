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

package controllers

import (
	"os"
	"strconv"

	"github.com/go-logr/logr"
	core "k8s.io/api/core/v1"
)

// const POSTHOG_TOKEN = '_uzX-_WJoVmE_tsLvu0OFD2tpd0HGz72D5sU1zM2hbs';

// TO COLLECT:

// Artillery version
// OS identifier
// whether artillery runs in CI or not
//
// General data:
// - DistinctId -> operator instance running in docker use https://github.com/panta/machineid
// - Source: "artillery-operator"
// - version: "alpha"
// - ContainerOS: os....
// - WorkersImage: workerImage("ghcr.io/artilleryio/artillery-metrics-enabled:experimental")
//
// Events
//  - test run
// 	extra props:
// 		- worker-count: 1
// 		- environment: true/false

type telemetryConfig struct {
	Disable bool
	Debug   bool
}

func newTelemetryConfig(logger logr.Logger) telemetryConfig {
	result := telemetryConfig{}

	if getTelemetryDisableConfig(logger) {
		result.Disable = true
	}

	if getTelemetryDebugConfig(logger) {
		result.Debug = true
	}

	return result
}

func getTelemetryDisableConfig(logger logr.Logger) bool {
	disable, ok := os.LookupEnv("ARTILLERY_DISABLE_TELEMETRY")
	if !ok {
		logger.Info("ARTILLERY_DISABLE_TELEMETRY was not set!")
	}

	parsedDisable, err := strconv.ParseBool(disable)
	if err != nil {
		logger.Info("ARTILLERY_DISABLE_TELEMETRY was not set with boolean type value. TELEMETRY REMAINS ENABLED")
	}
	return parsedDisable
}

func getTelemetryDebugConfig(logger logr.Logger) bool {
	debug, ok := os.LookupEnv("ARTILLERY_TELEMETRY_DEBUG")
	if !ok {
		logger.Info("ARTILLERY_TELEMETRY_DEBUG was not set!")
	}

	parsedDebug, err := strconv.ParseBool(debug)
	if err != nil {
		logger.Info("ARTILLERY_TELEMETRY_DEBUG was not set with boolean type value. TELEMETRY DEBUG REMAINS DISABLED")
	}
	return parsedDebug
}

func (t telemetryConfig) toEnvVar() []core.EnvVar {
	return []core.EnvVar{
		{
			Name:  "ARTILLERY_DISABLE_TELEMETRY",
			Value: strconv.FormatBool(t.Disable),
		},
		{
			Name:  "ARTILLERY_TELEMETRY_DEBUG",
			Value: strconv.FormatBool(t.Debug),
		},
	}
}
