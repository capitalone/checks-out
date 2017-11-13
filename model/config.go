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
package model

import (
	"regexp"

	"github.com/capitalone/checks-out/envvars"
	"github.com/capitalone/checks-out/hjson"
	"github.com/capitalone/checks-out/strings/rxserde"

	"github.com/mspiegel/go-multierror"
	"github.com/pelletier/go-toml"
)

type Config struct {
	Approvals   []*ApprovalPolicy   `json:"approvals"`
	Pattern     rxserde.RegexSerde  `json:"pattern"`
	AntiPattern *rxserde.RegexSerde `json:"antipattern,omitempty"`
	AntiTitle   *rxserde.RegexSerde `json:"antititle,omitempty"`
	Commit      CommitConfig        `json:"commit,omitempty"`
	Maintainers MaintainersConfig   `json:"maintainers,omitempty"`
	Merge       MergeConfig         `json:"merge,omitempty"`
	Feedback    FeedbackConfig      `json:"feedback,omitempty"`
	Tag         TagConfig           `json:"tag,omitempty"`
	Comment     CommentConfig       `json:"comment,omitempty"`
	Deployment  DeployConfig        `json:"deploy,omitempty"`
	Audit       AuditConfig         `json:"audit,omitempty"`
	IsOld       bool                `json:"-"`
}

type CommitConfig struct {
	Range         CommitRange `json:"range"`
	AntiRange     CommitRange `json:"antirange"`
	TagRange      CommitRange `json:"tagrange"`
	IgnoreUIMerge bool        `json:"ignoreuimerge,omitempty"`
}

type MaintainersConfig struct {
	Path string `json:"path"`
	Type string `json:"type"`
}

type MergeConfig struct {
	Enable   bool   `json:"enable"`
	UpToDate bool   `json:"uptodate"`
	Method   string `json:"method"`
	Delete   bool   `json:"delete"`
}

type FeedbackConfig struct {
	Types []FeedbackType `json:"types,omitempty"`
}

type CommentConfig struct {
	Enable  bool           `json:"enable"`
	Targets []TargetConfig `json:"targets"`
}

type TargetConfig struct {
	Target  string              `json:"target"`
	Pattern *rxserde.RegexSerde `json:"pattern"`
	Types   []CommentMessage    `json:"types"`
	Names   []string            `json:"names"`
	Url     string              `json:"-"`
}

type DeployConfig struct {
	Enable        bool              `json:"enable"`
	Path          string            `json:"path"`
	DeploymentMap DeploymentConfigs `json:"-"`
}

const (
	maintainers  = "MAINTAINERS"
	maintType    = "text"
	deployment   = "DEPLOYMENTS"
)

func DefaultConfig() *Config {
	c := new(Config)
	c.Approvals = nil
	c.Pattern = rxserde.RegexSerde{Regex: regexp.MustCompile(envvars.Env.Pattern.Default)}
	c.Maintainers.Path = maintainers
	c.Maintainers.Type = maintType
	c.Deployment.Path = deployment
	c.Feedback = DefaultFeedback()
	c.Merge = DefaultMerge()
	c.Tag = DefaultTag()
	c.Audit = DefaultAudit()
	_ = c.Tag.Compile()
	return c
}

func NonEmptyConfig() *Config {
	c := DefaultConfig()
	c.Approvals = []*ApprovalPolicy{DefaultApprovalPolicy()}
	c.Approvals[0].Position = 1
	return c
}

func (c *Config) Validate(caps *Capabilities) error {
	var errs error
	errs = multierror.Append(errs, validateCapabilities(c, caps))
	errs = multierror.Append(errs, validateApprovals(c.Approvals))
	errs = multierror.Append(errs, validateMaintainerConfig(&c.Maintainers))
	return errs
}

func (c *Config) GetFeedbackConfig(policy *ApprovalPolicy) *FeedbackConfig {
	if policy.Feedback == nil {
		return &c.Feedback
	}
	return policy.Feedback
}

func (c *Config) GetMergeConfig(policy *ApprovalPolicy) *MergeConfig {
	if policy.Merge == nil {
		return &c.Merge
	}
	return policy.Merge
}

// ParseConfig parses a project configuration file.
func ParseConfig(data []byte, caps *Capabilities) (*Config, error) {
	c := DefaultConfig()
	err := hjson.Unmarshal(data, c)
	if err != nil {
		return nil, err
	}
	var errs error
	errs = multierror.Append(errs, c.Tag.Compile())
	if errs != nil {
		return nil, errs
	}
	for i, policy := range c.Approvals {
		setupPolicyDefaults(i, policy)
	}
	errs = multierror.Append(errs, c.Validate(caps))
	if errs != nil {
		return nil, errs
	}
	return c, nil
}

func setupPolicyDefaults(i int, policy *ApprovalPolicy) {
	if policy.AuthorMatch == nil {
		policy.AuthorMatch = new(MatcherHolder)
	}
	if policy.AntiMatch == nil {
		policy.AntiMatch = new(MatcherHolder)
	}
	if policy.Scope == nil {
		policy.Scope = DefaultApprovalScope()
	}
	if policy.Match.Matcher == nil {
		policy.Match.Matcher = DefaultMatcher()
	}
	if policy.AntiMatch.Matcher == nil {
		policy.AntiMatch.Matcher = DefaultMatcher()
	}
	if policy.AuthorMatch.Matcher == nil {
		policy.AuthorMatch.Matcher = DefaultUniverseMatch()
	}
	if c, ok := policy.Match.Matcher.(ChangePolicy); ok {
		c.ChangePolicy(policy)
	}
	policy.Position = i + 1
}

// ParseOldConfig parses a projects .lgtm file
func ParseOldConfig(data []byte) (*Config, error) {
	c, err := parseOldConfigStr(string(data))
	if err != nil {
		return nil, err
	}
	// convert to a current config structure
	// we don't map Team because it wasn't publically supported before

	regex, err := regexp.Compile(c.Get("pattern").(string))
	if err != nil {
		return nil, err
	}

	c2 := NonEmptyConfig()
	c2.Pattern = rxserde.RegexSerde{Regex: regex}
	mm := &MaintainerMatch{
		CommonMatch{
			Approvals: int(c.Get("approvals").(int64)),
			Self:      !c.Get("self_approval_off").(bool),
		},
	}
	c2.Approvals[0].Match.Matcher = mm
	//clear out unused fields
	c2.Tag = TagConfig{}
	c2.Deployment = DeployConfig{}
	c2.IsOld = true
	c2.Maintainers.Type = "legacy"
	return c2, nil
}

//all of these variables are for old file format support
var (
	oldApprovals       = envvars.Env.Old.Approvals
	oldPattern         = envvars.Env.Old.Pattern
	oldSelfApprovalOff = envvars.Env.Old.SelfApprovalOff
)

// parseOldConfigStr parses a projects .lgtm file in string format.
func parseOldConfigStr(data string) (*toml.Tree, error) {
	c, err := toml.Load(data)
	if err != nil {
		return nil, err
	}
	if !c.Has("approvals") {
		c.Set("approvals", oldApprovals)
	}
	if !c.Has("pattern") {
		c.Set("pattern", oldPattern)
	}
	if !c.Has("self_approval_off") {
		c.Set("self_approval_off", oldSelfApprovalOff)
	}
	return c, nil
}
