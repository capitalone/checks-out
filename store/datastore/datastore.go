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
package datastore

import (
	"database/sql"
	"sync"
	"time"

	"github.com/capitalone/checks-out/envvars"
	"github.com/capitalone/checks-out/set"
	"github.com/capitalone/checks-out/store"
	"github.com/capitalone/checks-out/store/migration"

	"github.com/Sirupsen/logrus"
	// bindings for meddler
	_ "github.com/go-sql-driver/mysql"
	// bindings for meddler
	_ "github.com/mattn/go-sqlite3"
	// bindings for meddler
	_ "github.com/lib/pq"
	migrate "github.com/rubenv/sql-migrate"
	"github.com/russross/meddler"
)

type datastore struct {
	*sql.DB
	Migrations set.Set
	curDB      string
}

var once sync.Once
var cachedStore store.Store

func Get() store.Store {
	once.Do(func() {
		cachedStore = create(envvars.Env.Db.Driver, envvars.Env.Db.Datasource)
	})
	return cachedStore
}

// creates a database connection for the given driver and datasource
// and returns a new Store.
func create(driver, config string) store.Store {
	db, migrations := Open(driver, config)
	return From(db, migrations, driver)
}

// From returns a Store using an existing database connection.
func From(db *sql.DB, migrations set.Set, driver string) store.Store {
	return &datastore{db, migrations, driver}
}

// Open opens a new database connection with the specified
// driver and connection string and returns a store.
func Open(driver, config string) (*sql.DB, set.Set) {
	db, err := sql.Open(driver, config)
	if err != nil {
		logrus.Errorln(err)
		logrus.Fatalln("database connection failed")
	}

	setupMeddler(driver)

	logrus.Debugf("Driver %s", driver)
	logrus.Debugf("Data Source %s", config)

	if err = pingDatabase(db); err != nil {
		logrus.Errorln(err)
		logrus.Fatalln("database ping attempts failed")
	}
	ids, err := setupDatabase(driver, db)
	if err != nil {
		logrus.Errorln(err)
		logrus.Fatalln("migration failed")
	}
	return db, ids
}

// helper function to ping the database with backoff to ensure
// a connection can be established before we proceed with the
// database setup and migration.
func pingDatabase(db *sql.DB) (err error) {
	for i := 0; i < 30; i++ {
		err = db.Ping()
		if err == nil {
			return
		}
		logrus.Infof("database ping failed. retry in 1s. %s", err)
		time.Sleep(time.Second)
	}
	return
}

// helper function to setup the databsae by performing
// automated database migration steps.
func setupDatabase(driver string, db *sql.DB) (set.Set, error) {
	var migrations = &migrate.AssetMigrationSource{
		Asset:    migration.Asset,
		AssetDir: migration.AssetDir,
		Dir:      driver,
	}
	todo, _, err := migrate.PlanMigration(db, driver, migrations, migrate.Up, 0)
	if err != nil {
		return nil, err
	}
	_, err = migrate.Exec(db, driver, migrations, migrate.Up)
	if err != nil {
		return nil, err
	}
	done := set.Empty()
	for _, m := range todo {
		done.Add(m.Id)
	}
	return done, nil
}

// helper function to setup the meddler default driver
// based on the selected driver name.
func setupMeddler(driver string) {
	switch driver {
	case "sqlite3":
		meddler.Default = meddler.SQLite
	case "mysql":
		meddler.Default = meddler.MySQL
	case "postgres":
		meddler.Default = meddler.PostgreSQL
	}
}

const (
	POSTGRES = "postgres"
	MYSQL    = "mysql"
	SQLITE   = "sqlite3"
)
