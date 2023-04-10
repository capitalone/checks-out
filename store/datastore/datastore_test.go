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
package datastore

import (
	"database/sql"
	"github.com/capitalone/checks-out/set"
	"os"
)

// OpenTest opens a new database connection for testing purposes.
// The database driver and connection string are provided by
// environment variables, with fallback to in-memory sqlite.
func openTest() (*sql.DB, set.Set, string) {
	var (
		driver = "sqlite3"
		config = ":memory:"
	)
	if os.Getenv("TEST_DB_DRIVER") != "" && os.Getenv("TEST_DB_SOURCE") != "" {
		driver = os.Getenv("TEST_DB_DRIVER")
		config = os.Getenv("TEST_DB_SOURCE")
	}
	db, migrations := Open(driver, config)
	return db, migrations, driver
}
