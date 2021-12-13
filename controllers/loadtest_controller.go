/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"time"

	lt "github.com/artilleryio/artillery-operator/api/v1alpha1"
	v1 "k8s.io/api/batch/v1"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// LoadTestReconciler reconciles a LoadTest object
type LoadTestReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

// +kubebuilder:rbac:groups=loadtest.artillery.io,resources=loadtests,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=loadtest.artillery.io,resources=loadtests/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=loadtest.artillery.io,resources=loadtests/finalizers,verbs=update
// +kubebuilder:rbac:groups=batch,resources=jobs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=batch,resources=jobs/status,verbs=get
// +kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch;

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

	result, err = r.ensureJob(ctx, loadTest, logger, r.job(loadTest))
	if result != nil {
		return *result, err
	}

	result, err = r.updateStatus(ctx, loadTest)
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
