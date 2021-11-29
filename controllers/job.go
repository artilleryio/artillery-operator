package controllers

//goland:noinspection SpellCheckingInspection
import (
	"context"
	"time"

	lt "github.com/artilleryio/artillery-operator/api/v1alpha1"
	"github.com/go-logr/logr"
	v1 "k8s.io/api/batch/v1"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"

	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	workerImage   = "artilleryio/artillery:latest"
	testScriptVol = "test-script"
)

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
		},
		Spec: v1.JobSpec{
			Parallelism:  &parallelism,
			Completions:  &completions,
			BackoffLimit: &backoffLimit,
			Template: core.PodTemplateSpec{
				Spec: core.PodSpec{
					Containers: []core.Container{
						{
							Name:  v.Name,
							Image: workerImage,
							VolumeMounts: []core.VolumeMount{
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
					Volumes: []core.Volume{
						{
							Name: testScriptVol,
							VolumeSource: core.VolumeSource{
								ConfigMap: &core.ConfigMapVolumeSource{
									LocalObjectReference: core.LocalObjectReference{
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

func (r *LoadTestReconciler) updateJobStatus(ctx context.Context, v *lt.LoadTest) (*reconcile.Result, error) {
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
		v.Status.Active = found.Status.Active > 0
		return r.Status().Update(ctx, v)
	})
	if err != nil {
		return &ctrl.Result{}, err
	}

	// status updated successfully
	return nil, nil
}
