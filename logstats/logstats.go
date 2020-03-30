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
package logstats

import (
	"sync"
	"time"

	"github.com/capitalone/checks-out/envvars"
	"github.com/capitalone/checks-out/model"
	"github.com/capitalone/checks-out/set"
	"github.com/capitalone/checks-out/store/datastore"

	log "github.com/sirupsen/logrus"
)

var (
	lock         = sync.Mutex{}
	pr           = set.Empty()
	approvers    = set.Empty()
	disapprovers = set.Empty()
)

func RecordPR(id string) {
	lock.Lock()
	pr.Add(id)
	lock.Unlock()
}

func RecordApprover(id string) {
	lock.Lock()
	approvers.Add(id)
	lock.Unlock()
}

func RecordDisapprover(id string) {
	lock.Lock()
	disapprovers.Add(id)
	lock.Unlock()
}

func resetStats() {
	pr = set.Empty()
	approvers = set.Empty()
	disapprovers = set.Empty()
}

func usersAndOrgs(repos []*model.Repo) (set.Set, set.Set) {
	users := set.Empty()
	orgs := set.Empty()
	for _, repo := range repos {
		if repo.Org {
			orgs.Add(repo.Owner)
		} else {
			users.Add(repo.Owner)
		}
	}
	return users, orgs
}

func writeLog() {
	repos, err := datastore.Get().GetAllRepos()
	if err != nil {
		log.Error("Periodic logging unable to fetch repository list", err)
	} else {
		users, orgs := usersAndOrgs(repos)
		log.Infof("Monitoring %d repositories", len(repos))
		log.Infof("Monitoring %d users", len(users))
		log.Infof("Monitoring %d organizations", len(orgs))
	}
	log.Infof("Accepted %d pull requests in last period", len(pr))
	log.Infof("Accepted %d approvers in last period", len(approvers))
	log.Infof("Accepted %d disapprovers in last period", len(disapprovers))
}

func logTask() {
	period := envvars.Env.Monitor.LogPeriod
	if period == 0 {
		return
	}
	t := time.NewTicker(period)
	for {
		lock.Lock()
		writeLog()
		resetStats()
		lock.Unlock()
		<-t.C
	}
}

func Start() {
	go logTask()
}
