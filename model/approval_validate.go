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
package model

import (
	"fmt"

	"github.com/capitalone/checks-out/set"

	multierror "github.com/mspiegel/go-multierror"
	"github.com/pkg/errors"
)

func validateApprovals(approvals []*ApprovalPolicy) error {
	var errs error
	if len(approvals) == 0 {
		return errors.New("No approval policies specified")
	}
	tailScope := approvals[len(approvals)-1].Scope
	errs = multierror.Append(errs, tailScope.ValidateFinal())
	for _, approval := range approvals {
		errs = multierror.Append(errs, approval.Validate())
	}
	errs = validateNames(approvals, errs)
	return errs
}

func (a *ApprovalScope) ValidateFinal() error {
	var errs error
	if len(a.Paths) > 0 {
		errs = multierror.Append(errs, errors.New("Final scope must have no paths"))
	}
	if len(a.Branches) > 0 {
		errs = multierror.Append(errs, errors.New("Final scope must have no branches"))
	}
	if len(a.PathRegexp) > 0 {
		errs = multierror.Append(errs, errors.New("Final scope must have no path regexp"))
	}
	if len(a.BaseRegexp) > 0 {
		errs = multierror.Append(errs, errors.New("Final scope must have no base branch regular expressions"))
	}
	if len(a.CompareRegexp) > 0 {
		errs = multierror.Append(errs, errors.New("Final scope must have no compare branch regular expressions"))
	}
	if len(a.Nested) > 0 {
		errs = multierror.Append(errs, errors.New("Final scope must have no nested scopes"))
	}
	return errs
}

func (a *ApprovalPolicy) Validate() error {
	var errs error
	if a.Tag != nil {
		errs = multierror.Append(errs, a.Tag.Compile())
	}
	if len(a.Scope.Paths) > 0 && len(a.Scope.PathRegexp) > 0 {
		err := errors.New("'paths' and 'regexpaths' cannot be used together")
		errs = multierror.Append(errs, err)
	}
	if len(a.Scope.Branches) > 0 && len(a.Scope.BaseRegexp) > 0 {
		err := errors.New("'branches' and 'regexbase' cannot be used together")
		errs = multierror.Append(errs, err)
	}
	if len(a.Scope.Nested) > 0 && (len(a.Scope.Paths) > 0 || len(a.Scope.PathRegexp) > 0) {
		err := errors.New("nested scopes cannot be used with 'paths' or 'regexpaths'")
		errs = multierror.Append(errs, err)
	}
	return errs
}

func validateNames(approvals []*ApprovalPolicy, errs error) error {
	names := set.Empty()
	dupls := set.Empty()
	for _, approval := range approvals {
		name := approval.Name
		if len(name) > 0 {
			if names.Contains(name) {
				if !dupls.Contains(name) {
					err := fmt.Errorf("The approval scope name '%s' is used more than once", name)
					errs = multierror.Append(errs, err)
					dupls.Add(name)
				}
			} else {
				names.Add(name)
			}
		}
	}
	return errs
}

func (match CommonMatch) Validate(_ *MaintainerSnapshot) error {
	if match.Approvals <= 0 {
		return errors.New("approval count must be positive")
	}
	return nil
}

func (match *EntityMatch) Validate(m *MaintainerSnapshot) error {
	var errs error
	errs = multierror.Append(errs, match.CommonMatch.Validate(m))
	ent := match.Entity.String()
	_, org := m.Org[ent]
	_, person := m.People[ent]
	if !org && !person {
		err := fmt.Errorf("%s must be either org or person", ent)
		errs = multierror.Append(errs, err)
	}
	return errs
}

func (match *AnonymousMatch) Validate(m *MaintainerSnapshot) error {
	var errs error
	errs = multierror.Append(errs, match.CommonMatch.Validate(m))
	ents := match.Entities
	for e := range ents {
		ent := e.String()
		_, org := m.Org[ent]
		_, person := m.People[ent]
		if !org && !person {
			err := fmt.Errorf("%s must be either org or person", ent)
			errs = multierror.Append(errs, err)
		}
	}
	return errs
}

func (match *DisableMatch) Validate(_ *MaintainerSnapshot) error {
	return nil
}

func (match *TrueMatch) Validate(_ *MaintainerSnapshot) error {
	return nil
}

func (match *FalseMatch) Validate(_ *MaintainerSnapshot) error {
	return nil
}

func (match *AndMatch) Validate(m *MaintainerSnapshot) error {
	var errs error
	for _, a := range match.And {
		err := a.Validate(m)
		errs = multierror.Append(errs, err)
	}
	return errs
}

func (match *OrMatch) Validate(m *MaintainerSnapshot) error {
	var errs error
	for _, o := range match.Or {
		err := o.Validate(m)
		errs = multierror.Append(errs, err)
	}
	return errs
}

func (match *NotMatch) Validate(m *MaintainerSnapshot) error {
	return match.Not.Validate(m)
}

func (match *AtLeastMatch) Validate(m *MaintainerSnapshot) error {
	var errs error
	if match.Approvals <= 0 {
		errs = errors.New("approval count must be positive")
	}
	for _, c := range match.Choose {
		err := c.Validate(m)
		errs = multierror.Append(errs, err)
	}
	return errs
}

func (match *AuthorMatch) Validate(m *MaintainerSnapshot) error {
	return match.Inner.Validate(m)
}

func (match *IssueAuthorMatch) Validate(_ *MaintainerSnapshot) error {
	return nil
}
