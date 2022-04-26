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
	"fmt"
	"time"

	lt "github.com/artilleryio/artillery-operator/api/v1alpha1"
	"github.com/artilleryio/artillery-operator/internal/telemetry"
	"github.com/go-logr/logr"
	"github.com/thoas/go-funk"
	v1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/duration"
	"k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// ObservedStatus defines expected observed statuses for a LoadTest.
type ObservedStatus uint

const (
	LoadTestInactive ObservedStatus = iota
	LoadTestActive
	LoadTestCompleted
)

// updateStatus updates the LoadTestStatus based on the status of created LoadTest objects.
// This includes updates to,
// - Conditions
// - Status properties
// - PrinterColumns
func (r *LoadTestReconciler) updateStatus(ctx context.Context, v *lt.LoadTest, logger logr.Logger) (*reconcile.Result, error) {
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
		if err := configureStatus(ctx, v, r, found, logger); err != nil {
			return err
		}
		return r.Status().Update(ctx, v)
	})
	if err != nil {
		return &ctrl.Result{}, err
	}

	// status updated successfully
	return nil, nil
}

// configureStatus configures, and if need be, broadcasts the LoadTest status
// based on observed status of the created Job object.
func configureStatus(
	ctx context.Context,
	v *lt.LoadTest,
	r *LoadTestReconciler,
	job *v1.Job,
	logger logr.Logger,
) error {
	observedStatus := observedStatus(job.Status)
	configureStatesAndPrinterColumns(v, job)

	setConditions(v, observedStatus)
	if err := broadcastIfActiveOrCompleted(ctx, v, r, observedStatus, logger); err != nil {
		return err
	}
	// Configuration should always happen after any events are published.
	configureStartupAndCompletion(v, observedStatus)

	return nil
}

// observedStatus relays a load test's observed status from its related Job.
func observedStatus(s v1.JobStatus) ObservedStatus {
	switch {
	case jobConditionsActive(s) && (s.Active > (s.Succeeded + s.Failed)):
		return LoadTestActive
	case jobConditionsCompleted(s) && (s.Active == 0 && (s.Succeeded > 0 || s.Failed > 0)):
		return LoadTestCompleted
	default:
		return LoadTestInactive
	}
}

// jobConditionsActive use Job conditions to conclude if a job is active.
func jobConditionsActive(s v1.JobStatus) bool {
	suspended, complete, failed := jobConditions(s)
	running := suspended == false && complete == false && failed == false
	return s.StartTime != nil && running
}

// jobConditionsCompleted use Job conditions to conclude if a job has completed.
func jobConditionsCompleted(s v1.JobStatus) bool {
	suspended, complete, failed := jobConditions(s)
	completed := suspended == false && complete == true && failed == false
	return s.StartTime != nil && s.CompletionTime != nil && completed
}

// jobConditions returns whether a Job is suspended, complete or failed by traversing all found Conditions.
func jobConditions(s v1.JobStatus) (suspended bool, complete bool, failed bool) {
	for _, c := range s.Conditions {
		suspended = c.Type == v1.JobSuspended
		complete = c.Type == v1.JobComplete
		failed = c.Type == v1.JobFailed
	}
	return
}

// setConditions sets the LoadTest's Status Conditions based on
// the provided observed state.
func setConditions(v *lt.LoadTest, o ObservedStatus) {
	var progressing lt.LoadTestCondition
	var completed lt.LoadTestCondition

	conditionsMap := conditionsMap(v.Status.Conditions)
	progressing = conditionsMap[lt.LoadTestProgressing]

	switch o {
	case LoadTestInactive:
		progressing.Status = corev1.ConditionUnknown

	case LoadTestActive:
		progressing.Status = corev1.ConditionTrue

	case LoadTestCompleted:
		progressing.Status = corev1.ConditionFalse

		completed = lt.LoadTestCondition{
			Type:               lt.LoadTestCompleted,
			Status:             corev1.ConditionTrue,
			LastTransitionTime: metav1.Now(),
			LastProbeTime:      metav1.Now(),
		}
		conditionsMap[lt.LoadTestCompleted] = completed
	}

	progressing.LastProbeTime = metav1.Now()
	conditionsMap[lt.LoadTestProgressing] = progressing

	v.Status.Conditions = funk.Map(conditionsMap, func(key lt.LoadTestConditionType, val lt.LoadTestCondition) lt.LoadTestCondition {
		return val
	}).([]lt.LoadTestCondition)
}

// configureStatesAndPrinterColumns configures LoadTestStatus properties
// and printer column values.
// For printercolumns see:
// https://kubernetes.io/docs/tasks/extend-kubernetes/custom-resources/custom-resource-definitions/#additional-printer-columns
func configureStatesAndPrinterColumns(v *lt.LoadTest, job *v1.Job) {
	configureStates(v, job)
	configurePrinterColumns(v, job)
}

func configureStates(v *lt.LoadTest, job *v1.Job) {
	v.Status.Active = job.Status.Active
	v.Status.Succeeded = job.Status.Succeeded
	v.Status.Failed = job.Status.Failed
}

func configurePrinterColumns(v *lt.LoadTest, job *v1.Job) {
	v.Status.Duration = loadTestDuration(v.Status)
	v.Status.Completions = loadTestCompletions(v.Status, job.Spec.Completions, job.Spec.Parallelism)
	v.Status.Image = job.Spec.Template.Spec.Containers[0].Image
}

func conditionsMap(conditions []lt.LoadTestCondition) map[lt.LoadTestConditionType]lt.LoadTestCondition {
	out := funk.ToMap(conditions, "Type").(map[lt.LoadTestConditionType]lt.LoadTestCondition)
	if _, ok := out[lt.LoadTestProgressing]; !ok {
		out[lt.LoadTestProgressing] = lt.LoadTestCondition{
			Type:               lt.LoadTestProgressing,
			LastTransitionTime: metav1.Now(),
		}
	}
	return out
}

// configureStartupAndCompletion configures the start and completion time of a LoadTest
// based on the provided observed state.
func configureStartupAndCompletion(v *lt.LoadTest, o ObservedStatus) {
	switch o {
	case LoadTestActive:
		if v.Status.StartTime == nil {
			now := metav1.Now()
			v.Status.StartTime = &now
		}

	case LoadTestCompleted:
		if v.Status.CompletionTime == nil {
			now := metav1.Now()
			v.Status.CompletionTime = &now
		}
	}
}

// broadcastIfActiveOrCompleted broadcasts informational events to mark that a LoadTest has started or completed.
func broadcastIfActiveOrCompleted(ctx context.Context, v *lt.LoadTest, r *LoadTestReconciler, o ObservedStatus, logger logr.Logger) error {
	switch {
	case o == LoadTestActive && v.Status.StartTime == nil:
		podList, err := getPods(ctx, v, r.Client)
		if err != nil {
			return err
		}
		for _, pod := range podList.Items {
			r.Recorder.Eventf(v, "Normal", "Running", "Running Load Test worker pod: %s", pod.Name)
		}
		telemetry.TelemeterActive(v, r.TelemetryClient, r.TelemetryConfig, logger)

	case o == LoadTestCompleted && v.Status.CompletionTime == nil:
		msg := "Load Test completed"

		if v.Status.Failed > 0 {
			r.Recorder.Event(v, "Warning", "Failed", fmt.Sprintf("%s with failed workers", msg))
		} else {
			r.Recorder.Event(v, "Normal", "Completed", msg)
		}

		telemetry.TelemeterCompletion(v, r.TelemetryClient, r.TelemetryConfig, logger)
	}

	return nil
}

// getPods returns all the worker Pods for a given LoadTest.
func getPods(ctx context.Context, v *lt.LoadTest, ctl client.Client) (*corev1.PodList, error) {
	podList := &corev1.PodList{}
	listOpts := []client.ListOption{
		client.InNamespace(v.Namespace),
		client.MatchingLabels(labels(v, "loadtest-worker")),
	}
	if err := ctl.List(ctx, podList, listOpts...); err != nil {
		return nil, err
	}
	return podList, nil
}

// loadTestDuration returns a formatted LoadTest duration.
func loadTestDuration(status lt.LoadTestStatus) string {
	var d string
	switch {
	case status.StartTime == nil:

	case status.CompletionTime == nil:
		d = duration.HumanDuration(time.Since(status.StartTime.Time))
	default:
		d = duration.HumanDuration(status.CompletionTime.Sub(status.StartTime.Time))
	}
	return d
}

// loadTestCompletions returns a formatted value of how many LoadTest workers
// have completed from the total of all workers, e.g. 1/4.
// It takes worker parallelism into account.
func loadTestCompletions(status lt.LoadTestStatus, jobCompletions, jobParallelism *int32) string {
	if jobCompletions != nil {
		return fmt.Sprintf("%d/%d", status.Succeeded, *jobCompletions)
	}

	parallelism := int32(0)
	if jobParallelism != nil {
		parallelism = *jobParallelism
	}
	if parallelism > 1 {
		return fmt.Sprintf("%d/1 of %d", status.Succeeded, parallelism)
	} else {
		return fmt.Sprintf("%d/1", status.Succeeded)
	}
}
