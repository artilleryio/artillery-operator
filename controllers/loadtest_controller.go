/*
 * Copyright (c) 2021.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0.
 *
 * If a copy of the MPL was not distributed with
 * this file, You can obtain one at
 *
 *     http://mozilla.org/MPL/2.0/
 */

package controllers

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"time"

	lt "github.com/artilleryio/artillery-operator/api/v1alpha1"
	"github.com/go-logr/logr"

	v1 "k8s.io/api/batch/v1"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/retry"
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

	result, err = r.publishMetrics(ctx, loadTest, logger)
	if result != nil {
		logger.Error(err, "Failed to publish metrics")
		return *result, err
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

func (r *LoadTestReconciler) publishMetrics(ctx context.Context, v *lt.LoadTest, logger logr.Logger) (*ctrl.Result, error) {
	found := &v1.Job{}
	err := r.Get(ctx, types.NamespacedName{
		Name:      v.Name,
		Namespace: v.Namespace,
	}, found)
	if err != nil {
		// The job may not have been created yet, so requeue
		return &ctrl.Result{RequeueAfter: 5 * time.Second}, err
	}

	err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
		if observedStatus(found.Status) != LoadTestCompleted {
			return nil
		}

		restConfig, err := ctrl.GetConfig()
		if err != nil {
			return err
		}

		clientset, err := kubernetes.NewForConfig(restConfig)
		if err != nil {
			return err
		}

		podList, err := getPods(ctx, v, r.Client)
		if err != nil {
			return err
		}

		tailLines := int64(23) // no of lines for summary
		logger.Info("Pod list items", "podList.Items.Len", len(podList.Items))
		for _, pod := range podList.Items {
			req := clientset.CoreV1().Pods(pod.Namespace).GetLogs(pod.Name, &core.PodLogOptions{TailLines: &tailLines})
			podLogs, err := req.Stream(ctx)
			if err != nil {
				return fmt.Errorf("Error in opening stream, err:\n%s", err.Error())
			}

			buf := new(bytes.Buffer)
			if _, err := io.Copy(buf, podLogs); err != nil {
				return fmt.Errorf("Error in copy information from podLogs to buf, err:\n%s", err.Error())
			}

			logger.Info(fmt.Sprintf("Logs for Pod: [%s]\n%s\n-----------------\n", pod.Name, buf.String()))

			if err := podLogs.Close(); err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return &ctrl.Result{}, err
	}

	// status updated successfully
	return nil, nil
}
