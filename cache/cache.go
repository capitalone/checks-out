/*

SPDX-Copyright: Copyright (c) Brad Rydzewski, project contributors, Capital One Services, LLC
SPDX-License-Identifier: Apache-2.0
Copyright 2017 Brad Rydzewski, project contributors, Capital One Services, LLC

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and limitations under the License.

*/
package cache

import (
	"context"
	"time"

	"github.com/koding/cache"
	"github.com/capitalone/checks-out/envvars"
)

type Cache interface {
	Get(string) (interface{}, error)
	Set(string, interface{}) error
}

func Get(c context.Context, key string) (interface{}, error) {
	return FromContext(c).Get(key)
}

func Set(c context.Context, key string, value interface{}) error {
	return FromContext(c).Set(key, value)
}

// Default creates an in-memory cache with the default
// 30 minute expiration period.
func Default() Cache {
	return NewTTL(time.Minute * 30)
}

// NewTTL returns an in-memory cache with the specified
// ttl expiration period.
func NewTTL(t time.Duration) Cache {
	return cache.NewMemoryWithTTL(t)
}

var (
	Longterm = cache.NewMemoryWithTTL(time.Duration(envvars.Env.Cache.LongCacheTTL))
)
