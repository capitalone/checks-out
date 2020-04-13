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
package github

import (
	"context"
	"net/http"
	"net/url"
	"runtime"
	"strings"

	"github.com/capitalone/checks-out/exterror"
	"github.com/capitalone/checks-out/model"
	"github.com/capitalone/checks-out/usage"

	"github.com/google/go-github/v30/github"
	"golang.org/x/oauth2"
)

const (
	gitHubSubstring = "github.com/google/go-github/"
	gitHubCaller    = "github.com/google/go-github/github.(*Client).Do"
)

type UserTransport struct {
	Login string
	Event string
	// Transport is the underlying HTTP transport to use when making requests.
	// It will default to http.DefaultTransport if nil.
	Transport http.RoundTripper
}

// RoundTrip implements the RoundTripper interface.
func (t *UserTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	caller := t.caller()
	usage.RecordApiRequest(t.Login, t.Event, caller)
	return t.transport().RoundTrip(req)
}

func locateParent(frames *runtime.Frames) bool {
	for {
		frame, more := frames.Next()
		if strings.HasSuffix(frame.Function, gitHubCaller) {
			return true
		}
		if !more {
			return false
		}
	}
}

func (t *UserTransport) caller() string {
	pc := make([]uintptr, 16)
	n := runtime.Callers(2, pc)
	pc = pc[:n]
	frames := runtime.CallersFrames(pc)
	success := locateParent(frames)
	if success {
		frame, _ := frames.Next()
		idx := strings.Index(frame.Function, gitHubSubstring)
		if idx >= 0 {
			return frame.Function[idx+len(gitHubSubstring):]
		}
		return frame.Function
	}
	return ""
}

func (t *UserTransport) transport() http.RoundTripper {
	if t.Transport != nil {
		return t.Transport
	}
	return http.DefaultTransport
}

func setupClient(ctx context.Context, rawurl string, user *model.User) *github.Client {
	return createClient(ctx, rawurl, user.Token, user.Login)
}

func anonymousClient(ctx context.Context, rawurl, accessToken string) *github.Client {
	return createClient(ctx, rawurl, accessToken, "")
}

func createClient(ctx context.Context, rawurl, accessToken, login string) *github.Client {
	token := oauth2.Token{AccessToken: accessToken}
	source := oauth2.StaticTokenSource(&token)
	client := oauth2.NewClient(context.Background(), source)
	client.Transport = &UserTransport{
		Login:     login,
		Event:     usage.GetEventFromContext(ctx),
		Transport: client.Transport}
	g := github.NewClient(client)
	g.BaseURL, _ = url.Parse(rawurl)
	return g
}

func basicAuthClient(rawurl, id, secret string) *github.Client {
	transport := github.BasicAuthTransport{
		Username: id,
		Password: secret,
	}
	g := github.NewClient(transport.Client())
	g.BaseURL, _ = url.Parse(rawurl)
	return g
}

// getHook is a helper function that retrieves a hook by
// hostname. To do this, it will retrieve a list of all hooks
// and iterate through the list.
func getHook(ctx context.Context, client *github.Client, owner, name, rawurl string) (*github.Hook, error) {
	hooks, resp, err := client.Repositories.ListHooks(ctx, owner, name, nil)
	if err != nil {
		return nil, exterror.Create(resp.StatusCode, err)
	}
	newurl, err := url.Parse(rawurl)
	if err != nil {
		return nil, err
	}
	for _, hook := range hooks {
		hookurl, ok := hook.Config["url"].(string)
		if !ok {
			continue
		}
		oldurl, err := url.Parse(hookurl)
		if err != nil {
			continue
		}
		if newurl.Host == oldurl.Host {
			return hook, nil
		}
	}
	return nil, nil
}

// getOrgHook is a helper function that retrieves a hook by
// hostname. To do this, it will retrieve a list of all hooks
// and iterate through the list.
func getOrgHook(ctx context.Context, client *github.Client, owner, rawurl string) (*github.Hook, error) {
	hooks, resp, err := client.Organizations.ListHooks(ctx, owner, nil)
	if err != nil {
		return nil, exterror.Create(resp.StatusCode, err)
	}
	newurl, err := url.Parse(rawurl)
	if err != nil {
		return nil, err
	}
	for _, hook := range hooks {
		hookurl, ok := hook.Config["url"].(string)
		if !ok {
			continue
		}
		oldurl, err := url.Parse(hookurl)
		if err != nil {
			continue
		}
		if newurl.Host == oldurl.Host {
			return hook, nil
		}
	}
	return nil, nil
}

// createHook is a helper function that creates a post-commit hook
// for the specified repository.
func createHook(ctx context.Context, client *github.Client, owner, name, url string) (*github.Hook, error) {
	var hook = new(github.Hook)
	hook.Events = []string{"issue_comment", "status", "pull_request", "pull_request_review"}
	hook.Config = map[string]interface{}{}
	hook.Config["url"] = url
	hook.Config["content_type"] = "json"
	created, resp, err := client.Repositories.CreateHook(ctx, owner, name, hook)
	if err != nil {
		err = exterror.Create(resp.StatusCode, err)
	}
	return created, err
}

// createOrgHook is a helper function that creates a post-commit hook
// for the specified Organization.
func createOrgHook(ctx context.Context, client *github.Client, owner, url string) (*github.Hook, error) {
	var hook = new(github.Hook)
	hook.Events = []string{"repository"}
	hook.Config = map[string]interface{}{}
	hook.Config["url"] = url
	hook.Config["content_type"] = "json"
	created, resp, err := client.Organizations.CreateHook(ctx, owner, hook)
	if err != nil {
		err = exterror.Create(resp.StatusCode, err)
	}
	return created, err
}

// getUserRepos is a helper function that returns a list of
// all user repositories. Paginated results are aggregated into
// a single list.
func getUserRepos(ctx context.Context, client *github.Client, user string) ([]*github.Repository, error) {
	var repos []*github.Repository
	var loOpts = github.RepositoryListOptions{}
	resp, err := buildCompleteList(func(opts *github.ListOptions) (*github.Response, error) {
		loOpts.ListOptions = *opts
		next, resp, err := client.Repositories.List(ctx, user, &loOpts)
		repos = append(repos, next...)
		return resp, err
	})

	if err != nil {
		return nil, exterror.Create(resp.StatusCode, err)
	}

	return repos, nil
}

// getOrgRepos is a helper function that returns a list of
// all user repositories. Paginated results are aggregated into
// a single list.
func getOrgRepos(ctx context.Context, client *github.Client, org string) ([]*github.Repository, error) {
	var repos []*github.Repository
	var loOpts = github.RepositoryListByOrgOptions{}
	resp, err := buildCompleteList(func(opts *github.ListOptions) (*github.Response, error) {
		loOpts.ListOptions = *opts
		next, resp, err := client.Repositories.ListByOrg(ctx, org, &loOpts)
		repos = append(repos, next...)
		return resp, err
	})

	if err != nil {
		return nil, exterror.Create(resp.StatusCode, err)
	}

	return repos, nil
}
