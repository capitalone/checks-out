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
package usage

import (
	"context"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
)

type usageType int

const (
	event usageType = iota
)

func AddEventToContext(c context.Context, e string) context.Context {
	return context.WithValue(c, event, e)
}

func GetEventFromContext(c context.Context) string {
	e, ok := c.Value(event).(string)
	if !ok {
		return ""
	}
	return e
}

var lock = sync.Mutex{}

type Usage struct {
	Users     map[string]int `json:"users"`
	HookIn    map[string]int `json:"hook_in"`
	HookOut   map[string]int `json:"hook_out"`
	RemoteReq map[string]int `json:"remote"`
}

var data Usage

func init() {
	data = createUsage()
}

func createUsage() Usage {
	return Usage{
		Users:     make(map[string]int),
		HookIn:    make(map[string]int),
		HookOut:   make(map[string]int),
		RemoteReq: make(map[string]int),
	}
}

func RecordIncomingWebHook(event string) {
	lock.Lock()
	data.HookIn[event]++
	lock.Unlock()
}

func RecordApiRequest(user string, event string, req string) {
	lock.Lock()
	data.Users[user]++
	data.HookOut[event]++
	data.RemoteReq[req]++
	lock.Unlock()
}

func copyMap(dst, src map[string]int) {
	for k, v := range src {
		dst[k] = v
	}
}

func GetStats() Usage {
	stats := createUsage()
	lock.Lock()
	copyMap(stats.Users, data.Users)
	copyMap(stats.HookIn, data.HookIn)
	copyMap(stats.HookOut, data.HookOut)
	copyMap(stats.RemoteReq, data.RemoteReq)
	lock.Unlock()
	return stats
}

func writeLog() {
	log.Info("Usage statistics for the past hour")
	for k, v := range data.Users {
		log.Infof("User %s : %d api requests", k, v)
	}
	for k, v := range data.HookIn {
		log.Infof("Hook %s : %d incoming requests", k, v)
	}
	for k, v := range data.HookOut {
		log.Infof("Hook %s : %d outgoing requests", k, v)
	}
	for k, v := range data.RemoteReq {
		log.Infof("Remote request %s : %d requests", k, v)
	}
}

func resetStats() {
	data = createUsage()
}

func usageTask() {
	wait := 60 - time.Now().Minute()
	timer := time.NewTimer(time.Duration(wait) * time.Minute)
	<-timer.C
	ticker := time.NewTicker(time.Hour)
	for {
		lock.Lock()
		writeLog()
		resetStats()
		lock.Unlock()
		<-ticker.C
	}
}

func Start() {
	go usageTask()
}
