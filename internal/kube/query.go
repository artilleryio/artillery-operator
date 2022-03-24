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

type QueryResults []QueryResult

func (r QueryResults) HasQueryHits() bool {
	return len(r) > len(r.QueryMisses())
}

func (r QueryResults) QueryMisses() QueryResults {
	var out QueryResults
	for _, qr := range r {
		if !qr.QueryHit() {
			out = append(out, qr)
		}
	}
	return out
}

func (r QueryResults) HasLivenessHits() bool {
	return len(r.LivenessHits()) > 0
}

func (r QueryResults) LivenessMisses() QueryResults {
	var out QueryResults
	for _, queryResult := range r {
		if queryResult.QueryHit() && !queryResult.LivenessHit() {
			out = append(out, queryResult)
		}
	}
	return out
}

func (r QueryResults) LivenessHits() QueryResults {
	var out QueryResults
	for _, queryResult := range r {
		if queryResult.QueryHit() && queryResult.LivenessHit() {
			out = append(out, queryResult)
		}
	}
	return out
}

type QueryResult struct {
	serviceName string
	hit         bool
	selection   Selection
}

func (qr QueryResult) QueryHit() bool {
	return qr.hit
}

func (qr QueryResult) LivenessHit() bool {
	return len(qr.ServiceProbes()) > 0
}

func (qr QueryResult) ServiceProbes() ServiceProbes {
	return qr.selection.ServiceProbes()
}

func (qr QueryResult) QueriedServiceName() string {
	return qr.serviceName
}

func (qr QueryResult) SelectionServiceName() string {
	return qr.selection.serviceName()
}

type ServiceProbes []ServiceProbe

type ServiceProbe struct {
	Url      *url.URL
	HTTPGets []*corev1.HTTPGetAction
}

type Selection struct {
	Service corev1.Service
	Pod     corev1.Pod
}

func (s Selection) serviceName() string {
	return s.Service.Name
}

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

func selectorLabels(svc *corev1.Service) string {
	var labels string
	for k, v := range svc.Spec.Selector {
		labels = fmt.Sprintf("%s=%s,", k, v)
	}
	return strings.TrimRight(labels, ",")
}
