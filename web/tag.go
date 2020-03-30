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
package web

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/capitalone/checks-out/model"
	"github.com/capitalone/checks-out/remote"
	"github.com/hashicorp/go-version"
	log "github.com/sirupsen/logrus"
)

func tagIfEnabled(c context.Context, user *model.User, hook *StatusHook,
	req *model.ApprovalRequest, policy *model.ApprovalPolicy, SHA string) (string, error) {

	var tagConfig *model.TagConfig

	if policy.Tag == nil {
		tagConfig = &req.Config.Tag
	} else {
		tagConfig = policy.Tag
	}

	if tagConfig.Enable {
		return doTag(c, user, hook, req, policy, SHA)
	}
	return "", nil
}

func doTag(c context.Context, user *model.User, hook *StatusHook,
	req *model.ApprovalRequest, policy *model.ApprovalPolicy, SHA string) (string, error) {
	var tagConfig *model.TagConfig
	var vers string

	if policy.Tag == nil {
		tagConfig = &req.Config.Tag
	} else {
		tagConfig = policy.Tag
	}

	var err error
	if strings.HasPrefix(tagConfig.Alg, "timestamp") {
		vers, err = handleTimestamp(tagConfig)
	} else if tagConfig.Alg == "semver" {
		vers, err = handleSemver(c, user, hook, req, policy)
	} else if tagConfig.Alg == "explicit" {
		vers = handleExplicit(req, policy)
		//if no version was found, just return
		if vers == "" {
			return "", nil
		}
	} else {
		log.Warnf("Repo %s should have had a valid tag algorithm configured -- using semver",
			req.Repository.Name)
		vers, err = handleSemver(c, user, hook, req, policy)
	}
	if err != nil {
		return "", err
	}
	tag, err := tagConfig.GenerateTag(model.TemplateTag{Version: vers})
	if err != nil {
		return "", err
	}
	log.Debugf("Tagging merge from PR with tag: %s", tag)
	err = remote.Tag(c, user, req.Repository, tag, SHA)
	if err != nil {
		return "", err
	}
	return tag, nil
}

const modifiedRFC3339 = "2006-01-02T15.04.05Z"

func handleTimestamp(tag *model.TagConfig) (string, error) {
	/*
		All times are in UTC
		Valid format strings:
		- timestamp-rfc3339: RFC 3339 format
		- timestamp-millis: milliseconds since the epoch
	*/
	curTime := time.Now().UTC()
	var format string
	switch tag.Alg {
	case "timestamp-millis":
		//special case, return from here
		out := fmt.Sprintf("%d", curTime.Unix())
		return out, nil
	case "timestamp-rfc3339":
		format = modifiedRFC3339
	default:
		log.Warnf("invalid tag format %s. Using modified rfc3339", tag.Alg)
		format = modifiedRFC3339
	}
	out := curTime.Format(format)
	return out, nil
}

func increment(segments []int, tag *model.TagConfig) {
	switch tag.Increment {
	case model.Major:
		segments[0]++
		segments[1] = 0
		segments[2] = 0
	case model.Minor:
		segments[1]++
		segments[2] = 0
	case model.Patch:
		segments[2]++
	case model.None:
		// do nothing
	default:
		log.Errorf("Unknown semver increment %d", tag.Increment)
	}
}

func handleExplicit(req *model.ApprovalRequest, policy *model.ApprovalPolicy) string {
	foundVersion := getLastVersionComment(req, policy)
	return foundVersion
}

func handleSemver(c context.Context, user *model.User, hook *StatusHook, req *model.ApprovalRequest, policy *model.ApprovalPolicy) (string, error) {

	var tagConfig *model.TagConfig

	if policy.Tag == nil {
		tagConfig = &req.Config.Tag
	} else {
		tagConfig = policy.Tag
	}

	// to create the version, need to scan the comments on the pull request to see if anyone specified a version #
	// if so, use the largest specified version #. if not, increment the last version version # for the release
	tags, err := remote.ListTags(c, user, hook.Repo)
	if err != nil {
		log.Warnf("Unable to list tags for %s/%s: %s", hook.Repo.Owner, hook.Repo.Name, err)
	}
	maxVer := getMaxExistingTag(tags)

	foundVersion := getMaxVersionComment(req, policy)

	if foundVersion != nil && foundVersion.GreaterThan(maxVer) {
		maxVer = foundVersion
	} else {
		maxParts := maxVer.Segments()
		increment(maxParts, tagConfig)
		maxVer, _ = version.NewVersion(fmt.Sprintf("%d.%d.%d", maxParts[0], maxParts[1], maxParts[2]))
	}

	verStr := maxVer.String()
	return verStr, nil
}

// getMaxExistingTag is a helper function that scans all passed-in tags for a
// comments with semantic versions. It returns the max version found. If no version
// is found, the function returns a version with the value 0.0.0
func getMaxExistingTag(tags []model.Tag) *version.Version {
	//find the previous largest semver value
	maxVer, _ := version.NewVersion("v0.0.0")

	for _, v := range tags {
		curVer, err := version.NewVersion(string(v))
		if err != nil {
			continue
		}
		if curVer.GreaterThan(maxVer) {
			maxVer = curVer
		}
	}

	log.Debugf("maxVer found is %s", maxVer.String())
	return maxVer
}

func getGroupIndex(re *regexp.Regexp, name string) int {
	for i, n := range re.SubexpNames() {
		if n == name {
			return i
		}
	}
	return 0
}

// getMaxVersionComment is a helper function that analyzes the list of comments
// and returns the maximum version found in a comment. if no matching comment is found,
// the function returns version 0.0.0. If there's a bug in the version pattern,
// nil will be returned.
func getMaxVersionComment(request *model.ApprovalRequest, policy *model.ApprovalPolicy) *version.Version {
	var matcher *regexp.Regexp
	maxVersion, _ := version.NewVersion("0.0.0")
	if policy.Pattern != nil {
		matcher = policy.Pattern.Regex
	} else {
		matcher = request.Config.Pattern.Regex
	}
	if matcher == nil {
		return nil
	}
	index := getGroupIndex(matcher, "version")
	if index == 0 {
		return nil
	}
	model.Approve(request, policy,
		func(f model.Feedback, op model.ApprovalOp) {
			if op != model.Approval {
				return
			}
			body := f.GetBody()
			if len(body) == 0 {
				return
			}
			// verify the comment matches the approval pattern
			match := matcher.FindStringSubmatch(body)
			if len(match) > index {
				//has a version
				curVersion, err := version.NewVersion(match[index])
				if err != nil {
					return
				}
				if maxVersion == nil || curVersion.GreaterThan(maxVersion) {
					maxVersion = curVersion
				}
			}
		})

	return maxVersion
}

// getLastVersionComment is a helper function that analyzes the list of comments
// and returns the last version found in a comment. if no matching comment is found,
// the function returns empty string.
func getLastVersionComment(request *model.ApprovalRequest, policy *model.ApprovalPolicy) string {
	lastVersion := ""
	var matcher *regexp.Regexp
	if policy.Pattern != nil {
		matcher = policy.Pattern.Regex
	} else {
		matcher = request.Config.Pattern.Regex
	}
	if matcher == nil {
		return ""
	}
	index := getGroupIndex(matcher, "version")
	if index == 0 {
		return ""
	}
	model.Approve(request, policy,
		func(f model.Feedback, op model.ApprovalOp) {
			if op != model.Approval {
				return
			}
			body := f.GetBody()
			if len(body) == 0 {
				return
			}
			// verify the comment matches the approval pattern
			match := matcher.FindStringSubmatch(body)
			if len(match) > index {
				//has a version
				curVersion := match[index]
				if curVersion != "" {
					lastVersion = curVersion
				}
			}
		})

	return lastVersion
}

func getCommitComment(request *model.ApprovalRequest, policy *model.ApprovalPolicy) string {
	log.Debugf("request config: %+v", request.Config)
	log.Debugf("policy pattern: %+v", policy.Pattern)
	var matcher *regexp.Regexp
	var commentText string
	if policy.Pattern != nil {
		matcher = policy.Pattern.Regex
	} else {
		matcher = request.Config.Pattern.Regex
	}
	log.Debugf("Matcher: %+v", matcher)
	if matcher == nil {
		return ""
	}
	index := getGroupIndex(matcher, "comment")
	log.Debugf("index: %+v", index)
	if index == 0 {
		return ""
	}
	model.Approve(request, policy,
		func(f model.Feedback, op model.ApprovalOp) {
			if op != model.Approval {
				return
			}
			body := f.GetBody()
			if len(body) == 0 {
				return
			}
			// verify the comment matches the approval pattern
			log.Debugf("comment body: %+v", body)
			match := matcher.FindStringSubmatch(body)
			log.Debugf("match: %+v", match)
			if len(match) > index && len(match[index]) > 0 {
				//has a comment
				commentText = match[index]
				log.Debugf("commentText set to %s", commentText)
			}
		})

	log.Debugf("returning %s", commentText)
	return commentText
}
