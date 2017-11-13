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
	"fmt"
	"reflect"

	log "github.com/Sirupsen/logrus"
	"github.com/mspiegel/go-multierror"
)

type ExtError struct {
	Status int
	Err    error
}

func (e ExtError) Error() string {
	return e.Err.Error()
}

func Create(status int, err error) ExtError {
	return ExtError{Status: status, Err: err}
}

func Append(prev error, head string) error {
	prevMsg := prev.Error()
	if len(prevMsg) == 0 {
		return errors.New(head)
	}
	newMsg := fmt.Errorf("%s. %s", head, prevMsg)
	switch v := prev.(type) {
	case ExtError:
		return Create(v.Status, newMsg)
	case *multierror.Error:
		// flatten the multierror to retrieve the http response
		ext := Convert(v)
		return Create(ext.Status, newMsg)
	default:
		return newMsg
	}
}

func Convert(err error) ExtError {
	switch v := err.(type) {
	case ExtError:
		log.Debugf("No conversion necessary for ExtError %s", err.Error())
		return v
	case *multierror.Error:
		log.Debugf("Multierror conversion for %s", err.Error())
		return convertMultiError(v)
	default:
		log.Errorf("Automatic promotion to 500 response for %s", reflect.TypeOf(err).String())
		return ExtError{Status: 500, Err: err}
	}
}

func allExtError(errs *multierror.Error) bool {
	if len(errs.Errors) == 0 {
		return false
	}
	for _, e := range errs.Errors {
		if _, ok := e.(ExtError); !ok {
			return false
		}
	}
	return true
}

func allEqualStatus(errs *multierror.Error) bool {
	resp := errs.Errors[0].(ExtError).Status
	for _, e := range errs.Errors {
		if resp != e.(ExtError).Status {
			return false
		}
	}
	return true
}

func allRangeStatus(errs *multierror.Error, low int, high int) bool {
	for _, e := range errs.Errors {
		status := e.(ExtError).Status
		if (status < low) || (status >= high) {
			return false
		}
	}
	return true
}

func convertMultiError(errs *multierror.Error) ExtError {
	status := 500
	if !allExtError(errs) {
		return ExtError{Status: status, Err: errs}
	}
	if allEqualStatus(errs) {
		status = errs.Errors[0].(ExtError).Status
	} else if allRangeStatus(errs, 400, 500) {
		status = 400
	}
	return ExtError{Status: status, Err: errs}
}
