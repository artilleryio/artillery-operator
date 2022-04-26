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
	"context"

	lt "github.com/artilleryio/artillery-operator/api/v1alpha1"
	"github.com/go-logr/logr"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
)

// ensureTestScriptConfig ensures the test script ConfigMap defined
// in the LoadTest Custom Resource is available on the cluster.
// If not, a Warning event is triggered.
// This event is viewable when running: kubectl describe loadtest <loadtest-name>.
func (r *LoadTestReconciler) ensureTestScriptConfig(ctx context.Context,
	instance *lt.LoadTest,
	logger logr.Logger) (*ctrl.Result, error) {
	configMap := instance.Spec.TestScript.Config.ConfigMap

	found := &core.ConfigMap{}
	err := r.Get(ctx, types.NamespacedName{
		Name:      configMap,
		Namespace: instance.Namespace,
	}, found)

	if err != nil && errors.IsNotFound(err) {
		logger.Error(err, "TestScript ConfigMap is missing", "Testscript.Config.ConfigMap", configMap)
		r.Recorder.Eventf(instance, "Warning", "MissingTestScript", "Load Test test script ConfigMap is missing, see field .spec.testScript.config.configMap")

		return &ctrl.Result{}, err
	} else if err != nil {
		logger.Error(err, "Failed to get ConfigMap", "Testscript.Config.ConfigMap", configMap)
		return &ctrl.Result{}, err
	}

	// ConfigMap located
	return nil, nil
}
