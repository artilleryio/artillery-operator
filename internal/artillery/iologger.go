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
	"io"

	"github.com/go-logr/logr"
)

func NewIOLogger(stdOut io.Writer, stdErr io.Writer) logr.Logger {
	return &IOLogger{stdOut: stdOut, stdErr: stdErr, keysAndValues: []interface{}{}}
}

type IOLogger struct {
	stdOut        io.Writer
	stdErr        io.Writer
	keysAndValues []interface{}
}

func (l *IOLogger) Enabled() bool                 { return true }
func (l *IOLogger) V(_ int) logr.Logger           { return l }
func (l *IOLogger) WithName(_ string) logr.Logger { return l }

func (l *IOLogger) Info(msg string, keysAndValues ...interface{}) {
	if len(l.keysAndValues) > 0 {
		_, _ = l.stdOut.Write([]byte(fmt.Sprintf("%s %+v %+v\n", msg, l.keysAndValues, keysAndValues)))
	} else {
		_, _ = l.stdOut.Write([]byte(fmt.Sprintf("%s %+v\n", msg, keysAndValues)))
	}
}

func (l *IOLogger) Error(err error, msg string, keysAndValues ...interface{}) {
	if len(l.keysAndValues) > 0 {
		_, _ = l.stdErr.Write([]byte(fmt.Sprintf("%s ... %s %+v %+v\n", err.Error(), msg, l.keysAndValues, keysAndValues)))
	} else {
		_, _ = l.stdErr.Write([]byte(fmt.Sprintf("%s ... %s %+v\n", err.Error(), msg, keysAndValues)))
	}
}

func (l *IOLogger) WithValues(keysAndValues ...interface{}) logr.Logger {
	l.keysAndValues = keysAndValues
	return l
}
