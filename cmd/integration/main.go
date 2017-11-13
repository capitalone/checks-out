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
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/capitalone/checks-out/envvars"
	gh "github.com/capitalone/checks-out/remote/github"

	log "github.com/Sirupsen/logrus"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

var (
	configFileName = fmt.Sprintf(".%s", envvars.Env.Branding.Name)
	addr           = envvars.Env.Server.Addr
	githubToken    = envvars.Env.Test.GithubToken
	serviceToken   string
	githubAPI      string
	ctx            context.Context
)

var configText = `
approvals:
[
  {
    name: "master"
    match: "universe[count=1,self=true]"
  }
]
merge:
{
  enable: true
}
`

func init() {
	ctx = context.Background()
	githubParams := gh.Get()
	githubAPI = githubParams.API
	if len(addr) == 0 {
		log.Fatal("SERVER_ADDR environment variable must be defined")
	}
	if len(githubToken) == 0 {
		log.Fatal("GITHUB_TEST_TOKEN environment variable must be defined")
	}
}

func createGitHubClient() *github.Client {
	var err error
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: githubToken},
	)
	tc := oauth2.NewClient(oauth2.NoContext, ts)
	client := github.NewClient(tc)
	client.BaseURL, err = url.Parse(githubAPI)
	if err != nil {
		log.Fatalf("Unable to parse url '%s': %s", githubAPI, err.Error())
	}
	return client
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

func createRepo(client *github.Client) *github.Repository {
	name := randomString(16)
	repo := &github.Repository{
		Name:     github.String(name),
		Private:  github.Bool(false),
		AutoInit: github.Bool(true),
	}
	repo, _, err := client.Repositories.Create(ctx, "", repo)
	if err != nil {
		log.Error("Unable to create git repository: ", err)
		return nil
	}
	return repo
}

func createCommit(client *github.Client, repo *github.Repository, branch *github.Reference, filename string, contents string) *github.Reference {
	blob, _, err := client.Git.CreateBlob(ctx, *repo.Owner.Login, *repo.Name, &github.Blob{
		Content:  github.String(contents),
		Size:     github.Int(len(contents)),
		Encoding: github.String("utf-8"),
	})
	if err != nil {
		log.Error("Unable to create blob ", filename, err)
		return nil
	}
	tree, _, err := client.Git.CreateTree(ctx, *repo.Owner.Login, *repo.Name, *branch.Object.SHA, []github.TreeEntry{{
		Path: github.String(filename),
		Mode: github.String("100644"),
		Type: github.String("blob"),
		SHA:  blob.SHA,
	}})
	if err != nil {
		log.Error("Unable to get create tree ", filename, err)
		return nil
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
		log.Error("Unable to get create commit ", filename, err)
		return nil
	}
	branch.Object.SHA = commit.SHA
	branch, _, err = client.Git.UpdateRef(ctx, *repo.Owner.Login, *repo.Name, branch, false)
	if err != nil {
		log.Error("Unable to update reference ", filename, err)
		return nil
	}
	return branch
}

func initServiceToken() bool {
	route := fmt.Sprintf("/login?access_token=%s", githubToken)
	resp, err := http.Post(addr+route, "", nil)
	if err != nil {
		log.Error("Unable to send POST /login to server ", err)
		return false
	}
	if resp.StatusCode != 200 {
		log.Errorf("POST /login to server did not return 200: %v", resp)
		return false
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error("Unable to read POST /login bytes from server ", err)
		return false
	}
	var token struct {
		Access string `json:"access_token"`
	}
	err = json.NewDecoder(bytes.NewReader(body)).Decode(&token)
	if err != nil {
		log.Error("Unable to read POST /login response from server ", err)
		return false
	}
	serviceToken = token.Access
	return true
}

func registerRepo(client *http.Client, repo *github.Repository) bool {
	route := fmt.Sprintf("/api/repos/%s/%s", *repo.Owner.Login, *repo.Name)
	req, err := http.NewRequest("POST", addr+route, nil)
	if err != nil {
		log.Error("Unable to create POST /api/repos route to server ", err)
		return false
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", serviceToken))
	resp, err := client.Do(req)
	if err != nil {
		log.Error("Unable to send POST /api/repos to server ", err)
		return false
	}
	if resp.StatusCode != 200 {
		log.Errorf("POST /api/repos to server did not return 200: %v", resp)
		return false
	}
	return true
}

func validateRepo(client *http.Client, repo *github.Repository) bool {
	route := fmt.Sprintf("/api/repos/%s/%s/validate", *repo.Owner.Login, *repo.Name)
	req, err := http.NewRequest("GET", addr+route, nil)
	if err != nil {
		log.Error("Unable to create GET /api/repos/validate route to server ", err)
		return false
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", serviceToken))
	resp, err := client.Do(req)
	if err != nil {
		log.Error("Unable to send GET /api/repos/validate to server ", err)
		return false
	}
	if resp.StatusCode != 204 {
		log.Errorf("GET /api/repos/validate to server did not return 204: %v", resp)
		return false
	}
	return true
}

func createPR(client *github.Client, repo *github.Repository) bool {
	ref, _, err := client.Git.GetRef(ctx, *repo.Owner.Login, *repo.Name, "refs/heads/master")
	if err != nil {
		log.Error("Unable to get master head ", err)
		return false
	}
	branch, _, err := client.Git.CreateRef(ctx, *repo.Owner.Login, *repo.Name, &github.Reference{
		Ref:    github.String("refs/heads/foobar"),
		Object: ref.Object,
	})
	if err != nil {
		log.Error("Unable to create branch ", err)
		return false
	}
	if createCommit(client, repo, branch, "foobar", "") == nil {
		return false
	}
	_, _, err = client.PullRequests.Create(ctx, *repo.Owner.Login, *repo.Name, &github.NewPullRequest{
		Title: github.String("Adds foobar feature"),
		Head:  github.String("foobar"),
		Base:  github.String("master"),
	})
	if err != nil {
		log.Error("Unable to create pull request ", err)
		return false
	}
	return true
}

func statusPR(client *http.Client, repo *github.Repository, expect bool) bool {
	route := fmt.Sprintf("/api/pr/%s/%s/1/status", *repo.Owner.Login, *repo.Name)
	req, err := http.NewRequest("GET", addr+route, nil)
	if err != nil {
		log.Error("Unable to create GET /api/pr/repos/status route to server ", err)
		return false
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", serviceToken))
	resp, err := client.Do(req)
	if err != nil {
		log.Error("Unable to send GET /api/pr/repos/status to server ", err)
		return false
	}
	if resp.StatusCode != 200 {
		log.Errorf("GET /api/pr/repos/status to server did not return 200: %v", resp)
		return false
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error("Unable to read GET /api/pr/repos/status bytes from server ", err)
		return false
	}
	var status struct {
		Approved bool `json:"approved"`
	}
	err = json.NewDecoder(bytes.NewReader(body)).Decode(&status)
	if err != nil {
		log.Error("Unable to read GET /api/pr/repos/status response from server ", err)
		return false
	}
	if status.Approved != expect {
		log.Errorf("pull request expected %t status received %t", expect, status.Approved)
		return false
	}
	return true
}

func addComment(client *github.Client, repo *github.Repository) bool {
	_, _, err := client.Issues.CreateComment(ctx, *repo.Owner.Login, *repo.Name, 1, &github.IssueComment{
		Body: github.String("I approve"),
	})
	if err != nil {
		log.Error("Unable to create GitHub comment ", err)
		return false
	}
	return true
}

func unregisterRepo(webclient *http.Client, repo *github.Repository) {
	if len(serviceToken) == 0 {
		return
	}
	route := fmt.Sprintf("/api/repos/%s/%s", *repo.Owner.Login, *repo.Name)
	req, err := http.NewRequest("DELETE", addr+route, nil)
	if err != nil {
		log.Error("Unable to create DELETE /api/repos route to server ", err)
		return
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", serviceToken))
	resp, err := webclient.Do(req)
	if err != nil {
		log.Error("Unable to send DELETE /api/repos to server ", err)
		return
	}
	if resp.StatusCode != 200 {
		log.Errorf("DELETE /api/repos to server did not return 200: %v", resp)
		return
	}
}

func deleteRepo(ghclient *github.Client, repo *github.Repository) {
	_, err := ghclient.Repositories.Delete(ctx, *repo.Owner.Login, *repo.Name)
	if err != nil {
		log.Error("Unable to delete git repository ", err)
		return
	}
}

func cleanupTests(ghclient *github.Client, webclient *http.Client, repo *github.Repository) {
	unregisterRepo(webclient, repo)
	deleteRepo(ghclient, repo)
}

func runTests() int {
	ghclient := createGitHubClient()
	webclient := &http.Client{}
	repo := createRepo(ghclient)
	if repo == nil {
		return 1
	}
	defer cleanupTests(ghclient, webclient, repo)
	ref, _, err := ghclient.Git.GetRef(ctx, *repo.Owner.Login, *repo.Name, "refs/heads/master")
	if err != nil {
		log.Error("Unable to get master head ", err)
		return 1
	}
	ref = createCommit(ghclient, repo, ref, configFileName, configText)
	if ref == nil {
		return 1
	}
	ref = createCommit(ghclient, repo, ref, "MAINTAINERS", "")
	if ref == nil {
		return 1
	}
	if !initServiceToken() {
		return 1
	}
	if !registerRepo(webclient, repo) {
		return 1
	}
	if !validateRepo(webclient, repo) {
		return 1
	}
	if !createPR(ghclient, repo) {
		return 1
	}
	if !statusPR(webclient, repo, false) {
		return 1
	}
	if !addComment(ghclient, repo) {
		return 1
	}
	if !statusPR(webclient, repo, true) {
		return 1
	}
	log.Info("All tests passed.")
	return 0
}

func main() {
	os.Exit(runTests())
}
