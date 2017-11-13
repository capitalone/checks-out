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
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/capitalone/checks-out/set"

	"github.com/ianschenck/envflag"
	"github.com/mspiegel/go-multierror"
)

type EnvValues struct {
	// Server configuration
	Server struct {
		Addr string
		Cert string
		Key  string
	}
	// Database configuration
	Db struct {
		Driver     string
		Datasource string
	}
	// Default pattern customization
	Pattern struct {
		Default string
	}
	// External (user-facing) customization
	Branding struct {
		Name      string
		ShortName string
	}
	// Github integration
	Github struct {
		Email      string
		Url        string
		Client     string
		Secret     string
		Scope      string
		AdminOrg   string
		RequestsHz int
	}
	// Slack integration
	Slack struct {
		TargetUrl string
	}
	// Logging/debug config
	Monitor struct {
		LogLevel  string
		Sunlight  bool
		UaList    string
		LogPeriod time.Duration
		DocsUrl   string
	}
	// Caching config
	Cache struct {
		CacheTTL     time.Duration
		LongCacheTTL time.Duration
	}
	// Github testing config
	Test struct {
		GithubToken      string
		GithubTestEnable bool
	}
	// Access config
	Access struct {
		LimitUsers bool
		LimitOrgs  bool
	}

	Old struct {
		Approvals       int64
		Pattern         string
		SelfApprovalOff bool
	}
}

var Env EnvValues

var logLevels = set.New("debug", "info", "warn", "error", "fatal", "panic")

func init() {
	configure()
}

const (
	//splitting up the patterns because one day we might want to automatically append the fieldPattern when it is missing
	fieldPattern = `\s*(version:\s*(?P<version>\S+))?\s*(comment:\s*(?P<comment>.*\S))?\s*`
	pattern      = `(?i)^I approve` + fieldPattern
)

func configure() {
	envflag.StringVar(&Env.Server.Addr, "SERVER_ADDR", ":8000", "Server ip address and port")
	envflag.StringVar(&Env.Server.Cert, "SERVER_CERT", "", "Path to SSL certificate")
	envflag.StringVar(&Env.Server.Key, "SERVER_KEY", "", "SSL certificate key")

	envflag.StringVar(&Env.Db.Driver, "DB_DRIVER", "", "One of sqlite3|postgres|mysql. Required")
	envflag.StringVar(&Env.Db.Datasource, "DB_SOURCE", "", "Database data source. Required")

	envflag.StringVar(&Env.Branding.Name, "BRANDING_NAME", "checks-out", "Branding of this service")
	envflag.StringVar(&Env.Branding.ShortName, "BRANDING_SHORT_NAME", "checks-out", "Abbreviated branding of this service")

	envflag.StringVar(&Env.Pattern.Default, "DEFAULT_PATTERN", pattern, "Default pattern used for matchers")

	envflag.StringVar(&Env.Github.Email, "GITHUB_EMAIL", "", "Email for git commits. Required")
	envflag.StringVar(&Env.Github.Url, "GITHUB_URL", "https://github.com", "Github url")
	envflag.StringVar(&Env.Github.Client, "GITHUB_CLIENT", "", "OAuth2 client id. Required")
	envflag.StringVar(&Env.Github.Secret, "GITHUB_SECRET", "", "OAuth2 secret. Required")
	envflag.StringVar(&Env.Github.Scope, "GITHUB_SCOPE", "read:org,repo:status,admin:repo_hook,admin:org_hook", "Permission scope")
	envflag.StringVar(&Env.Github.AdminOrg, "GITHUB_ADMIN_ORG", "", "GitHub organization with admin privileges")
	envflag.IntVar(&Env.Github.RequestsHz, "GITHUB_BATCH_PER_SECOND", 10, "GitHub batch access rate limiter")

	envflag.StringVar(&Env.Slack.TargetUrl, "SLACK_TARGET_URL", "", "Slack notification url")

	envflag.StringVar(&Env.Monitor.LogLevel, "LOG_LEVEL", "info", "One of debug|info|warn|error|fatal|panic")
	envflag.BoolVar(&Env.Monitor.Sunlight, "CHECKS_OUT_SUNLIGHT", false, "Exposes additional endpoints")
	envflag.StringVar(&Env.Monitor.UaList, "BLACKLIST_USER_AGENTS", "", "Skip logging of these agents")
	envflag.DurationVar(&Env.Monitor.LogPeriod, "LOG_STATS_PERIOD", 0, "Period logging of statistics")
	envflag.StringVar(&Env.Monitor.DocsUrl, "CHECKS_OUT_DOCS_URL", "https://capitalone.github.com/checks-out/docs", "Provides the base URL for links to the documentation.")

	envflag.DurationVar(&Env.Cache.CacheTTL, "CACHE_TTL", time.Minute*15, "Cache length for short lived entries")
	envflag.DurationVar(&Env.Cache.LongCacheTTL, "LONG_CACHE_TTL", time.Hour*24, "Cache length for long lived entries")

	envflag.StringVar(&Env.Test.GithubToken, "GITHUB_TEST_TOKEN", "", "GitHub integration test token")
	envflag.BoolVar(&Env.Test.GithubTestEnable, "GITHUB_TEST_ENABLE", false, "GitHub integration testing")

	envflag.BoolVar(&Env.Access.LimitUsers, "LIMIT_USERS", false, "Only allow users who are listed in the allowed_users table")
	envflag.BoolVar(&Env.Access.LimitOrgs, "LIMIT_ORGS", false, "Only allow users who are members of the orgs listed in the allowed_orgs table")

	envflag.Int64Var(&Env.Old.Approvals, "CHECKS_OUT_APPROVALS", 2, "Legacy default number of approvals")
	envflag.StringVar(&Env.Old.Pattern, "CHECKS_OUT_PATTERN", "(?i)LGTM", "Legacy matching pattern")
	envflag.BoolVar(&Env.Old.SelfApprovalOff, "CHECKS_OUT_SELF_APPROVAL_OFF", false, "Legacy self-approval behavior")

	envflag.Parse()

	Env.Monitor.LogLevel = strings.ToLower(Env.Monitor.LogLevel)
	Env.Github.Url = strings.TrimRight(Env.Github.Url, "/")
}

func Usage() {
	envflag.EnvironmentFlags.PrintDefaults()
}

func Validate() error {
	var errs error
	if Env.Db.Driver == "" {
		err := errors.New("Missing required environment variable DB_DRIVER")
		errs = multierror.Append(errs, err)
	}
	if Env.Db.Datasource == "" {
		err := errors.New("Missing required environment variable DB_SOURCE")
		errs = multierror.Append(errs, err)
	}
	if Env.Github.Email == "" {
		err := errors.New("Missing required environment variable GITHUB_EMAIL")
		errs = multierror.Append(errs, err)
	}
	if Env.Github.Client == "" {
		err := errors.New("Missing required environment variable GITHUB_CLIENT")
		errs = multierror.Append(errs, err)
	}
	if Env.Github.Secret == "" {
		err := errors.New("Missing required environment variable GITHUB_SECRET")
		errs = multierror.Append(errs, err)
	}
	if Env.Github.Url == "" {
		err := errors.New("Environment variable GITHUB_URL is empty")
		errs = multierror.Append(errs, err)
	}
	if (Env.Server.Cert != "" && Env.Server.Key == "") || (Env.Server.Cert == "" && Env.Server.Key != "") {
		err := errors.New("Both server SSL certificate and SSL must be specified for SSL.")
		errs = multierror.Append(errs, err)
	}
	if !logLevels.Contains(Env.Monitor.LogLevel) {
		err := fmt.Errorf("Environment variable LOG_LEVEL '%s' must be one of: %s",
			Env.Monitor.LogLevel,
			"'debug', 'info', 'warn', 'error', 'fatal', 'panic'")
		errs = multierror.Append(errs, err)
	}
	if !strings.HasPrefix(Env.Github.Url, "https://") {
		err := errors.New("GITHUB_URL must have prefix 'https://'")
		errs = multierror.Append(errs, err)
	}
	return errs
}
