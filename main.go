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
package main

import (
	"flag"
	"fmt"
	"net/http"
	"time"

	"github.com/capitalone/checks-out/envvars"
	"github.com/capitalone/checks-out/logstats"
	"github.com/capitalone/checks-out/migration"
	_ "github.com/capitalone/checks-out/notifier/github"
	_ "github.com/capitalone/checks-out/notifier/slack"
	"github.com/capitalone/checks-out/remote"
	"github.com/capitalone/checks-out/router"
	"github.com/capitalone/checks-out/store/datastore"
	"github.com/capitalone/checks-out/usage"
	"github.com/capitalone/checks-out/version"

	"github.com/Sirupsen/logrus"
	_ "github.com/joho/godotenv/autoload"
)

func setLogLevel(level string) {
	switch level {
	case "panic":
		logrus.SetLevel(logrus.PanicLevel)
	case "fatal":
		logrus.SetLevel(logrus.FatalLevel)
	case "error":
		logrus.SetLevel(logrus.ErrorLevel)
	case "warn":
		logrus.SetLevel(logrus.WarnLevel)
	case "info":
		logrus.SetLevel(logrus.InfoLevel)
	case "debug":
		logrus.SetLevel(logrus.DebugLevel)
	default:
		logrus.Fatal("Unrecognized log level ", level)
	}
}

func startService() {

	err := envvars.Validate()
	if err != nil {
		logrus.Fatal(err)
	}

	setLogLevel(envvars.Env.Monitor.LogLevel)

	logstats.Start()
	usage.Start()

	r := remote.Get()
	ds := datastore.Get()

	err = migration.Migrate(r, ds)

	if err != nil {
		logrus.Fatal(err)
	}

	handler := router.Load()

	logrus.Infof("Starting %s service on %s", envvars.Env.Branding.ShortName, time.Now().Format(time.RFC1123))

	if envvars.Env.Server.Cert != "" {
		logrus.Fatal(
			http.ListenAndServeTLS(envvars.Env.Server.Addr, envvars.Env.Server.Cert, envvars.Env.Server.Key, handler),
		)
	} else {
		logrus.Fatal(
			http.ListenAndServe(envvars.Env.Server.Addr, handler),
		)
	}

}

func main() {
	ver := flag.Bool("version", false, "print version")
	env := flag.Bool("env", false, "print environment variables")
	help := flag.Bool("help", false, "print help information")
	flag.Parse()
	if *help {
		flag.PrintDefaults()
	} else if *ver {
		fmt.Println(version.Version)
	} else if *env {
		envvars.Usage()
	} else {
		startService()
	}
}
