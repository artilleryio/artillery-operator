/*
 * Copyright (c) 2021-2022.
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
	"time"

	lt "github.com/artilleryio/artillery-operator/api/v1alpha1"
	"github.com/artilleryio/artillery-operator/internal/telemetry"
	"github.com/posthog/posthog-go"
	v1 "k8s.io/api/batch/v1"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// LoadTestReconciler reconciles a LoadTest object.
type LoadTestReconciler struct {
	client.Client
	Scheme          *runtime.Scheme
	Recorder        record.EventRecorder
	TelemetryConfig telemetry.Config
	TelemetryClient posthog.Client
}

/*
	This code relies on the kubebuilder library to toggle K8s controller features
	including RBAC.

	For more details, check the controller documentation:
	https://www.kubebuilder.io/cronjob-tutorial/controller-overview.html
*/

// +kubebuilder:rbac:groups=loadtest.artillery.io,resources=loadtests,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=loadtest.artillery.io,resources=loadtests/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=loadtest.artillery.io,resources=loadtests/finalizers,verbs=update
// +kubebuilder:rbac:groups=batch,resources=jobs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=batch,resources=jobs/status,verbs=get
// +kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch;
// +kubebuilder:rbac:groups="",resources=events,verbs=create;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.10.0/pkg/reconcile
func (r *LoadTestReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.WithValues("LoadTest.Name", req.Name, "LoadTest.Namespace", req.Namespace)
	logger.Info("Reconciling LoadTest")

	var (
		result   *ctrl.Result
		loadTest = &lt.LoadTest{}
	)

	err := r.Get(ctx, req.NamespacedName, loadTest)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			logger.Info("LoadTest resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}

		// Error reading the object - requeue the request.
		logger.Error(err, "Failed to get LoadTest")
		return ctrl.Result{}, err
	}

	result, err = r.ensureTestScriptConfig(ctx, loadTest, logger)
	if result != nil {
		return *result, err
	}

	result, err = r.ensureJob(ctx, loadTest, logger, r.job(loadTest))
	if result != nil {
		return *result, err
	}

	result, err = r.updateStatus(ctx, loadTest, logger)
	if result != nil {
		logger.Error(err, "Failed to update LoadTest status")
		return *result, err
	}

	// Track duration for progressing LoadTest
	if loadTest.Status.CompletionTime == nil {
		return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
	}

	// == Finish == == == == ==
	// Everything went fine, don't requeue
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *LoadTestReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&lt.LoadTest{}).
		Owns(&v1.Job{}).
		Owns(&core.Pod{}).
		Complete(r)
}
