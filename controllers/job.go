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

//goland:noinspection SpellCheckingInspection
import (
	"context"

	lt "github.com/artilleryio/artillery-operator/api/v1alpha1"
	"github.com/go-logr/logr"
	v1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// ensureJob creates a Job that in turn creates the required worker Pods
// to run load tests using an Artillery image.
func (r *LoadTestReconciler) ensureJob(
	ctx context.Context,
	instance *lt.LoadTest,
	logger logr.Logger,
	job *v1.Job,
) (*reconcile.Result, error) {

	found := &v1.Job{}
	err := r.Get(ctx, types.NamespacedName{
		Name:      job.Name,
		Namespace: instance.Namespace,
	}, found)

	if err != nil && errors.IsNotFound(err) {
		// Create a new job
		logger.Info("Creating a new Job", "Job.Namespace", job.Namespace, "Job.Name", job.Name)

		err = r.Create(ctx, job)
		if err != nil {
			logger.Error(err, "Failed to create new Job", "Job.Namespace", job.Namespace, "Job.Name", job.Name)
			return &ctrl.Result{}, err
		}

		r.Recorder.Eventf(instance, "Normal", "Created", "Created Load Test worker master job: %s", job.Name)

		// job created successfully
		return nil, nil
	} else if err != nil {
		// Error reading the object - requeue the request.
		logger.Error(err, "Failed to get Job", "Job.Namespace", found.Namespace, "Job.Name", found.Name)
		return &ctrl.Result{}, err
	}

	// job found successfully
	return nil, nil
}

// job creates a Job spec based on the LoadTest Custom Resource.
func (r *LoadTestReconciler) job(v *lt.LoadTest) *v1.Job {
	var (
		parallelism  int32 = 1
		completions  int32 = 1
		backoffLimit int32 = 0
	)

	if v.Spec.Count > 0 {
		parallelism = int32(v.Spec.Count)
		completions = int32(v.Spec.Count)
	}

	job := &v1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      v.Name,
			Namespace: v.Namespace,
			Labels:    labels(v, "loadtest-worker-master"),
		},
		Spec: v1.JobSpec{
			Parallelism:  &parallelism,
			Completions:  &completions,
			BackoffLimit: &backoffLimit,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels(v, "loadtest-worker"),
				},

				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:            v.Name,
							Image:           WorkerImage,
							ImagePullPolicy: corev1.PullAlways,
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      TestScriptVol,
									MountPath: "/data",
								},
							},
							Args: []string{
								"run",
								"/data/" + TestScriptFilename,
							},
							Env: append(
								[]corev1.EnvVar{
									// published metrics use WORKER_ID to connect the pod (worker) to a Pushgateway JobID
									// Uses the downward API:
									// https://kubernetes.io/docs/tasks/inject-data-application/downward-api-volume-expose-pod-information/#the-downward-api
									{
										Name: "WORKER_ID",
										ValueFrom: &corev1.EnvVarSource{
											FieldRef: &corev1.ObjectFieldSelector{
												FieldPath: "metadata.name",
											},
										},
									},
								},
								r.TelemetryConfig.ToK8sEnvVar()...,
							),
						},
					},
					// Provides access to the ConfigMap holding the test script config
					Volumes: []corev1.Volume{
						{
							Name: TestScriptVol,
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: v.Spec.TestScript.Config.ConfigMap,
									},
								},
							},
						},
					},
					RestartPolicy: "Never",
				},
			},
		},
	}

	_ = ctrl.SetControllerReference(v, job, r.Scheme)
	return job
}

// labels creates K8s labels used to organize
// and categorize (scope and select) Load Test objects.
func labels(v *lt.LoadTest, component string) map[string]string {
	return map[string]string{
		"artillery.io/test-name": v.Name,
		"artillery.io/component": component,
		"artillery.io/part-of":   "loadtest",
	}
}
