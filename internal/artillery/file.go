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

package artillery

import (
	"io/ioutil"
	"os"
	"path/filepath"
)

func DirOrFileExists(path string) bool {
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		return true
	}
	return false
}

func GetOrCreateTargetDir(workingDir string, outPath string) (string, error) {
	result := outPath

	if len(result) == 0 {
		result = filepath.Join(workingDir, "artillery-manifests")
	}

	return result, os.MkdirAll(result, 0700)
}

func CopyFileTo(dir string, srcPath string) error {
	input, err := ioutil.ReadFile(srcPath)
	if err != nil {
		return err
	}

	dest := filepath.Join(dir, filepath.Base(srcPath))
	err = ioutil.WriteFile(dest, input, 0644)
	if err != nil {
		return err
	}
	return nil
}
