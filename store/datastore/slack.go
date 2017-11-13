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

// Returns the slack URL for the specified hostname and user
// if the user string is blank, the default (admin-level) hostname is returned
// if no hostname is found, an empty string is returned
func (db *datastore) GetSlackUrl(hostname string, user string) (string, error) {
	rows, err := db.Query(getSlackStmt[db.curDB], hostname, user)
	if err != nil {
		return "", err
	}
	defer rows.Close()
	if !rows.Next() {
		return "", nil
	}
	url := new(string)
	err = rows.Scan(url)

	if url == nil {
		*url = ""
	}
	return *url, err
}

// Stores or updates the slack URL for the specified hostname and user
// if the user string is blank, the default (admin-level) hostname is stored or updated
func (db *datastore) AddUpdateSlackUrl(hostname string, user string, url string) error {
	query := upsertSlackStmt[db.curDB]
	var values []interface{}
	switch db.curDB {
	case POSTGRES:
		values = []interface{}{hostname, user, url, url, hostname, user}
	case MYSQL:
		values = []interface{}{hostname, user, url, url}
	case SQLITE:
		values = []interface{}{hostname, user, url}
	}
	_, err := db.Exec(query, values...)
	return err
}

// Deletes the slack URL for the specified hostname and user
// if the user string is blank, the default (admin-level) hostname is deleted
func (db *datastore) DeleteSlackUrl(hostname string, user string) error {
	_, err := db.Exec(deleteSlackStmt[db.curDB], hostname, user)
	return err
}

var upsertSlackStmt = map[string]string{
	POSTGRES: `
	INSERT into slack_urls(host_name, user, url)
	VALUES ($1, $2, $3)
	ON CONFLICT DO UPDATE
	SET url = $4 WHERE host_name = $5 and user = $6
	`,
	MYSQL: `
	INSERT into slack_urls(host_name, user, url)
	VALUES (?, ?, ?)
	ON DUPLICATE KEY UPDATE url=?
	`,
	SQLITE: `
	INSERT OR UPDATE into slack_urls(host_name, user, url)
	VALUES (?, ?, ?)
	`,
}

var getSlackStmt = map[string]string{
	POSTGRES: `
	SELECT url FROM slack_urls
	WHERE host_name = $1 and user = $2
	`,
	MYSQL: `
	SELECT url FROM slack_urls
	WHERE host_name = ? and user = ?
	`,
	SQLITE: `
	SELECT url FROM slack_urls
	WHERE host_name = ? and user = ?
	`,
}

var deleteSlackStmt = map[string]string{
	POSTGRES: `
	DELETE FROM slack_urls
	WHERE host_name = $1 and user = $2
	`,
	MYSQL: `
	DELETE FROM slack_urls
	WHERE host_name = ? and user = ?
	`,
	SQLITE: `
	DELETE FROM slack_urls
	WHERE host_name = ? and user = ?
	`,
}
