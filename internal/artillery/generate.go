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
	"fmt"
	"os"
	"path/filepath"
)

type Generatable struct {
	Path      string
	Marshaler FileMarshaler
}

type Generatables []Generatable

func (g Generatable) generate(indent int) (n int64, err error) {
	data, err := g.Marshaler.MarshalWithIndent(indent)
	if err != nil {
		return int64(0), err
	}

	absPath, err := filepath.Abs(g.Path)
	if err != nil {
		return int64(0), err
	}

	file, err := os.Create(absPath)
	if err != nil {
		return int64(0), err
	}

	written, err := file.Write(data)
	if err != nil {
		return int64(0), err
	}

	return int64(written), file.Close()
}

func (gs Generatables) Generate(indent int) (string, error) {
	var msg string
	for i, g := range gs {
		mWritten, err := g.generate(indent)
		if err != nil {
			return "", err
		}

		if mWritten > 0 && i == 0 {
			msg = fmt.Sprintf("%s generated", g.Path)
		} else {
			msg = fmt.Sprintf("%s\n%s generated", msg, g.Path)
		}
	}

	return msg, nil
}
