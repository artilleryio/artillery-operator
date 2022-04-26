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
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
)

// Client defines the K8s client.
type Client struct {
	CfgNamespace string
	*kubernetes.Clientset
}

// NewClient returns a K8s client configured with ConfigFlags.
// ConfigFlags compose the set of values necessary for obtaining K8s REST client config.
func NewClient(configFlags *genericclioptions.ConfigFlags) (*Client, error) {
	config := configFlags.ToRawKubeConfigLoader()

	cfgNamespace, _, err := config.Namespace()
	if err != nil {
		return nil, err
	}

	clientConfig, err := config.ClientConfig()
	if err != nil {
		return nil, err
	}

	ctl, err := kubernetes.NewForConfig(clientConfig)
	if err != nil {
		return nil, err
	}

	return &Client{
		CfgNamespace: cfgNamespace,
		Clientset:    ctl,
	}, nil
}
