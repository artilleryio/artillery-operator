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

package kube

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DoQuery queries a K8s cluster for K8s services using specified service names and namespace.
// It returns a list of query results, one for each found and missed service name.
func DoQuery(ctx context.Context, svcNames []string, ns string, ctl *Client) (QueryResults, error) {
	var result QueryResults

	for _, svcName := range svcNames {
		qr := QueryResult{serviceName: svcName, selection: Selection{}}

		service, err := ctl.CoreV1().Services(ns).Get(ctx, svcName, metav1.GetOptions{})
		if err != nil {
			result = append(result, qr)
			continue
		}

		if strings.ToLower(service.Name) == strings.ToLower(svcName) {
			pods, err := ctl.CoreV1().Pods(ns).List(ctx, metav1.ListOptions{
				LabelSelector: selectorLabels(service),
			})
			if err != nil {
				return nil, err
			}

			if len(pods.Items) > 0 {
				qr.hit = true
				qr.selection = Selection{Service: *service, Pod: pods.Items[0]}
			}
		}

		result = append(result, qr)
	}
	return result, nil
}

// QueryResults defines a list query results.
type QueryResults []QueryResult

// HasQueryHits returns whether there are any query results found matching any K8s Services.
func (r QueryResults) HasQueryHits() bool {
	return len(r) > len(r.QueryMisses())
}

// QueryMisses returns any K8s Services that could not be found.
func (r QueryResults) QueryMisses() QueryResults {
	var out QueryResults
	for _, qr := range r {
		if !qr.QueryHit() {
			out = append(out, qr)
		}
	}
	return out
}

// HasLivenessHits returns whether any K8s Services can expose any HTTP Get liveness probes.
func (r QueryResults) HasLivenessHits() bool {
	return len(r.LivenessHits()) > 0
}

// LivenessMisses return any K8s Services do exist BUT CANNOT expose any HTTP Get liveness probes.
func (r QueryResults) LivenessMisses() QueryResults {
	var out QueryResults
	for _, queryResult := range r {
		if queryResult.QueryHit() && !queryResult.LivenessHit() {
			out = append(out, queryResult)
		}
	}
	return out
}

// LivenessHits returns any K8s Services that DO EXIST AND can expose any HTTP Get liveness probes.
func (r QueryResults) LivenessHits() QueryResults {
	var out QueryResults
	for _, queryResult := range r {
		if queryResult.QueryHit() && queryResult.LivenessHit() {
			out = append(out, queryResult)
		}
	}
	return out
}

// QueryResult defines the result of a K8s Services query.
type QueryResult struct {
	serviceName string
	hit         bool
	selection   Selection
}

// QueryHit returns whether a query result found a K8s Service.
func (qr QueryResult) QueryHit() bool {
	return qr.hit
}

// LivenessHit returns whether a query result found a K8s Service that can expose HTTP Get liveness Probes.
func (qr QueryResult) LivenessHit() bool {
	return len(qr.ServiceProbes()) > 0
}

// ServiceProbes returns a K8s Service's exposed HTTP Get liveness Probes for a query result
// based on a service + pod selection.
func (qr QueryResult) ServiceProbes() ServiceProbes {
	return qr.selection.ServiceProbes()
}

// QueriedServiceName returns a query result's queried service name.
func (qr QueryResult) QueriedServiceName() string {
	return qr.serviceName
}

// SelectionServiceName returns a K8s Service name for a service + pod selection.
func (qr QueryResult) SelectionServiceName() string {
	return qr.selection.serviceName()
}

type ServiceProbes []ServiceProbe

// ServiceProbe is a list of HTTP Get liveness probes for a K8s Service.
type ServiceProbe struct {
	Url      *url.URL
	HTTPGets []*corev1.HTTPGetAction
}

// Selection a selection pairs a K8s Service and a Pod based on a Service's selector labels.
type Selection struct {
	Service corev1.Service
	Pod     corev1.Pod
}

// serviceName the name of the K8s Service in service + pod selection.
func (s Selection) serviceName() string {
	return s.Service.Name
}

// ServiceProbes returns Pod HTTP Get liveness probes that a Service can expose
// using one of it's configured ports.
func (s Selection) ServiceProbes() ServiceProbes {
	var out ServiceProbes

	svcHasNoSelector := s.Service.Spec.Selector == nil || len(s.Service.Spec.Selector) == 0
	if svcHasNoSelector {
		return out
	}

	svcExternal := s.Service.Spec.Type == "ExternalName"
	if svcExternal {
		return out
	}

	for _, servicePort := range s.Service.Spec.Ports {
		var livenessCollector []*corev1.HTTPGetAction
		svcTargetPort := servicePort.TargetPort.IntVal

		for _, cntnr := range s.Pod.Spec.Containers {
			if cntnr.LivenessProbe == nil {
				continue
			}

			httpGet := cntnr.LivenessProbe.HTTPGet
			if httpGet == nil || svcTargetPort != httpGet.Port.IntVal {
				continue
			}

			livenessCollector = append(livenessCollector, httpGet)
		}

		if len(livenessCollector) > 0 {
			probe := ServiceProbe{
				Url: &url.URL{
					Scheme: "http",
					Host:   fmt.Sprintf("%s:%d", s.serviceName(), servicePort.Port),
				},
				HTTPGets: livenessCollector,
			}
			out = append(out, probe)
		}
	}

	return out
}

// selectorLabels returns a K8s Service's selector labels.
// selector labels provide labels to identify a service's downstream Pods.
func selectorLabels(svc *corev1.Service) string {
	var labels string
	for k, v := range svc.Spec.Selector {
		labels = fmt.Sprintf("%s=%s,", k, v)
	}
	return strings.TrimRight(labels, ",")
}
