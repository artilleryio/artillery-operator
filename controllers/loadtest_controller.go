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

	lt "github.com/artilleryio/artillery-operator/api/v1alpha1"
	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// LoadTestReconciler reconciles a LoadTest object
type LoadTestReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=loadtest.artillery.io,resources=loadtests,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=loadtest.artillery.io,resources=loadtests/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=loadtest.artillery.io,resources=loadtests/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the LoadTest object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.10.0/pkg/reconcile
func (r *LoadTestReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.WithValues("LoadTest.Name", req.Name, "LoadTest.Namespace", req.Namespace)
	logger.Info("Reconciling LoadTest")

	loadTest := &lt.LoadTest{}
	err := r.Get(ctx, req.NamespacedName, loadTest)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not foundJob, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			logger.Info("LoadTest resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		logger.Error(err, "Failed to get LoadTest")
		return ctrl.Result{}, err
	}

	foundJob := &batchv1.Job{}
	err = r.Get(ctx, types.NamespacedName{Name: loadTest.Name, Namespace: loadTest.Namespace}, foundJob)
	if err != nil && errors.IsNotFound(err) {
		// Define a new job
		job, err := r.jobForLoadTest(loadTest)
		if err != nil {
			logger.Error(err, "Failed to define and attach Job", "Job.Namespace", job.Namespace, "Job.Name", job.Name)
			return ctrl.Result{}, err
		}
		logger.Info("Creating a new Job", "Job.Namespace", job.Namespace, "Job.Name", job.Name)
		err = r.Create(ctx, job)
		if err != nil {
			logger.Error(err, "Failed to create new Job", "Job.Namespace", job.Namespace, "Job.Name", job.Name)
			return ctrl.Result{}, err
		}

		// Job created successfully - return and requeue
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		// Error reading the object - requeue the request.
		logger.Error(err, "Failed to get Job", "Job.Namespace", foundJob.Namespace, "Job.Name", foundJob.Name)
		return ctrl.Result{}, err
	}

	err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
		loadTest.Status.Active = foundJob.Status.Active > 0
		return r.Status().Update(ctx, loadTest)
	})

	if err != nil {
		logger.Error(err, "Failed to update LoadTest status")
		return ctrl.Result{}, err
	}

	logger.Info("LoadTest Reconciled")

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *LoadTestReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&lt.LoadTest{}).
		Owns(&batchv1.Job{}).
		Complete(r)
}

func (r *LoadTestReconciler) jobForLoadTest(loadTest *lt.LoadTest) (*batchv1.Job, error) {
	var (
		parallelism    int32 = 1
		completions    int32 = 1
		backoffLimit   int32 = 0
		containerImage       = "artilleryio/artillery:latest"
		testScriptVol        = "test-script"
	)

	if loadTest.Spec.Count > 0 {
		parallelism = int32(loadTest.Spec.Count)
		completions = int32(loadTest.Spec.Count)
	}

	var out = &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      loadTest.Name,
			Namespace: loadTest.Namespace,
		},
		Spec: batchv1.JobSpec{
			Parallelism:  &parallelism,
			Completions:  &completions,
			BackoffLimit: &backoffLimit,
			Template: v1.PodTemplateSpec{
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:  loadTest.Name,
							Image: containerImage,
							VolumeMounts: []v1.VolumeMount{
								{
									Name:      testScriptVol,
									MountPath: "/data",
								},
							},
							Args: []string{
								"run",
								"/data/test-script.yaml",
							},
						},
					},
					Volumes: []v1.Volume{
						{
							Name: testScriptVol,
							VolumeSource: v1.VolumeSource{
								ConfigMap: &v1.ConfigMapVolumeSource{
									LocalObjectReference: v1.LocalObjectReference{
										Name: loadTest.Spec.TestScript.Config.ConfigMap,
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

	err := ctrl.SetControllerReference(loadTest, out, r.Scheme)
	if err != nil {
		return nil, err
	}

	return out, nil
}
