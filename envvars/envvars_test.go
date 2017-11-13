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
package envvars

import (
	"flag"
	"os"
	"testing"

	"github.com/ianschenck/envflag"
	"github.com/mspiegel/go-multierror"
)

var driver, source, email, client, secret string

func setup() {
	envflag.EnvironmentFlags = flag.NewFlagSet("environment", flag.ExitOnError)
	driver = os.Getenv("DB_DRIVER")
	source = os.Getenv("DB_SOURCE")
	email = os.Getenv("GITHUB_EMAIL")
	client = os.Getenv("GITHUB_CLIENT")
	secret = os.Getenv("GITHUB_SECRET")

	os.Unsetenv("DB_DRIVER")
	os.Unsetenv("DB_SOURCE")
	os.Unsetenv("GITHUB_EMAIL")
	os.Unsetenv("GITHUB_CLIENT")
	os.Unsetenv("GITHUB_SECRET")
}

func required() {
	os.Setenv("DB_DRIVER", "sqlite3")
	os.Setenv("DB_SOURCE", "checks-out.sqlite")
	os.Setenv("GITHUB_EMAIL", "broken@example.com")
	os.Setenv("GITHUB_CLIENT", "foo")
	os.Setenv("GITHUB_SECRET", "bar")
}

func teardown() {
	envflag.EnvironmentFlags = flag.NewFlagSet("environment", flag.ExitOnError)
	os.Setenv("DB_DRIVER", driver)
	os.Setenv("DB_SOURCE", source)
	os.Setenv("GITHUB_EMAIL", email)
	os.Setenv("GITHUB_CLIENT", client)
	os.Setenv("GITHUB_SECRET", secret)
}

func TestRequiredVars(t *testing.T) {

	setup()
	configure()

	errs, ok := Validate().(*multierror.Error)

	if !ok {
		t.Error("Validation did not return a multierror")
	}
	if len(errs.Errors) != 5 {
		t.Errorf("Expected 5 errors and %d were generated: %s", len(errs.Errors), errs.Error())
	}

	teardown()
	configure()
}

func TestSSL(t *testing.T) {

	setup()
	required()
	cert := os.Getenv("SERVER_CERT")
	key := os.Getenv("SERVER_KEY")
	os.Setenv("SERVER_CERT", "foobar")
	os.Unsetenv("SERVER_KEY")
	configure()

	err := Validate()

	if err.Error() != "Both server SSL certificate and SSL must be specified for SSL." {
		t.Error("SSL error not reported ", err.Error())
	}

	teardown()
	os.Setenv("SERVER_CERT", cert)
	os.Setenv("SERVER_KEY", key)
	configure()
}

func TestLogLevel(t *testing.T) {
	setup()
	required()
	logLevel := os.Getenv("LOG_LEVEL")
	os.Setenv("LOG_LEVEL", "foobar")
	configure()

	err := Validate()
	exp := "Environment variable LOG_LEVEL 'foobar' must be one of: 'debug', 'info', 'warn', 'error', 'fatal', 'panic'"
	if err.Error() != exp {
		t.Error("Log level error incorrect ", err.Error())
	}

	teardown()
	os.Setenv("LOG_LEVEL", logLevel)
	configure()
}
