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
package model

import (
	"errors"

	"github.com/capitalone/checks-out/hjson"
)

type DeploymentConfigs map[string]DeploymentConfig

type DeploymentConfig struct {
	Tasks       []string `json:"tasks"`
	Environment *string  `json:"env"`
}

type DeploymentInfo struct {
	Ref         string
	Task        string
	Environment string
}

func (c *Config) LoadDeploymentMap(deployData []byte) error {
	if len(deployData) == 0 {
		return errors.New("No content in deployment map")
	}
	err := hjson.Unmarshal(deployData, &c.Deployment.DeploymentMap)
	return err
}
