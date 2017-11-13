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
	"fmt"
	"math/rand"
	"net/url"
	"testing"
	"time"

	"github.com/google/go-github/github"
	"github.com/capitalone/checks-out/envvars"
	"github.com/capitalone/checks-out/exterror"
	"github.com/capitalone/checks-out/model"
	"golang.org/x/oauth2"
)

func createTestClient(t *testing.T) *github.Client {
	var err error
	githubParams := Get()
	urlstring := githubParams.API
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: envvars.Env.Test.GithubToken},
	)
	tc := oauth2.NewClient(oauth2.NoContext, ts)
	client := github.NewClient(tc)
	client.BaseURL, err = url.Parse(urlstring)
	if err != nil {
		t.Fatalf("Unable to parse url '%s': %s", urlstring, err.Error())
	}
	return client
}

func TestGitHubClient(t *testing.T) {
	if !envvars.Env.Test.GithubTestEnable {
		t.Log("Skipping GitHub integration tests")
		return
	}
	if len(envvars.Env.Test.GithubToken) == 0 {
		t.Fatal("GITHUB_TEST_TOKEN environment variable must be defined")
	}
	if len(envvars.Env.Github.Email) == 0 {
		t.Fatal("GITHUB_EMAIL environment variable must be defined")
	}
	client := createTestClient(t)
	repo := createRepo(t, client)
	branch := createTestBranch(t, client, repo)
	branch = createTestCommit(t, client, repo, branch, "foo")
	pr := createTestPR(t, client, repo)
	createPRComment(t, client, repo, pr, "this is a comment")
	branch = createTestCommit(t, client, repo, branch, "bar")
	commit := getCommit(t, client, repo, branch)
	createPRComment(t, client, repo, pr, "this is another comment")
	createPRReview(t, client, repo, pr, "this is a review")
	t.Run("GetRepo", func(t *testing.T) { testGetRepo(t, client, repo) })
	t.Run("GetComments", func(t *testing.T) { testGetComments(t, client, repo, pr) })
	t.Run("GetReviews", func(t *testing.T) { testGetReviews(t, client, repo, pr) })
	t.Run("GetOrgs", func(t *testing.T) { testGetOrgs(t, client) })
	t.Run("GetModelRepo", func(t *testing.T) { testGetModelRepo(t, client, repo) })
	t.Run("GetRepos", func(t *testing.T) { testGetUserRepos(t, client, repo) })
	t.Run("GetContents", func(t *testing.T) { testGetContents(t, client, repo) })
	t.Run("GetPullRequest", func(t *testing.T) { testGetPullRequest(t, client, repo, pr) })
	t.Run("CompareBranches", func(t *testing.T) { testCompareBranches(t, client, repo, pr) })
	t.Run("GetPullRequestFiles", func(t *testing.T) { testGetPullRequestFiles(t, client, repo, pr) })
	t.Run("GetAllComments", func(t *testing.T) { testGetAllComments(t, client, repo, pr) })
	t.Run("Tags", func(t *testing.T) { testTags(t, client, repo, branch) })
	t.Run("GetPullRequestForCommit", func(t *testing.T) { testGetPullRequestForCommit(t, client, repo, commit) })
	t.Run("MergeAndDelete", func(t *testing.T) { testMergeAndDelete(t, client, repo, pr) })
	deleteRepo(t, client, repo)
}

func TestGitHubMergeSquash(t *testing.T) {
	if !envvars.Env.Test.GithubTestEnable {
		t.Log("Skipping GitHub integration tests")
		return
	}
	if len(envvars.Env.Test.GithubToken) == 0 {
		t.Fatal("GITHUB_TEST_TOKEN environment variable must be defined")
	}
	client := createTestClient(t)
	repo := createRepo(t, client)
	branch := createTestBranch(t, client, repo)
	branch = createTestCommit(t, client, repo, branch, "foo")
	pr := createTestPR(t, client, repo)
	createPRComment(t, client, repo, pr, "this is a comment")
	branch = createTestCommit(t, client, repo, branch, "bar")
	createPRComment(t, client, repo, pr, "this is another comment")
	createPRReview(t, client, repo, pr, "this is a review")
	t.Run("SquashAndDelete", func(t *testing.T) { testSquashAndDelete(t, client, repo, pr) })
	deleteRepo(t, client, repo)
}

func TestGitHubMergeRebase(t *testing.T) {
	if !envvars.Env.Test.GithubTestEnable {
		t.Log("Skipping GitHub integration tests")
		return
	}
	if len(envvars.Env.Test.GithubToken) == 0 {
		t.Fatal("GITHUB_TEST_TOKEN environment variable must be defined")
	}
	client := createTestClient(t)
	repo := createRepo(t, client)
	branch := createTestBranch(t, client, repo)
	branch = createTestCommit(t, client, repo, branch, "foo")
	pr := createTestPR(t, client, repo)
	createPRComment(t, client, repo, pr, "this is a comment")
	branch = createTestCommit(t, client, repo, branch, "bar")
	createPRComment(t, client, repo, pr, "this is another comment")
	createPRReview(t, client, repo, pr, "this is a review")
	t.Run("RebaseAndDelete", func(t *testing.T) { testRebaseAndDelete(t, client, repo, pr) })
	deleteRepo(t, client, repo)
}

func testGetRepo(t *testing.T, client *github.Client, repo *github.Repository) {
	ctx := context.Background()
	_, _, err := client.Repositories.Get(ctx, *repo.Owner.Login, *repo.Name)
	if err != nil {
		t.Error("Unable to get repository", err)
	}
}

func testGetComments(t *testing.T, client *github.Client, repo *github.Repository, pr *github.PullRequest) {
	ctx := context.Background()
	opts := github.IssueListCommentsOptions{Direction: "desc", Sort: "created"}
	opts.PerPage = 100
	comm, _, err := client.Issues.ListComments(ctx, *repo.Owner.Login, *repo.Name, *pr.Number, &opts)
	if err != nil {
		t.Error("Unable to get comments", err)
	}
	if len(comm) != 2 {
		t.Error("Did not find 2 comments", len(comm))
	}
}

func testGetReviews(t *testing.T, client *github.Client, repo *github.Repository, pr *github.PullRequest) {
	ctx := context.Background()
	reviews, _, err := client.PullRequests.ListReviews(ctx, *repo.Owner.Login, *repo.Name, *pr.Number, nil)
	if err != nil {
		t.Error("Unable to get pull request reviews", err)
	}
	if len(reviews) != 1 {
		t.Error("Did not find 1 pull request review", len(reviews))
	}
}

func testGetOrgs(t *testing.T, client *github.Client) {
	ctx := context.Background()
	orgs, err := getOrgs(ctx, client)
	if err != nil {
		t.Error("Unable to get github organizations", err)
	}
	if orgs == nil {
		t.Error("organizations must be non-nil value")
	}
}

func testGetModelRepo(t *testing.T, client *github.Client, repo *github.Repository) {
	ctx := context.Background()
	res, err := getRepo(ctx, client, *repo.Owner.Login, *repo.Name)
	if err != nil {
		t.Error("Unable to get repository", err)
	}
	if res.Name != *repo.Name {
		t.Error("Repository name is incorrect", res.Name)
	}
}

func testGetUserRepos(t *testing.T, client *github.Client, repo *github.Repository) {
	ctx := context.Background()
	repos, err := getUserRepos(ctx, client, repo.Owner.GetLogin())
	if err != nil {
		t.Error("Unable to get user's repositories", err)
	}
	if repos == nil {
		t.Error("repos must be non-nil value")
	}
}

func testGetContents(t *testing.T, client *github.Client, repo *github.Repository) {
	ctx := context.Background()
	r := &model.Repo{Owner: *repo.Owner.Login, Name: *repo.Name}
	res, err := getContents(ctx, client, r, "README.md")
	if err != nil {
		t.Error("Unable to get README.md", err)
	}
	if len(res) == 0 {
		t.Error("Unable to get README.md contents", res)
	}
	res, err = getContents(ctx, client, r, "foobar.md")
	if err == nil {
		t.Error("Failed to produce error message")
	}
	if err.(exterror.ExtError).Status != 404 {
		t.Error("Error status is not a 404")
	}
}

func testGetPullRequest(t *testing.T, client *github.Client, repo *github.Repository, pr *github.PullRequest) {
	ctx := context.Background()
	r := &model.Repo{Owner: *repo.Owner.Login, Name: *repo.Name}
	res, err := getPullRequest(ctx, client, r, *pr.Number)
	if err != nil {
		t.Error("Unable to get pull request", err)
	}
	if res.Number != *pr.Number {
		t.Error("Pull request number is incorrect", res.Number)
	}
}

func testCompareBranches(t *testing.T, client *github.Client, repo *github.Repository, pr *github.PullRequest) {
	ctx := context.Background()
	r := &model.Repo{Owner: *repo.Owner.Login, Name: *repo.Name}
	res, err := compareBranches(ctx, client, r, pr.Base.GetRef(), pr.Head.GetRef(), pr.Head.User.GetLogin())
	if err != nil {
		t.Error("Unable to compare branches", err)
	}
	if res.AheadBy != 2 {
		t.Error("Ahead by number is incorrect", res.AheadBy)
	}
}

func testGetPullRequestFiles(t *testing.T, client *github.Client, repo *github.Repository, pr *github.PullRequest) {
	ctx := context.Background()
	r := &model.Repo{Owner: *repo.Owner.Login, Name: *repo.Name}
	res, err := getPullRequestFiles(ctx, client, r, *pr.Number)
	if err != nil {
		t.Error("Unable to get pull request files", err)
	}
	if len(res) == 0 {
		t.Error("Length of pull request files should be nonzero")
	}
}

func testGetAllComments(t *testing.T, client *github.Client, repo *github.Repository, pr *github.PullRequest) {
	ctx := context.Background()
	r := &model.Repo{Owner: *repo.Owner.Login, Name: *repo.Name}
	res, err := getAllComments(ctx, client, r, *pr.Number)
	if err != nil {
		t.Error("Unable to get all comments", err)
	}
	if len(res) != 2 {
		t.Error("Unable to find two comments", len(res))
	}
}

func testGetPullRequestForCommit(t *testing.T, client *github.Client, repo *github.Repository, commit *github.Commit) {
	ctx := context.Background()
	r := &model.Repo{Owner: *repo.Owner.Login, Name: *repo.Name}
	_, err := getPullRequestsForCommit(ctx, client, r, commit.SHA)
	if err != nil {
		t.Error("Unable to get pull requests for commit", err)
	}
}

func testMergeAndDelete(t *testing.T, client *github.Client, repo *github.Repository, pr *github.PullRequest) {
	testMergeAndDeleteInner(t, client, repo, pr, "merge")
}

func testSquashAndDelete(t *testing.T, client *github.Client, repo *github.Repository, pr *github.PullRequest) {
	testMergeAndDeleteInner(t, client, repo, pr, "squash")
}

func testRebaseAndDelete(t *testing.T, client *github.Client, repo *github.Repository, pr *github.PullRequest) {
	testMergeAndDeleteInner(t, client, repo, pr, "rebase")
}

func testMergeAndDeleteInner(t *testing.T, client *github.Client, repo *github.Repository, pr *github.PullRequest, mergeMethod string) {
	ctx := context.Background()
	r := &model.Repo{Owner: *repo.Owner.Login, Name: *repo.Name}
	p := model.PullRequest{Issue: model.Issue{Number: *pr.Number}}
	latest, _, err := client.PullRequests.Get(ctx, r.Owner, r.Name, p.Number)
	if err != nil {
		t.Error("Unable to get latest status of pull request", err)
	}
	if latest.Mergeable != nil && *latest.Mergeable {
		_, err = mergePR(ctx, client, r, p, []*model.Person{}, "", "")
		if err != nil && err.(exterror.ExtError).Status != 405 {
			t.Error("Unable to merge pull request", err)
		}
	}
	err = deleteBranch(ctx, client, r, "foobar")
	if err != nil {
		t.Error("Unable to delete branch", err)
	}
}

func testTags(t *testing.T, client *github.Client, repo *github.Repository, branch *github.Reference) {
	ctx := context.Background()
	r := &model.Repo{Owner: *repo.Owner.Login, Name: *repo.Name}
	tags, err := listTags(ctx, client, r)
	if err != nil {
		t.Error("Unable to list tags", err)
	}
	if len(tags) != 0 {
		t.Error("Expected zero tags", len(tags))
	}
	err = doTag(ctx, client, r, github.String("foo"), branch.Object.SHA)
	if err != nil {
		t.Error("Unable to create tags", err)
	}
}

func createRepo(t *testing.T, client *github.Client) *github.Repository {
	ctx := context.Background()
	name := randomString(16)
	repo := &github.Repository{
		Name:     github.String(name),
		Private:  github.Bool(false),
		AutoInit: github.Bool(true),
	}
	repo, _, err := client.Repositories.Create(ctx, "", repo)
	if err != nil {
		t.Fatal("Unable to create git repository", err)
	}
	return repo
}

func createTestBranch(t *testing.T, client *github.Client, repo *github.Repository) *github.Reference {
	ctx := context.Background()
	ref, _, err := client.Git.GetRef(ctx, *repo.Owner.Login, *repo.Name, "refs/heads/master")
	if err != nil {
		t.Fatal("Unable to get master head", err)
	}
	ref, _, err = client.Git.CreateRef(ctx, *repo.Owner.Login, *repo.Name, &github.Reference{
		Ref:    github.String("refs/heads/foobar"),
		Object: ref.Object,
	})
	if err != nil {
		t.Fatal("Unable to create branch", err)
	}
	return ref
}

func createTestCommit(t *testing.T, client *github.Client, repo *github.Repository, branch *github.Reference, filename string) *github.Reference {
	ctx := context.Background()
	blob, _, err := client.Git.CreateBlob(ctx, *repo.Owner.Login, *repo.Name, &github.Blob{
		Content:  github.String(""),
		Size:     github.Int(0),
		Encoding: github.String("utf-8"),
	})
	if err != nil {
		t.Fatal("Unable to get create blob", filename, err)
	}
	tree, _, err := client.Git.CreateTree(ctx, *repo.Owner.Login, *repo.Name, *branch.Object.SHA, []github.TreeEntry{{
		Path: github.String(filename),
		Mode: github.String("100644"),
		Type: github.String("blob"),
		SHA:  blob.SHA,
	}})
	if err != nil {
		t.Fatal("Unable to get create tree", filename, err)
	}
	commit, _, err := client.Git.CreateCommit(ctx, *repo.Owner.Login, *repo.Name, &github.Commit{
		Message: github.String(fmt.Sprintf("%s commit", filename)),
		Tree:    tree,
		Parents: []github.Commit{{
			SHA: branch.Object.SHA,
		},
		},
	})
	if err != nil {
		t.Fatal("Unable to get create commit", filename, err)
	}
	branch.Object.SHA = commit.SHA
	branch, _, err = client.Git.UpdateRef(ctx, *repo.Owner.Login, *repo.Name, branch, false)
	if err != nil {
		t.Fatal("Unable to update reference", filename, err)
	}
	return branch
}

func getCommit(t *testing.T, client *github.Client, repo *github.Repository, branch *github.Reference) *github.Commit {
	ctx := context.Background()
	commit, _, err := client.Git.GetCommit(ctx, *repo.Owner.Login, *repo.Name, *branch.Object.SHA)
	if err != nil {
		t.Fatal("Unable to get commit", *branch.Object.SHA, err)
	}
	return commit
}

func createTestPR(t *testing.T, client *github.Client, repo *github.Repository) *github.PullRequest {
	ctx := context.Background()
	pr, _, err := client.PullRequests.Create(ctx, *repo.Owner.Login, *repo.Name, &github.NewPullRequest{
		Title: github.String("Adds foobar feature"),
		Head:  github.String("foobar"),
		Base:  github.String("master"),
	})
	if err != nil {
		t.Fatal("Unable to create pull request", err)
	}
	return pr
}

func createPRComment(t *testing.T, client *github.Client,
	repo *github.Repository, pr *github.PullRequest, comment string) {
	ctx := context.Background()
	r := &model.Repo{Owner: *repo.Owner.Login, Name: *repo.Name}
	err := writeComment(ctx, client, r, *pr.Number, comment)
	if err != nil {
		t.Error("Unable to add comment", comment, err)
	}
}

func createPRReview(t *testing.T, client *github.Client,
	repo *github.Repository, pr *github.PullRequest, comment string) {
	ctx := context.Background()
	review := github.PullRequestReviewRequest{}
	review.Body = &comment
	_, resp, err := client.PullRequests.CreateReview(ctx, *repo.Owner.Login, *repo.Name, *pr.Number, &review)
	if err != nil {
		t.Error("Unable to add pull request review", err, resp)
	}
}

func deleteRepo(t *testing.T, client *github.Client, repo *github.Repository) {
	ctx := context.Background()
	_, err := client.Repositories.Delete(ctx, *repo.Owner.Login, *repo.Name)
	if err != nil {
		t.Error("Unable to delete git repository", err)
	}
}

func randomString(strlen int) string {
	rand.Seed(time.Now().UTC().UnixNano())
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, strlen)
	for i := 0; i < strlen; i++ {
		result[i] = chars[rand.Intn(len(chars))]
	}
	return string(result)
}
