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
