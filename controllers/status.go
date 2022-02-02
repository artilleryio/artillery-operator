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
	"github.com/go-logr/logr"
	"github.com/thoas/go-funk"
	v1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/duration"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type ObservedStatus uint

const (
	LoadTestInactive ObservedStatus = iota
	LoadTestActive
	LoadTestCompleted
)

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
		if err := setStatus(ctx, v, r, found, logger); err != nil {
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

func setStatus(ctx context.Context, v *lt.LoadTest, r *LoadTestReconciler, job *v1.Job, logger logr.Logger) error {
	observedStatus := observedStatus(job.Status)

	setConditions(v, observedStatus)
	if err := publishEventsIfAny(ctx, v, r.Client, r.Recorder, observedStatus); err != nil {
		return err
	}
	// Configuration should always happen after any events are published
	configureStartupAndCompletion(v, observedStatus)
	configureStatusAttrs(v, job)
	enqueueCompletionIfDone(v, r, observedStatus, logger)

	return nil
}

func enqueueCompletionIfDone(v *lt.LoadTest, r *LoadTestReconciler, o ObservedStatus, logger logr.Logger) {
	switch o {
	case LoadTestCompleted:
		err := telemetryEnqueue(
			r.TelemetryClient,
			r.TelemetryConfig,
			telemetryEvent{
				Name: "operator load test completed",
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
}

func observedStatus(s v1.JobStatus) ObservedStatus {
	switch {
	case s.Active > (s.Succeeded + s.Failed):
		return LoadTestActive
	case s.Active == 0 && (s.Succeeded > 0 || s.Failed > 0):
		return LoadTestCompleted
	default:
		return LoadTestInactive
	}
}

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
		if v.Status.Failed > 0 {
			completed.Status = corev1.ConditionFalse
		}
		conditionsMap[lt.LoadTestCompleted] = completed
	}

	progressing.LastProbeTime = metav1.Now()
	conditionsMap[lt.LoadTestProgressing] = progressing

	v.Status.Conditions = funk.Map(conditionsMap, func(key lt.LoadTestConditionType, val lt.LoadTestCondition) lt.LoadTestCondition {
		return val
	}).([]lt.LoadTestCondition)
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

		if v.Status.Failed > 0 {
			v.Status.CompletionTime = nil
		}
	}
}

func publishEventsIfAny(ctx context.Context, v *lt.LoadTest, ctl client.Client, r record.EventRecorder, o ObservedStatus) error {
	switch {
	case o == LoadTestActive && v.Status.StartTime == nil:
		podList, err := getPods(ctx, v, ctl)
		if err != nil {
			return err
		}
		for _, pod := range podList.Items {
			r.Eventf(v, "Normal", "Running", "Running Load Test worker pod: %s", pod.Name)
		}

	case o == LoadTestCompleted && v.Status.CompletionTime == nil:
		r.Event(v, "Normal", "Completed", "Load Test Completed")
	}

	return nil
}

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

func configureStatusAttrs(v *lt.LoadTest, job *v1.Job) {
	v.Status.Active = job.Status.Active
	v.Status.Succeeded = job.Status.Succeeded
	v.Status.Failed = job.Status.Failed
	v.Status.Duration = loadTestDuration(v.Status)
	v.Status.Completions = loadTestCompletions(v.Status, job.Spec.Completions, job.Spec.Parallelism)
	v.Status.Image = job.Spec.Template.Spec.Containers[0].Image
}

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
