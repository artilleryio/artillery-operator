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
	"crypto/sha1"
	"encoding/base64"
	"os"
	"runtime"
	"strconv"
	"strings"

	"github.com/go-logr/logr"
	"github.com/panta/machineid"
	"github.com/posthog/posthog-go"
	core "k8s.io/api/core/v1"
)

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
//  - test created
// 	extra props:
// 	  - name: hashed test name
// 	  - namespace: hashed namespace name
// 		- workers: 1
// 		- environment: true/false
//  - test completed
// 	extra props:
// 	  - name: hashed test name
//    - namespace: hashed namespace
// 		- workers: 1

type telemetryEvent struct {
	Name       string
	Properties map[string]interface{}
}

func protectedDistinctId() (string, error) {
	return machineid.ProtectedID(AppName)
}

func hashEncode(v string) string {
	h := sha1.New()
	h.Write([]byte(v))
	hashed := h.Sum(nil)
	return base64.StdEncoding.EncodeToString(hashed)
}

type noopClient struct{}

func (n *noopClient) Close() error                                              { return nil }
func (n *noopClient) Enqueue(_ posthog.Message) error                           { return nil }
func (n *noopClient) IsFeatureEnabled(_ string, _ string, _ bool) (bool, error) { return true, nil }
func (n *noopClient) ReloadFeatureFlags() error                                 { return nil }
func (n *noopClient) GetFeatureFlags() ([]posthog.FeatureFlag, error)           { return nil, nil }

func NewTelemetryClient(tCfg TelemetryConfig) (posthog.Client, error) {
	if tCfg.Disable {
		return &noopClient{}, nil
	}
	return posthog.NewWithConfig(PosthogToken, posthog.Config{})
}

func telemetryEnqueue(client posthog.Client, config TelemetryConfig, event telemetryEvent, logger logr.Logger) error {
	properties := buildProperties(event.Properties)
	if config.Debug {
		debugTelemetryProperties(properties, logger)
		return nil
	}

	distinctId, err := protectedDistinctId()
	if err != nil {
		return err
	}

	if err := client.Enqueue(posthog.Capture{
		DistinctId: distinctId,
		Event:      event.Name,
		Properties: properties,
	}); err != nil {
		return err
	}

	return nil
}

func debugTelemetryProperties(props map[string]interface{}, logger logr.Logger) {
	for k, v := range props {
		logger.Info("ARTILLERY_TELEMETRY_DEBUG=true", k, v)
	}
}

func buildProperties(extra map[string]interface{}) map[string]interface{} {
	properties := map[string]interface{}{
		"source":      AppName,
		"version":     Version,
		"containerOS": strings.ToLower(runtime.GOOS),
		"workerImage": WorkerImage,
	}

	for key, val := range extra {
		properties[key] = val
	}

	return properties
}

type TelemetryConfig struct {
	Disable bool
	Debug   bool
}

func NewTelemetryConfig(logger logr.Logger) TelemetryConfig {
	result := TelemetryConfig{}

	if getTelemetryDisableConfig(logger) {
		result.Disable = true
	}

	if getTelemetryDebugConfig(logger) {
		result.Debug = true
	}

	return result
}

func (t TelemetryConfig) toEnvVar() []core.EnvVar {
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
