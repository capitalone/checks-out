/*

SPDX-Copyright: Copyright (c) Capital One Services, LLC
SPDX-License-Identifier: Apache-2.0
Copyright 2017 Capital One Services, LLC

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and limitations under the License.

*/
package exterror

import (
	"errors"
	"strings"
	"testing"

	"github.com/mspiegel/go-multierror"
)

func TestConvert(t *testing.T) {
	i := ExtError{Status: 404, Err: errors.New("foobar")}
	o := Convert(i)
	if o.Status != i.Status {
		t.Error("Expected status 404")
	}
	if o.Err.Error() != i.Err.Error() {
		t.Error("Incorrect error message")
	}
}

func TestAppend(t *testing.T) {
	prev := ExtError{Status: 404, Err: errors.New("foobar")}
	err := Append(prev, "baz")
	if prev.Status != err.(ExtError).Status {
		t.Error("Expected status 404")
	}
	if err.Error() != "baz. foobar" {
		t.Errorf("Incorrect error message: %s", err.Error())
	}
	plain := errors.New("foobar")
	err = Append(plain, "baz")
	if err.Error() != "baz. foobar" {
		t.Errorf("Incorrect error message: %s", err.Error())
	}
	var errs error
	e1 := Create(400, errors.New("foo"))
	e2 := Create(400, errors.New("bar"))
	e3 := Create(400, errors.New("baz"))
	errs = multierror.Append(errs, e1, e2, e3)
	errs = Append(errs, "prefix")
	if 400 != errs.(ExtError).Status {
		t.Error("Expected status 400")
	}
	if !strings.HasPrefix(errs.Error(), "prefix.") {
		t.Errorf("Incorrect error message: %s", errs.Error())
	}
}

func TestConvertMultiError(t *testing.T) {
	e1 := Create(404, nil)
	e2 := Create(401, nil)
	e3 := Create(500, nil)
	out := convertMultiError(new(multierror.Error))
	if out.Status != 500 {
		t.Error("Expected status 500")
	}
	out = convertMultiError(multierror.Append(nil, e1, e2, e3).(*multierror.Error))
	if out.Status != 500 {
		t.Error("Expected status 500")
	}
	out = convertMultiError(multierror.Append(nil, e1, e2).(*multierror.Error))
	if out.Status != 400 {
		t.Error("Expected status 400")
	}
	out = convertMultiError(multierror.Append(nil, e1, e2, errors.New("foobar")).(*multierror.Error))
	if out.Status != 500 {
		t.Error("Expected status 500")
	}
}
