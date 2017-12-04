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
package remote

import (
	"context"
	"net/http"
	"sync"

	"github.com/capitalone/checks-out/model"
	"github.com/capitalone/checks-out/remote/github"
	"github.com/capitalone/checks-out/set"
)

type Remote interface {
	// Capabilities generates the user's capabilities
	Capabilities(context.Context, *model.User) (*model.Capabilities, error)

	// GetUser authenticates a user with the remote system.
	GetUser(context.Context, http.ResponseWriter, *http.Request) (*model.User, error)

	// GetUserToken authenticates a user with the remote system using
	// the remote systems OAuth token.
	GetUserToken(context.Context, string) (string, error)

	// RevokeAuthorization revokes the OAuth token
	RevokeAuthorization(c context.Context, user *model.User) error

	// GetPerson retrieves metadata information about a user with the remote system.
	GetPerson(c context.Context, user *model.User, login string) (*model.Person, error)

	// GetOrgs gets a organization list from the remote system.
	GetOrgs(context.Context, *model.User) ([]*model.GitHubOrg, error)

	// ListTeams gets the repo's list of team names
	ListTeams(c context.Context, user *model.User, org string) (set.Set, error)

	// GetOrgMembers gets an organization member list from the remote system.
	GetOrgMembers(context.Context, *model.User, string) (set.Set, error)

	// GetTeamMembers gets an org's team members list from the remote system.
	GetTeamMembers(c context.Context, user *model.User, org string, team string) (set.Set, error)

	// GetCollaborators gets a collaborators list from the remote system.
	GetCollaborators(context.Context, *model.User, string, string) (set.Set, error)

	// GetRepo gets a repository from the remote system.
	GetRepo(context.Context, *model.User, string, string) (*model.Repo, error)

	// GetPerm gets a repository permission from the remote system.
	GetPerm(context.Context, *model.User, string, string) (*model.Perm, error)

	// SetHook adds a webhook to the remote repository.
	SetHook(context.Context, *model.User, *model.Repo, string) error

	// DelHook deletes a webhook from the remote repository.
	DelHook(context.Context, *model.User, *model.Repo, string) error

	// GetAllComments gets pull request comments from the remote system.
	GetAllComments(context.Context, *model.User, *model.Repo, int) ([]*model.Comment, error)

	// GetCommentsSinceHead gets pull request comments from the remote system since the head commit was committed.
	GetCommentsSinceHead(context.Context, *model.User, *model.Repo, int, bool) ([]*model.Comment, error)

	// GetAllReviews gets pull request reviews from the remote system.
	GetAllReviews(context.Context, *model.User, *model.Repo, int) ([]*model.Review, error)

	// GetReviewsSinceHead gets pull request reviews from the remote system since the head commit was committed.
	GetReviewsSinceHead(context.Context, *model.User, *model.Repo, int, bool) ([]*model.Review, error)

	// IsHeadUIMerge tests whether the HEAD of the pull request is a user interface merge.
	IsHeadUIMerge(c context.Context, u *model.User, r *model.Repo, num int) (bool, error)

	// CreateCompareURL creates a URL that prepares a diff of two commits
	CreateURLCompare(c context.Context, u *model.User, r *model.Repo, sha1, sha2 string) string

	// GetCommits gets one page of git commits
	GetCommits(context.Context, *model.User, *model.Repo, string, int, int) ([]string, int, error)

	// GetContents gets the file contents from the remote system.
	GetContents(context.Context, *model.User, *model.Repo, string) ([]byte, error)

	// GetStatus gets the commit statuses in the remote system.
	GetStatus(c context.Context, u *model.User, r *model.Repo, sha string) (model.CombinedStatus, error)

	// HasRequiredStatus tests whether the required commit statuses are passing.
	HasRequiredStatus(c context.Context, u *model.User, r *model.Repo, branch, sha string) (bool, error)

	// SetStatus adds or updates the commit status in the remote system.
	SetStatus(c context.Context, u *model.User, r *model.Repo, sha, context, status, desc string) error

	// CreateEmptyCommit creates an empty commit from the provided parent sha.
	CreateEmptyCommit(c context.Context, u *model.User, r *model.Repo, sha, msg string) (string, error)

	// CreateReference creates a reference pointing to the provided commit
	CreateReference(ctx context.Context, u *model.User, r *model.Repo, sha, name string) (string, error)

	// CreatePR creates a new pull request
	CreatePR(c context.Context, u *model.User, r *model.Repo, title, head, base, body string) (int, error)

	// MergePR merges the named pull request from the remote system
	MergePR(c context.Context, u *model.User, r *model.Repo, pullRequest model.PullRequest, approvers []*model.Person, message string, mergeMethod string) (string, error)

	// CompareBranches compares two branches for changes
	CompareBranches(c context.Context, u *model.User, repo *model.Repo, base string, head string, owner string) (model.BranchCompare, error)

	// DeleteBranch deletes a branch with the given reference
	DeleteBranch(c context.Context, u *model.User, repo *model.Repo, ref string) error

	// GetMaxExistingTag finds the highest tag across all tags
	ListTags(c context.Context, u *model.User, r *model.Repo) ([]model.Tag, error)

	// Tag applies a tag with the specified string to the specified sha
	Tag(c context.Context, u *model.User, r *model.Repo, tag string, sha string) error

	// GetPullRequest returns the pull request associated with a pull request number
	GetPullRequest(c context.Context, u *model.User, r *model.Repo, number int) (model.PullRequest, error)

	// GetPullRequestFiles returns the changed files associated with a pull request number
	GetPullRequestFiles(c context.Context, u *model.User, r *model.Repo, number int) ([]model.CommitFile, error)

	// GetPullRequestsForCommit returns all pull requests associated with a commit SHA
	GetPullRequestsForCommit(c context.Context, u *model.User, r *model.Repo, sha *string) ([]model.PullRequest, error)

	// GetIssue returns the issue associated with a issue number
	GetIssue(c context.Context, u *model.User, r *model.Repo, number int) (model.Issue, error)

	// WriteComment puts a new comment into the PR
	WriteComment(c context.Context, u *model.User, r *model.Repo, num int, message string) error

	ScheduleDeployment(c context.Context, u *model.User, r *model.Repo, d model.DeploymentInfo) error

	GetOrgPerm(c context.Context, user *model.User, owner string) (*model.Perm, error)

	// SetOrgHook adds a webhook to the remote organization.
	SetOrgHook(context.Context, *model.User, *model.OrgDb, string) error

	// DelOrgHook deletes a webhook from the remote organization.
	DelOrgHook(context.Context, *model.User, *model.OrgDb, string) error

	GetOrg(c context.Context, user *model.User, owner string) (*model.OrgDb, error)

	// GetOrgRepos gets the organization repository list from the remote system.
	GetOrgRepos(c context.Context, u *model.User, owner string) ([]*model.Repo, error)

	// GetUserRepos gets the user repository list from the remote system.
	GetUserRepos(context.Context, *model.User) ([]*model.Repo, error)
}

// Capabilities generates the user's capabilities
func Capabilities(c context.Context, u *model.User) (*model.Capabilities, error) {
	return FromContext(c).Capabilities(c, u)
}

// GetUser authenticates a user with the remote system.
func GetUser(c context.Context, w http.ResponseWriter, r *http.Request) (*model.User, error) {
	return FromContext(c).GetUser(c, w, r)
}

// GetUserToken authenticates a user with the remote system using
// the remote systems OAuth token.
func GetUserToken(c context.Context, token string) (string, error) {
	return FromContext(c).GetUserToken(c, token)
}

// RevokeAuthorization revokes the OAuth token
func RevokeAuthorization(c context.Context, user *model.User) error {
	return FromContext(c).RevokeAuthorization(c, user)
}

// GetPerson retrieves metadata information about a user with the remote system.
func GetPerson(c context.Context, user *model.User, login string) (*model.Person, error) {
	return FromContext(c).GetPerson(c, user, login)
}

// GetRepo gets a repository from the remote system.
func GetRepo(c context.Context, u *model.User, owner, name string) (*model.Repo, error) {
	return FromContext(c).GetRepo(c, u, owner, name)
}

// GetAllComments gets pull request comments from the remote system.
func GetAllComments(c context.Context, u *model.User, r *model.Repo, num int) ([]*model.Comment, error) {
	return FromContext(c).GetAllComments(c, u, r, num)
}

// IsHeadUIMerge tests whether the HEAD of the pull request is a user interface merge.
func IsHeadUIMerge(c context.Context, u *model.User, r *model.Repo, num int) (bool, error) {
	return FromContext(c).IsHeadUIMerge(c, u, r, num)
}

// GetCommentsSinceHead gets pull request comments from the remote system since the head commit was committed
func GetCommentsSinceHead(c context.Context, u *model.User, r *model.Repo, num int, noUIMerge bool) ([]*model.Comment, error) {
	return FromContext(c).GetCommentsSinceHead(c, u, r, num, noUIMerge)
}

// GetAllReviews gets pull request reviews from the remote system.
func GetAllReviews(c context.Context, u *model.User, r *model.Repo, num int) ([]*model.Review, error) {
	return FromContext(c).GetAllReviews(c, u, r, num)
}

// GetReviewsSinceHead gets pull request reviews from the remote system since the head commit was committed
func GetReviewsSinceHead(c context.Context, u *model.User, r *model.Repo, num int, noUIMerge bool) ([]*model.Review, error) {
	return FromContext(c).GetReviewsSinceHead(c, u, r, num, noUIMerge)
}

// CreateURLCompare creates a URL that prepares a diff of two commits
func CreateURLCompare(c context.Context, u *model.User, r *model.Repo, sha1, sha2 string) string {
	return FromContext(c).CreateURLCompare(c, u, r, sha1, sha2)
}

// GetCommits gets one page of git commits
func GetCommits(c context.Context, u *model.User, r *model.Repo, sha string, page, perPage int) ([]string, int, error) {
	return FromContext(c).GetCommits(c, u, r, sha, page, perPage)
}

// GetContents gets the file contents from the remote system.
func GetContents(c context.Context, u *model.User, r *model.Repo, path string) ([]byte, error) {
	return FromContext(c).GetContents(c, u, r, path)
}

// SetHook adds a webhook to the remote repository.
func SetHook(c context.Context, u *model.User, r *model.Repo, hook string) error {
	return FromContext(c).SetHook(c, u, r, hook)
}

// DelHook deletes a webhook from the remote repository.
func DelHook(c context.Context, u *model.User, r *model.Repo, hook string) error {
	return FromContext(c).DelHook(c, u, r, hook)
}

// GetStatus gets the commit statuses in the remote system.
func GetStatus(c context.Context, u *model.User, r *model.Repo, sha string) (model.CombinedStatus, error) {
	return FromContext(c).GetStatus(c, u, r, sha)
}

// SetStatus adds or updates the commit status in the remote system.
func SetStatus(c context.Context, u *model.User, r *model.Repo, sha, context, status, desc string) error {
	return FromContext(c).SetStatus(c, u, r, sha, context, status, desc)
}

// HasRequiredStatus tests whether the required commit statuses are passing.
func HasRequiredStatus(c context.Context, u *model.User, r *model.Repo, branch, sha string) (bool, error) {
	return FromContext(c).HasRequiredStatus(c, u, r, branch, sha)
}

// CreateEmptyCommit creates an empty commit from the provided parent sha.
func CreateEmptyCommit(c context.Context, u *model.User, r *model.Repo, sha, msg string) (string, error) {
	return FromContext(c).CreateEmptyCommit(c, u, r, sha, msg)
}

// CreateReference creates a reference pointing to the provided commit
func CreateReference(c context.Context, u *model.User, r *model.Repo, sha, name string) (string, error) {
	return FromContext(c).CreateReference(c, u, r, sha, name)
}

// CreatePR creates a new pull request
func CreatePR(c context.Context, u *model.User, r *model.Repo, title, head, base, body string) (int, error) {
	return FromContext(c).CreatePR(c, u, r, title, head, base, body)
}

func MergePR(c context.Context, u *model.User, r *model.Repo, pullRequest model.PullRequest, approvers []*model.Person, message string, mergeMethod string) (string, error) {
	return FromContext(c).MergePR(c, u, r, pullRequest, approvers, message, mergeMethod)
}

func CompareBranches(c context.Context, u *model.User, r *model.Repo, ref1 string, ref2 string, owner string) (model.BranchCompare, error) {
	return FromContext(c).CompareBranches(c, u, r, ref1, ref2, owner)
}

func DeleteBranch(c context.Context, u *model.User, repo *model.Repo, ref string) error {
	return FromContext(c).DeleteBranch(c, u, repo, ref)
}

func ListTags(c context.Context, u *model.User, r *model.Repo) ([]model.Tag, error) {
	return FromContext(c).ListTags(c, u, r)
}

func Tag(c context.Context, u *model.User, r *model.Repo, tag string, sha string) error {
	return FromContext(c).Tag(c, u, r, tag, sha)
}

func GetPullRequest(c context.Context, u *model.User, r *model.Repo, number int) (model.PullRequest, error) {
	return FromContext(c).GetPullRequest(c, u, r, number)
}

func GetPullRequestFiles(c context.Context, u *model.User, r *model.Repo, number int) ([]model.CommitFile, error) {
	return FromContext(c).GetPullRequestFiles(c, u, r, number)
}

func GetPullRequestsForCommit(c context.Context, u *model.User, r *model.Repo, sha *string) ([]model.PullRequest, error) {
	return FromContext(c).GetPullRequestsForCommit(c, u, r, sha)
}

// GetIssue returns the issue associated with a issue number
func GetIssue(c context.Context, u *model.User, r *model.Repo, number int) (model.Issue, error) {
	return FromContext(c).GetIssue(c, u, r, number)
}

func WriteComment(c context.Context, u *model.User, r *model.Repo, num int, message string) error {
	return FromContext(c).WriteComment(c, u, r, num, message)
}

func ScheduleDeployment(c context.Context, u *model.User, r *model.Repo, d model.DeploymentInfo) error {
	return FromContext(c).ScheduleDeployment(c, u, r, d)
}

func GetOrgPerm(c context.Context, user *model.User, owner string) (*model.Perm, error) {
	return FromContext(c).GetOrgPerm(c, user, owner)
}

// SetOrgHook adds a webhook to the remote organization.
func SetOrgHook(c context.Context, u *model.User, o *model.OrgDb, hook string) error {
	return FromContext(c).SetOrgHook(c, u, o, hook)
}

// DelOrgHook deletes a webhook from the remote organization.
func DelOrgHook(c context.Context, u *model.User, o *model.OrgDb, hook string) error {
	return FromContext(c).DelOrgHook(c, u, o, hook)
}

func GetOrg(c context.Context, user *model.User, owner string) (*model.OrgDb, error) {
	return FromContext(c).GetOrg(c, user, owner)
}

func GetUserRepos(c context.Context, user *model.User) ([]*model.Repo, error) {
	return FromContext(c).GetUserRepos(c, user)
}

func GetOrgRepos(c context.Context, user *model.User, owner string) ([]*model.Repo, error) {
	return FromContext(c).GetOrgRepos(c, user, owner)
}

var once sync.Once
var cachedRemote Remote

//todo use some more dynamic way to register which git repository we are interacting with. For now,
//todo we only have github support, but might add support for other repo types at some point.
func Get() Remote {
	once.Do(func() {
		cachedRemote = github.Get()
	})
	return cachedRemote
}
