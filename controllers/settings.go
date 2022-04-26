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

const (
	// AppName the controller and CLI app name.
	AppName = "artillery-operator"

	// Version controller version.
	Version = "alpha"

	// WorkerImage the Artillery image used by workers to run load tests.
	WorkerImage = "artilleryio/artillery:latest"

	// TestScriptVol the volume used by created LoadTest Pods to load the test script ConfigMap.
	TestScriptVol = "test-script"

	// TestScriptFilename expected filename used by the test script ConfigMap.
	TestScriptFilename = "test-script.yaml"
)
