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
package github

import (
	"context"

	"github.com/capitalone/checks-out/exterror"
	"github.com/capitalone/checks-out/model"
	"github.com/capitalone/checks-out/set"

	"github.com/mspiegel/go-github/github"
)

var systemAccounts = set.New("GitHub", "GitHub Enterprise")

func isCommitUIMerge(commit *github.RepositoryCommit) bool {
	return len(commit.Parents) == 2 && systemAccounts.Contains(*commit.Commit.Committer.Name)
}

// ignoreUIMerge will find the first commit that is not a merge
// created through the user interface.
func ignoreUIMerge(ctx context.Context, client *github.Client, r *model.Repo, pr *github.PullRequest,
	commit *github.RepositoryCommit) (*github.RepositoryCommit, error) {

	// Test the current commit before fetching commits from the base branch
	if !isCommitUIMerge(commit) {
		return commit, nil
	}

	// Fetch commits from the base branch that have occurred since
	// the pull request was opened. These are the commits to ignore.
	lcOpts := github.CommitsListOptions{SHA: *pr.Base.Ref, Since: *pr.CreatedAt}
	commits := set.Empty()
	resp, err := buildCompleteList(func(opts *github.ListOptions) (*github.Response, error) {
		lcOpts.ListOptions = *opts
		next, resp, err := client.Repositories.ListCommits(ctx, r.Owner, r.Name, &lcOpts)
		for _, c := range next {
			commits.Add(*c.SHA)
		}
		return resp, err
	})
	if err != nil {
		return nil, exterror.Create(resp.StatusCode, err)
	}
	return followCommit(ctx, client, r, commit, commits)
}

// followCommit walks up the merge commits and selects the
// first commit that is not a merge commit. Commits from
// the base branch are ignored.
func followCommit(ctx context.Context, client *github.Client, r *model.Repo, commit *github.RepositoryCommit,
	baseref set.Set) (*github.RepositoryCommit, error) {

	if !isCommitUIMerge(commit) {
		return commit, nil
	}
	left := &commit.Parents[0]
	right := &commit.Parents[1]
	if baseref.Contains(*left.SHA) && baseref.Contains(*right.SHA) {
		return commit, nil
	}
	if !baseref.Contains(*left.SHA) && !baseref.Contains(*right.SHA) {
		return commit, nil
	}
	var target *github.Commit
	if baseref.Contains(*left.SHA) {
		target = right
	}
	if baseref.Contains(*right.SHA) {
		target = left
	}
	commit, resp, err := client.Repositories.GetCommit(ctx, r.Owner, r.Name, *target.SHA)
	if err != nil {
		return nil, exterror.Create(resp.StatusCode, err)
	}
	return followCommit(ctx, client, r, commit, baseref)
}
