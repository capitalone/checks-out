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
package store

import (
	"context"
	"testing"
	"time"

	"github.com/capitalone/checks-out/cache"
)

type mockStore struct {
	Store
	callCount int
}

func (ms *mockStore) GetValidOrgs() ([]string, error) {
	ms.callCount++
	return []string{"beatles", "stones", "ledzep"}, nil
}

func TestGetValidOrgs(t *testing.T) {
	s := &mockStore{}

	c := context.Background()
	c = context.WithValue(c, "store", s)
	c = context.WithValue(c, "cache", cache.NewTTL(40*time.Millisecond))

	orgs, err := GetValidOrgs(c)
	validHelper := func() {
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if orgs == nil {
			t.Fatal("Expected non-nil Set returned")
		}
		if !orgs.Contains("beatles") {
			t.Error("Should have contained beatles")
		}
		if !orgs.Contains("stones") {
			t.Error("Should have contained stones")
		}
		if !orgs.Contains("ledzep") {
			t.Error("Should have contained ledzep")
		}
	}
	validHelper()
	if s.callCount != 1 {
		t.Errorf("Expected 1 call to store, got %d", s.callCount)
	}
	//test caching is doing its job
	orgs, err = GetValidOrgs(c)
	validHelper()
	if s.callCount != 1 {
		t.Errorf("Expected 1 call to store, got %d", s.callCount)
	}
	//sleep to let cache expire
	time.Sleep(50 * time.Millisecond)
	orgs, err = GetValidOrgs(c)
	validHelper()
	if s.callCount != 2 {
		t.Errorf("Expected 2 calls to store, got %d", s.callCount)
	}
}
