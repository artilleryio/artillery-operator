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
	lt "github.com/artilleryio/artillery-operator/api/v1alpha1"
	"github.com/go-logr/logr"
)

func telemeterActive(v *lt.LoadTest, r *LoadTestReconciler, logger logr.Logger) {
	if err := telemetryEnqueue(
		r.TelemetryClient,
		r.TelemetryConfig,
		telemetryEvent{
			Name: "operator test started",
			Properties: map[string]interface{}{
				"name":        hashEncode(v.Name),
				"namespace":   hashEncode(v.Namespace),
				"workers":     v.Spec.Count,
				"environment": len(v.Spec.Environment) > 0,
			},
		},
		logger,
	); err != nil {
		logger.Error(err,
			"could not broadcast telemetry",
			"telemetry disable",
			r.TelemetryConfig.Disable,
			"telemetry debug",
			r.TelemetryConfig.Debug,
			"event",
			"operator load test created",
		)
	}
}

func telemeterCompletion(v *lt.LoadTest, r *LoadTestReconciler, logger logr.Logger) {
	err := telemetryEnqueue(
		r.TelemetryClient,
		r.TelemetryConfig,
		telemetryEvent{
			Name: "operator test completed",
			Properties: map[string]interface{}{
				"name":        hashEncode(v.Name),
				"namespace":   hashEncode(v.Namespace),
				"workers":     v.Spec.Count,
				"environment": len(v.Spec.Environment) > 0,
			},
		},
		logger,
	)
	if err != nil {
		logger.Error(err,
			"could not broadcast telemetry",
			"telemetry disable",
			r.TelemetryConfig.Disable,
			"telemetry debug",
			r.TelemetryConfig.Debug,
			"event",
			"operator load test completed",
		)
	}
}
