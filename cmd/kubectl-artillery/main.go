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

package main

import (
	"os"

	"github.com/artilleryio/artillery-operator/commands"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

func main() {
	wd := "."
	root := commands.NewCmdArtillery(wd, genericclioptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr})
	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}
