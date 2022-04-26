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

package telemetry

import (
	lt "github.com/artilleryio/artillery-operator/api/v1alpha1"
	"github.com/go-logr/logr"
	"github.com/posthog/posthog-go"
)

// TelemeterActive enqueues a load test has started event.
func TelemeterActive(v *lt.LoadTest, tClient posthog.Client, tConfig Config, logger logr.Logger) {
	if err := enqueue(
		tClient,
		tConfig,
		event{
			Name: "operator test started",
			Properties: map[string]interface{}{
				"name":        hashEncode(v.Name),
				"namespace":   hashEncode(v.Namespace),
				"count":       v.Spec.Count,
				"environment": len(v.Spec.Environment) > 0,
			},
		},
		logger,
	); err != nil {
		logger.Error(err,
			"could not broadcast telemetry",
			"telemetry disable",
			tConfig.Disable,
			"telemetry debug",
			tConfig.Debug,
			"event",
			"operator load test created",
		)
	}
}

// TelemeterCompletion enqueues a load test has completed event.
func TelemeterCompletion(v *lt.LoadTest, tClient posthog.Client, tConfig Config, logger logr.Logger) {
	err := enqueue(
		tClient,
		tConfig,
		event{
			Name: "operator test completed",
			Properties: map[string]interface{}{
				"name":        hashEncode(v.Name),
				"namespace":   hashEncode(v.Namespace),
				"count":       v.Spec.Count,
				"environment": len(v.Spec.Environment) > 0,
			},
		},
		logger,
	)
	if err != nil {
		logger.Error(err,
			"could not broadcast telemetry",
			"telemetry disable",
			tConfig.Disable,
			"telemetry debug",
			tConfig.Debug,
			"event",
			"operator load test completed",
		)
	}
}

// TelemeterGenerateManifests enqueues a kubectl-artillery generate command event.
func TelemeterGenerateManifests(
	name, testScriptPath, env, outPath string,
	count int,
	tClient posthog.Client,
	tConfig Config,
	logger logr.Logger,
) {
	if err := enqueue(
		tClient,
		tConfig,
		event{
			Name: "operator kubectl-artillery generate",
			Properties: map[string]interface{}{
				"source":           "artillery-operator-kubectl-plugin",
				"name":             hashEncode(name),
				"testScript":       hashEncode(testScriptPath),
				"count":            count,
				"environment":      hashEncode(env),
				"defaultOutputDir": len(outPath) == 0,
			},
		},
		logger,
	); err != nil {
		logger.Error(err,
			"could not broadcast telemetry",
			"telemetry disable", tConfig.Disable,
			"telemetry debug", tConfig.Debug,
			"event", "operator kubectl-artillery generate",
		)
	}
}

// TelemeterServicesScaffold enqueues a kubectl-artillery scaffold command event.
func TelemeterServicesScaffold(
	serviceNames []string,
	namespace, outPath string,
	tClient posthog.Client,
	tConfig Config,
	logger logr.Logger,
) {
	if err := enqueue(
		tClient,
		tConfig,
		event{
			Name: "operator kubectl-artillery scaffold",
			Properties: map[string]interface{}{
				"source":           "artillery-operator-kubectl-plugin",
				"serviceCount":     len(serviceNames),
				"namespace":        hashEncode(namespace),
				"defaultOutputDir": len(outPath) == 0,
			},
		},
		logger,
	); err != nil {
		logger.Error(err,
			"could not broadcast telemetry",
			"telemetry disable", tConfig.Disable,
			"telemetry debug", tConfig.Debug,
			"event", "operator kubectl-artillery scaffold",
		)
	}
}
