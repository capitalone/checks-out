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
	"time"

	"github.com/capitalone/checks-out/model"
	"github.com/capitalone/checks-out/remote"
)

const noAudit = "the entire branch history. No audits found in the past 30 commits"

func hasAuditStatus(c context.Context, u *model.User, r *model.Repo, context, sha string) (bool, error) {
	status, err := remote.GetStatus(c, u, r, sha)
	if err != nil {
		return false, err
	}
	_, ok := status.Statuses[context]
	return ok, nil
}

func requireAudit(cfg *model.Config, pr *model.PullRequest) bool {
	return cfg.Audit.Enable && cfg.Audit.Branches.Contains(pr.Branch.BaseName)
}

// testAudit determines whether a pull request preserves the audit chain
func testAudit(c context.Context, u *model.User, r *model.Repo, pr *model.PullRequest) (bool, error) {
	ctx := model.AuditName + "/" + pr.Branch.BaseName
	// If the base branch has an audit stamp then the chain is preserved
	ok, err := hasAuditStatus(c, u, r, ctx, pr.Branch.BaseSHA)
	if err != nil {
		return false, err
	}
	if ok {
		return true, nil
	}
	// If the compare branch has an audit stamp then the stamp was created
	// by the service to indicate manual approval of the audit chain.
	ok, err = hasAuditStatus(c, u, r, ctx, pr.Branch.CompareSHA)
	if err != nil {
		return false, err
	}
	if ok {
		return true, nil
	}
	return false, nil
}

func applyAudit(c context.Context, u *model.User, r *model.Repo, pr *model.PullRequest) error {
	ctx := model.AuditName + "/" + pr.Branch.BaseName
	state := "success"
	desc := fmt.Sprintf("audited by pr %d", pr.Number)
	return remote.SetStatus(c, u, r, pr.Branch.MergeCommitSHA, ctx, state, desc)
}

func findAuditRange(c context.Context, u *model.User, r *model.Repo, pr *model.PullRequest) (string, error) {
	ctx := model.AuditName + "/" + pr.Branch.BaseName
	commits, _, err := remote.GetCommits(c, u, r, pr.Branch.BaseSHA, 1, 30)
	if err != nil {
		return "", err
	}
	for _, commit := range commits {
		var ok bool
		ok, err = hasAuditStatus(c, u, r, ctx, commit)
		if err != nil {
			return "", err
		}
		if ok {
			auditRange := fmt.Sprintf("the range %s",
				remote.CreateURLCompare(c, u, r, commit, pr.Branch.BaseSHA))
			return auditRange, nil
		}
	}
	return noAudit, nil
}

func manualAudit(c context.Context, u *model.User, r *model.Repo, pr *model.PullRequest) error {
	base := pr.Branch.BaseName
	ctx := model.AuditName + "/" + base
	branchName := fmt.Sprintf("pr-%d-audit-%d", pr.Number, time.Now().Unix())
	prTitle := fmt.Sprintf("Audit branch %s for pr %d", base, pr.Number)
	auditRange, err := findAuditRange(c, u, r, pr)
	if err != nil {
		return nil
	}
	prBody := fmt.Sprintf(`Please review the commits in %s. `+
		`You must review commits that were not submitted via pull request. `+
		`Approving this pull request indicates `+
		`that you have reviewed the commits to branch %s.`, auditRange, base)
	commitMsg := fmt.Sprintf(`empty commit. Added commit status branch '%s' manual audit

The commits have been reviewed in %s`, base, auditRange)
	commitSha, err := remote.CreateEmptyCommit(c, u, r, pr.Branch.BaseSHA, commitMsg)
	if err != nil {
		return err
	}
	state := "success"
	desc := fmt.Sprintf("manual audit of branch %s", base)
	err = remote.SetStatus(c, u, r, commitSha, ctx, state, desc)
	if err != nil {
		return err
	}
	_, err = remote.CreateReference(c, u, r, commitSha, "heads/"+branchName)
	if err != nil {
		return err
	}
	num, err := remote.CreatePR(c, u, r, prTitle, branchName, base, prBody)
	if err != nil {
		return err
	}
	message := fmt.Sprintf(`You must approve pr #%d to preserve the audit chain. `+
		`Use the "Update branch" button after you have merged the other pull request.`, num)
	err = remote.WriteComment(c, u, r, pr.Number, message)
	if err != nil {
		return err
	}
	return nil
}
