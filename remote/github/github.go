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
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/capitalone/checks-out/envvars"
	"github.com/capitalone/checks-out/exterror"
	"github.com/capitalone/checks-out/model"
	"github.com/capitalone/checks-out/set"
	"github.com/capitalone/checks-out/shared/httputil"
	"github.com/capitalone/checks-out/strings/lowercase"

	log "github.com/Sirupsen/logrus"
	"github.com/google/go-github/github"
	multierror "github.com/mspiegel/go-multierror"
	"golang.org/x/oauth2"
)

const (
	DefaultURL   = "https://github.com"
	DefaultAPI   = "https://api.github.com/"
	DefaultScope = "read:org,repo:status"
)

func createErrorFallback(resp *github.Response, err error, fallback int) error {
	if resp != nil {
		return exterror.Create(resp.StatusCode, err)
	}
	return exterror.Create(fallback, err)
}

func createError(resp *github.Response, err error) error {
	return createErrorFallback(resp, err, http.StatusInternalServerError)
}

type Github struct {
	URL    string
	API    string
	Client string
	Secret string
}

func Get() *Github {
	remote := &Github{
		API:    DefaultAPI,
		URL:    envvars.Env.Github.Url,
		Client: envvars.Env.Github.Client,
		Secret: envvars.Env.Github.Secret,
	}
	if remote.URL != DefaultURL {
		remote.URL = strings.TrimSuffix(remote.URL, "/")
		remote.API = remote.URL + "/api/v3/"
	}
	return remote
}

func (g *Github) Capabilities(ctx context.Context, u *model.User) (*model.Capabilities, error) {
	var errs error
	s := set.New(strings.Split(u.Scopes, ",")...)
	caps := new(model.Capabilities)
	caps.Org.Read = s.Contains("read:org") || s.Contains("write:org") || s.Contains("admin:org")
	caps.Repo.CommitStatus = s.Contains("repo:status") || s.Contains("repo") || s.Contains("public_repo")
	caps.Repo.DeploymentStatus = s.Contains("repo_deployment") || s.Contains("repo") || s.Contains("public_repo")
	caps.Repo.DeleteBranch = s.Contains("repo") || s.Contains("public_repo")
	caps.Repo.Merge = s.Contains("repo") || s.Contains("public_repo")
	caps.Repo.Tag = s.Contains("repo") || s.Contains("public_repo")
	caps.Repo.PRWriteComment = s.Contains("repo") || s.Contains("public_repo")
	if !caps.Repo.CommitStatus {
		errs = multierror.Append(errs, errors.New("commit status OAuth scope is required"))
	}
	if errs != nil {
		return nil, exterror.Create(http.StatusUnauthorized, errs)
	}
	return caps, nil
}

func (g *Github) GetUser(ctx context.Context, res http.ResponseWriter, req *http.Request) (*model.User, error) {
	scopes := envvars.Env.Github.Scope
	var config = &oauth2.Config{
		ClientID:     g.Client,
		ClientSecret: g.Secret,
		RedirectURL:  fmt.Sprintf("%s/login", httputil.GetURL(req)),
		Endpoint: oauth2.Endpoint{
			AuthURL:  fmt.Sprintf("%s/login/oauth/authorize", g.URL),
			TokenURL: fmt.Sprintf("%s/login/oauth/access_token", g.URL),
		},
		Scopes: strings.Split(scopes, ","),
	}

	// get the oauth code from the incoming request. if no code is present
	// redirec the user to GitHub login to retrieve a code.
	var code = req.FormValue("code")
	if len(code) == 0 {
		state := fmt.Sprintln(time.Now().Unix())
		http.Redirect(res, req, config.AuthCodeURL(state), http.StatusSeeOther)
		return nil, nil
	}

	// exchanges the oauth2 code for an access token
	token, err := config.Exchange(oauth2.NoContext, code)
	if err != nil {
		err = fmt.Errorf("Exchanging token. %s", err)
		return nil, exterror.Create(http.StatusBadRequest, err)
	}

	// get the currently authenticated user details for the access token
	client := anonymousClient(ctx, g.API, token.AccessToken)
	user, resp, err := client.Users.Get(ctx, "")
	if err != nil {
		err = fmt.Errorf("Fetching user. %s", err)
		return nil, createError(resp, err)
	}

	// get the subset of requested scopes granted to the access token
	scopes, err = g.GetScopes(ctx, token.AccessToken)
	if err != nil {
		err = fmt.Errorf("Fetching user. %s", err)
		return nil, createError(resp, err)
	}

	return &model.User{
		Login:  user.GetLogin(),
		Token:  token.AccessToken,
		Avatar: user.GetAvatarURL(),
		Scopes: scopes,
	}, nil
}

func (g *Github) GetUserToken(ctx context.Context, token string) (string, error) {
	client := anonymousClient(ctx, g.API, token)
	user, resp, err := client.Users.Get(ctx, "")
	if err != nil {
		err = fmt.Errorf("Fetching user. %s", err)
		return "", createError(resp, err)
	}
	return user.GetLogin(), nil
}

func (g *Github) GetScopes(ctx context.Context, token string) (string, error) {
	client := basicAuthClient(g.API, g.Client, g.Secret)
	auth, resp, err := client.Authorizations.Check(ctx, g.Client, token)
	if err != nil {
		err = fmt.Errorf("Checking authorization. %s", err)
		return "", createError(resp, err)
	}
	scopes := set.Empty()
	for _, scope := range auth.Scopes {
		scopes.Add(string(scope))
	}
	return scopes.Print(","), nil
}

func (g *Github) RevokeAuthorization(ctx context.Context, user *model.User) error {
	client := basicAuthClient(g.API, g.Client, g.Secret)
	resp, err := client.Authorizations.Revoke(ctx, g.Client, user.Token)
	if err != nil {
		err = fmt.Errorf("Revoking authorization. %s", err)
		return createError(resp, err)
	}
	return nil
}

func (g *Github) GetOrgs(ctx context.Context, user *model.User) ([]*model.GitHubOrg, error) {
	client := setupClient(ctx, g.API, user)
	return getOrgs(ctx, client)
}

func getOrgs(ctx context.Context, client *github.Client) ([]*model.GitHubOrg, error) {
	var orgs []*github.Organization
	resp, err := buildCompleteList(func(opts *github.ListOptions) (*github.Response, error) {
		newOrgs, resp, err := client.Organizations.List(ctx, "", opts)
		orgs = append(orgs, newOrgs...)
		return resp, err
	})
	if err != nil {
		err = fmt.Errorf("Fetching orgs. %s", err)
		return nil, createError(resp, err)
	}
	res := []*model.GitHubOrg{}
	for _, org := range orgs {
		o := model.GitHubOrg{
			Login:  org.GetLogin(),
			Avatar: org.GetAvatarURL(),
			Admin:  orgIsAdmin(ctx, client, org.GetLogin()),
		}
		res = append(res, &o)
	}
	sort.Slice(res, func(i, j int) bool {
		return strings.ToLower(res[i].Login) < strings.ToLower(res[j].Login)
	})
	return res, nil
}

func orgIsAdmin(ctx context.Context, client *github.Client, owner string) bool {
	m, err := getOrgPerm(ctx, client, owner)
	if err != nil {
		//don't care, just return false
		return false
	}
	return m.Admin
}

func (g *Github) GetPerson(ctx context.Context, user *model.User, login string) (*model.Person, error) {
	client := setupClient(ctx, g.API, user)
	return getPerson(ctx, client, login)
}

func getPerson(ctx context.Context, client *github.Client, login string) (*model.Person, error) {
	user, resp, err := client.Users.Get(ctx, login)
	if err != nil {
		err = fmt.Errorf("Accessing information for user %s. %s", login, err)
		return nil, createError(resp, err)
	}
	return &model.Person{
		Login: login,
		Name:  user.GetName(),
		Email: user.GetEmail(),
	}, nil
}

func (g *Github) ListTeams(ctx context.Context, user *model.User, org string) (set.Set, error) {
	client := setupClient(ctx, g.API, user)
	resp, err := getTeams(ctx, client, org)
	if err != nil {
		return nil, err
	}
	teams := set.Empty()
	for _, t := range resp {
		teams.Add(t.GetSlug())
	}
	return teams, nil
}

func getTeams(ctx context.Context, client *github.Client, org string) ([]*github.Team, error) {
	var teams []*github.Team
	resp, err := buildCompleteList(func(opts *github.ListOptions) (*github.Response, error) {
		newTeams, response, err := client.Organizations.ListTeams(ctx, org, opts)
		teams = append(teams, newTeams...)
		return response, err
	})
	if err != nil {
		err = fmt.Errorf("Accessing teams for organization %s. %s", org, err)
		return nil, createError(resp, err)
	}
	return teams, nil
}

func (g *Github) GetTeamMembers(ctx context.Context, user *model.User, org string, team string) (set.Set, error) {
	client := setupClient(ctx, g.API, user)
	return getTeamMembers(ctx, client, org, team)
}

func getTeamMembers(ctx context.Context, client *github.Client, org string, team string) (set.Set, error) {
	teams, err := getTeams(ctx, client, org)
	if err != nil {
		return nil, err
	}
	var id *int
	for _, t := range teams {
		if strings.EqualFold(t.GetSlug(), team) {
			id = t.ID
			break
		}
	}
	if id == nil {
		err = fmt.Errorf("Team %s not found for organization %s.", team, org)
		return nil, exterror.Create(http.StatusNotFound, err)
	}
	topts := github.OrganizationListTeamMembersOptions{}
	var teammates []*github.User
	resp, err := buildCompleteList(func(opts *github.ListOptions) (*github.Response, error) {
		topts.ListOptions = *opts
		newTmates, resp2, err2 := client.Organizations.ListTeamMembers(ctx, *id, &topts)
		teammates = append(teammates, newTmates...)
		return resp2, err2
	})
	if err != nil {
		err = fmt.Errorf("Fetching team %s members for organization %s. %s", team, org, err)
		return nil, createError(resp, err)
	}
	names := set.Empty()
	for _, u := range teammates {
		names.Add(*u.Login)
	}
	return names, nil
}

func (g *Github) GetOrgMembers(ctx context.Context, user *model.User, org string) (set.Set, error) {
	client := setupClient(ctx, g.API, user)
	return getOrgMembers(ctx, client, org)
}

func getOrgMembers(ctx context.Context, client *github.Client, org string) (set.Set, error) {
	var orgmembers []*github.User
	lmOptions := github.ListMembersOptions{}
	resp, err := buildCompleteList(func(opts *github.ListOptions) (*github.Response, error) {
		lmOptions.ListOptions = *opts
		newOrgs, resp, err := client.Organizations.ListMembers(ctx, org, &lmOptions)
		orgmembers = append(orgmembers, newOrgs...)
		return resp, err
	})
	if err != nil {
		err = fmt.Errorf("Accessing organization %s. %s", org, err)
		return nil, createError(resp, err)
	}
	names := set.Empty()
	for _, u := range orgmembers {
		names.Add(u.GetLogin())
	}
	return names, nil
}

func (g *Github) GetCollaborators(ctx context.Context, user *model.User, owner, name string) (set.Set, error) {
	client := setupClient(ctx, g.API, user)
	return getCollaborators(ctx, client, owner, name)
}

func getCollaborators(ctx context.Context, client *github.Client, owner, name string) (set.Set, error) {
	var collab []*github.User
	lcOptions := github.ListCollaboratorsOptions{}
	resp, err := buildCompleteList(func(opts *github.ListOptions) (*github.Response, error) {
		lcOptions.ListOptions = *opts
		next, resp, err := client.Repositories.ListCollaborators(ctx, owner, name, &lcOptions)
		collab = append(collab, next...)
		return resp, err
	})
	if err != nil {
		err = fmt.Errorf("Accessing collaborators for %s/%s. %s", owner, name, err)
		return nil, createError(resp, err)
	}
	names := set.Empty()
	for _, u := range collab {
		names.Add(u.GetLogin())
	}
	return names, nil
}

func (g *Github) GetRepo(ctx context.Context, user *model.User, owner, name string) (*model.Repo, error) {
	client := setupClient(ctx, g.API, user)
	return getRepo(ctx, client, owner, name)
}

func getRepo(ctx context.Context, client *github.Client, owner, name string) (*model.Repo, error) {
	repo, resp, err := client.Repositories.Get(ctx, owner, name)
	if err != nil {
		err = fmt.Errorf("Fetching repository. %s", err)
		return nil, createError(resp, err)
	}
	return &model.Repo{
		Owner:   owner,
		Name:    name,
		Slug:    repo.GetFullName(),
		Link:    repo.GetHTMLURL(),
		Private: repo.GetPrivate(),
		Org:     repo.Organization != nil,
	}, nil
}

func (g *Github) GetOrg(ctx context.Context, user *model.User, owner string) (*model.OrgDb, error) {
	client := setupClient(ctx, g.API, user)
	return getOrg(ctx, client, owner)
}

func getOrg(ctx context.Context, client *github.Client, owner string) (*model.OrgDb, error) {
	org, resp, err := client.Organizations.Get(ctx, owner)
	if err != nil {
		err = fmt.Errorf("Fetching org %s. %s", owner, err)
		return nil, createError(resp, err)
	}
	return &model.OrgDb{
		Owner:   owner,
		Link:    org.GetHTMLURL(),
		Private: false,
	}, nil
}

func (g *Github) GetPerm(ctx context.Context, user *model.User, owner, name string) (*model.Perm, error) {
	client := setupClient(ctx, g.API, user)
	return getPerm(ctx, client, owner, name)
}

func getPerm(ctx context.Context, client *github.Client, owner, name string) (*model.Perm, error) {
	repo, resp, err := client.Repositories.Get(ctx, owner, name)
	if err != nil {
		err = fmt.Errorf("Fetching repository. %s", err)
		return nil, createError(resp, err)
	}
	m := &model.Perm{}
	m.Admin = (*repo.Permissions)["admin"]
	m.Push = (*repo.Permissions)["push"]
	m.Pull = (*repo.Permissions)["pull"]
	return m, nil
}

func (g *Github) GetOrgPerm(ctx context.Context, user *model.User, owner string) (*model.Perm, error) {
	client := setupClient(ctx, g.API, user)
	return getOrgPerm(ctx, client, owner)
}

func getOrgPerm(ctx context.Context, client *github.Client, owner string) (*model.Perm, error) {
	perms, resp, err := client.Organizations.GetOrgMembership(ctx, "", owner)
	if err != nil {
		err = fmt.Errorf("Fetching org permission. %s", err)
		return nil, createError(resp, err)
	}
	m := &model.Perm{}
	m.Admin = (*perms.Role == "admin")
	return m, nil
}

func (g *Github) GetUserRepos(ctx context.Context, u *model.User) ([]*model.Repo, error) {
	client := setupClient(ctx, g.API, u)
	all, err := getUserRepos(ctx, client, u.Login)
	if err != nil {
		return nil, err
	}

	repos := []*model.Repo{}
	for _, repo := range all {
		repos = append(repos, &model.Repo{
			Owner:   repo.Owner.GetLogin(),
			Name:    repo.GetName(),
			Slug:    repo.GetFullName(),
			Link:    repo.GetHTMLURL(),
			Private: repo.GetPrivate(),
			Org:     repo.Organization != nil,
		})
	}

	return repos, nil
}

func (g *Github) GetOrgRepos(ctx context.Context, u *model.User, owner string) ([]*model.Repo, error) {
	client := setupClient(ctx, g.API, u)
	all, err := getOrgRepos(ctx, client, owner)
	if err != nil {
		return nil, err
	}

	repos := []*model.Repo{}
	for _, repo := range all {
		// only list repositories that I can admin
		if repo.Permissions == nil || (*repo.Permissions)["admin"] == false {
			continue
		}
		repos = append(repos, &model.Repo{
			Owner:   repo.Owner.GetLogin(),
			Name:    repo.GetName(),
			Slug:    repo.GetFullName(),
			Link:    repo.GetHTMLURL(),
			Private: repo.GetPrivate(),
			Org:     repo.Organization != nil,
		})
	}

	return repos, nil
}

func (g *Github) SetHook(ctx context.Context, user *model.User, repo *model.Repo, link string) error {
	client := setupClient(ctx, g.API, user)
	return g.setHook(ctx, client, user, repo, link)
}

// createProtectionRequest takes the GitHub protection status returned by
// GET /repos/:owner/:repo/branches/:branch/protection and converts it into
// a format suitable for PUT /repos/:owner/:repo/branches/:branch/protection
// (GitHub please fix your API)
func createProtectionRequest(input *github.Protection) *github.ProtectionRequest {
	output := &github.ProtectionRequest{
		RequiredStatusChecks:       nil,
		EnforceAdmins:              false,
		RequiredPullRequestReviews: nil,
		Restrictions:               nil,
	}
	if input == nil {
		return output
	}
	output.RequiredStatusChecks = input.RequiredStatusChecks
	if input.EnforceAdmins != nil {
		output.EnforceAdmins = input.EnforceAdmins.Enabled
	}
	if input.RequiredPullRequestReviews != nil {
		inReviews := input.RequiredPullRequestReviews
		outReviews := &github.PullRequestReviewsEnforcementRequest{
			DismissalRestrictionsRequest: nil,
			DismissStaleReviews:          inReviews.DismissStaleReviews,
			RequireCodeOwnerReviews:      inReviews.RequireCodeOwnerReviews,
		}
		inDismissal := inReviews.DismissalRestrictions
		if len(inDismissal.Users) > 0 || len(inDismissal.Teams) > 0 {
			outDismissal := &github.DismissalRestrictionsRequest{
				Users: []string{},
				Teams: []string{},
			}
			for _, user := range inDismissal.Users {
				outDismissal.Users = append(outDismissal.Users, user.GetLogin())
			}
			for _, team := range inDismissal.Teams {
				outDismissal.Teams = append(outDismissal.Teams, team.GetSlug())
			}
			outReviews.DismissalRestrictionsRequest = outDismissal
		}
		output.RequiredPullRequestReviews = outReviews
	}
	if input.Restrictions != nil {
		inRestrict := input.Restrictions
		outRestrict := &github.BranchRestrictionsRequest{
			Users: []string{},
			Teams: []string{},
		}
		for _, user := range inRestrict.Users {
			outRestrict.Users = append(outRestrict.Users, user.GetLogin())
		}
		for _, team := range inRestrict.Teams {
			outRestrict.Teams = append(outRestrict.Teams, team.GetSlug())
		}
		output.Restrictions = outRestrict
	}
	return output
}

func addBranchProtection(ctx context.Context, client *github.Client, owner, repo, branch string) error {
	protect, resp, err := client.Repositories.GetBranchProtection(ctx, owner, repo, branch)
	if err != nil && resp.StatusCode != http.StatusNotFound {
		return createError(resp, err)
	}
	preq := createProtectionRequest(protect)
	if preq.RequiredStatusChecks == nil {
		preq.RequiredStatusChecks = &github.RequiredStatusChecks{
			Strict:   true,
			Contexts: []string{},
		}
	}
	for _, ctx := range preq.RequiredStatusChecks.Contexts {
		if ctx == model.ServiceName {
			return nil
		}
	}
	preq.RequiredStatusChecks.Contexts = append(preq.RequiredStatusChecks.Contexts, model.ServiceName)
	_, resp, err = client.Repositories.UpdateBranchProtection(ctx, owner, repo, branch, preq)
	if err != nil {
		return createError(resp, err)
	}
	return nil
}

func removeBranchProtection(ctx context.Context, client *github.Client, owner, repo, branch string) error {
	protect, resp, err := client.Repositories.GetBranchProtection(ctx, owner, repo, branch)
	if resp.StatusCode == http.StatusNotFound {
		return nil
	}
	if err != nil {
		return createError(resp, err)
	}
	preq := createProtectionRequest(protect)
	if preq.RequiredStatusChecks == nil {
		return nil
	}
	success := false
	ctxs := preq.RequiredStatusChecks.Contexts
	for i, ctx := range ctxs {
		if ctx == model.ServiceName {
			success = true
			ctxs[i] = ctxs[len(ctxs)-1]
			ctxs = ctxs[:len(ctxs)-1]
			break
		}
	}
	if !success {
		return nil
	}
	preq.RequiredStatusChecks.Contexts = ctxs
	_, resp, err = client.Repositories.UpdateBranchProtection(ctx, owner, repo, branch, preq)
	if err != nil {
		return createError(resp, err)
	}
	return nil
}

func (g *Github) setHook(ctx context.Context, client *github.Client, user *model.User, repo *model.Repo, link string) error {
	old, err := getHook(ctx, client, repo.Owner, repo.Name, link)
	if err == nil && old != nil {
		client.Repositories.DeleteHook(ctx, repo.Owner, repo.Name, old.GetID())
	}

	_, err = createHook(ctx, client, repo.Owner, repo.Name, link)
	if err != nil {
		log.Debugf("Creating the webhook at %s. %s", link, err)
		return err
	}

	r, resp, err := client.Repositories.Get(ctx, repo.Owner, repo.Name)
	if err != nil {
		return createError(resp, err)
	}

	_, resp, err = client.Repositories.GetBranch(ctx, repo.Owner, repo.Name, r.GetDefaultBranch())
	if resp.StatusCode == http.StatusNotFound {
		return nil
	}
	if err != nil {
		return createError(resp, err)
	}

	return addBranchProtection(ctx, client, repo.Owner, repo.Name, r.GetDefaultBranch())
}

func (g *Github) DelHook(ctx context.Context, user *model.User, repo *model.Repo, link string) error {
	client := setupClient(ctx, g.API, user)
	return g.delHook(ctx, client, user, repo, link)
}

func (g *Github) delHook(ctx context.Context, client *github.Client, user *model.User, repo *model.Repo, link string) error {
	hook, err := getHook(ctx, client, repo.Owner, repo.Name, link)
	if err != nil {
		return err
	} else if hook == nil {
		return nil
	}
	resp, err := client.Repositories.DeleteHook(ctx, repo.Owner, repo.Name, hook.GetID())
	if err != nil {
		return createError(resp, err)
	}

	r, resp, err := client.Repositories.Get(ctx, repo.Owner, repo.Name)
	if err != nil {
		return createError(resp, err)
	}

	_, resp, err = client.Repositories.GetBranch(ctx, repo.Owner, repo.Name, r.GetDefaultBranch())
	if resp.StatusCode == http.StatusNotFound {
		return nil
	}
	if err != nil {
		return createError(resp, err)
	}

	return removeBranchProtection(ctx, client, repo.Owner, repo.Name, r.GetDefaultBranch())
}

func (g *Github) SetOrgHook(ctx context.Context, user *model.User, org *model.OrgDb, link string) error {
	client := setupClient(ctx, g.API, user)

	old, err := getOrgHook(ctx, client, org.Owner, link)
	if err == nil && old != nil {
		client.Organizations.DeleteHook(ctx, org.Owner, old.GetID())
	}

	_, err = createOrgHook(ctx, client, org.Owner, link)
	if err != nil {
		log.Debugf("Creating the webhook at %s. %s", link, err)
		return err
	}

	return nil
}

func (g *Github) DelOrgHook(ctx context.Context, user *model.User, org *model.OrgDb, link string) error {
	client := setupClient(ctx, g.API, user)

	hook, err := getOrgHook(ctx, client, org.Owner, link)
	if err != nil {
		return err
	} else if hook == nil {
		return nil
	}
	resp, err := client.Organizations.DeleteHook(ctx, org.Owner, hook.GetID())
	if err != nil {
		return createError(resp, err)
	}

	return nil
}

// buildOtherContextSlice returns all contexts besides the one for this service
func buildOtherContextSlice(branch *Branch) []string {
	checks := []string{}
	for _, check := range branch.Protection.Checks.Contexts {
		if check != model.ServiceName {
			checks = append(checks, check)
		}
	}
	return checks
}

func (g *Github) GetAllComments(ctx context.Context, u *model.User, r *model.Repo, num int) ([]*model.Comment, error) {
	client := setupClient(ctx, g.API, u)
	return getAllComments(ctx, client, r, num)
}

func getAllComments(ctx context.Context, client *github.Client, r *model.Repo, num int) ([]*model.Comment, error) {
	lcOpts := github.IssueListCommentsOptions{Direction: "desc", Sort: "created"}
	var comm []*github.IssueComment
	resp, err := buildCompleteList(func(opts *github.ListOptions) (*github.Response, error) {
		lcOpts.ListOptions = *opts
		newCom, resp, err := client.Issues.ListComments(ctx, r.Owner, r.Name, num, &lcOpts)
		comm = append(comm, newCom...)
		return resp, err
	})
	if err != nil {
		return nil, createError(resp, err)
	}
	comments := []*model.Comment{}
	for _, comment := range comm {
		comments = append(comments, &model.Comment{
			Author:      lowercase.Create(*comment.User.Login),
			Body:        comment.GetBody(),
			SubmittedAt: comment.GetCreatedAt(),
		})
	}
	return comments, nil
}

func (g *Github) IsHeadUIMerge(ctx context.Context, u *model.User, r *model.Repo, num int) (bool, error) {
	client := setupClient(ctx, g.API, u)
	pr, resp, err := client.PullRequests.Get(ctx, r.Owner, r.Name, num)
	if err != nil {
		return false, createError(resp, err)
	}
	commit, resp, err := client.Repositories.GetCommit(ctx, r.Owner, r.Name, *pr.Head.SHA)
	if err != nil {
		return false, createError(resp, err)
	}
	return isCommitUIMerge(commit), nil
}

func (g *Github) GetCommentsSinceHead(ctx context.Context, u *model.User, r *model.Repo, num int, noUIMerge bool) ([]*model.Comment, error) {
	client := setupClient(ctx, g.API, u)
	return getCommentsSinceHead(ctx, client, r, num, noUIMerge)
}

func getHead(ctx context.Context, client *github.Client, r *model.Repo, num int, noUIMerge bool) (*github.RepositoryCommit, error) {
	pr, resp, err := client.PullRequests.Get(ctx, r.Owner, r.Name, num)
	if err != nil {
		return nil, createError(resp, err)
	}
	commit, resp, err := client.Repositories.GetCommit(ctx, r.Owner, r.Name, *pr.Head.SHA)
	if err != nil {
		return nil, createError(resp, err)
	}
	if noUIMerge {
		commit, err = ignoreUIMerge(ctx, client, r, pr, commit)
		if err != nil {
			return nil, err
		}
	}
	return commit, nil
}

func getCommentsSinceHead(ctx context.Context, client *github.Client, r *model.Repo, num int, noUIMerge bool) ([]*model.Comment, error) {
	commit, err := getHead(ctx, client, r, num, noUIMerge)
	if err != nil {
		return nil, err
	}
	lcOpts := github.IssueListCommentsOptions{
		Direction: "desc",
		Sort:      "created",
		Since:     commit.Commit.Committer.GetDate()}
	var comm []*github.IssueComment
	resp, err := buildCompleteList(func(opts *github.ListOptions) (*github.Response, error) {
		lcOpts.ListOptions = *opts
		newCom, resp2, err2 := client.Issues.ListComments(ctx, r.Owner, r.Name, num, &lcOpts)
		comm = append(comm, newCom...)
		return resp2, err2
	})

	if err != nil {
		return nil, createError(resp, err)
	}
	comments := []*model.Comment{}
	for _, comment := range comm {
		comments = append(comments, &model.Comment{
			Author: lowercase.Create(*comment.User.Login),
			Body:   comment.GetBody(),
		})
	}
	return comments, nil
}

func (g *Github) GetAllReviews(ctx context.Context, u *model.User, r *model.Repo, num int) ([]*model.Review, error) {
	client := setupClient(ctx, g.API, u)
	return getAllReviews(ctx, client, r, num)
}

func getAllReviews(ctx context.Context, client *github.Client, r *model.Repo, num int) ([]*model.Review, error) {
	var extReviews []*github.PullRequestReview
	resp, err := buildCompleteList(func(opts *github.ListOptions) (*github.Response, error) {
		newReviews, resp2, err2 := client.PullRequests.ListReviews(ctx, r.Owner, r.Name, num, opts)
		extReviews = append(extReviews, newReviews...)
		return resp2, err2
	})
	if err != nil {
		return nil, createError(resp, err)
	}
	// error checking to handle https://github.com/google/go-github/issues/540
	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("Unable to retrieve PR reviews for %s %d", r.Slug, num)
		return nil, createError(resp, err)
	}
	reviews := []*model.Review{}
	for i, rev := range extReviews {
		if rev == nil || rev.User == nil {
			log.Errorf("Found a nil rev or nil rev.User in the github.PullRequestReview element %d for review %d in repo %+v.", i, num, *r)
			continue
		}
		reviews = append(reviews, &model.Review{
			ID:          rev.GetID(),
			Author:      lowercase.Create(rev.User.GetLogin()),
			State:       lowercase.Create(rev.GetState()),
			Body:        rev.GetBody(),
			SubmittedAt: rev.GetSubmittedAt(),
		})
	}
	return reviews, nil
}

func (g *Github) GetReviewsSinceHead(ctx context.Context, u *model.User, r *model.Repo, num int, noUIMerge bool) ([]*model.Review, error) {
	client := setupClient(ctx, g.API, u)
	return getReviewsSinceHead(ctx, client, r, num, noUIMerge)
}

func getReviewsSinceHead(ctx context.Context, client *github.Client, r *model.Repo, num int, noUIMerge bool) ([]*model.Review, error) {
	commit, err := getHead(ctx, client, r, num, noUIMerge)
	if err != nil {
		return nil, err
	}
	all, err2 := getAllReviews(ctx, client, r, num)
	if err2 != nil {
		return nil, err2
	}
	reviews := []*model.Review{}
	for _, rev := range all {
		if rev.SubmittedAt.After(commit.Commit.Committer.GetDate()) {
			reviews = append(reviews, rev)
		}
	}
	return reviews, nil
}

func (g *Github) CreateURLCompare(c context.Context, u *model.User, r *model.Repo, sha1, sha2 string) string {
	return fmt.Sprintf("%s/%s/%s/compare/%s...%s", g.URL, r.Owner, r.Name, sha1, sha2)
}

func (g *Github) GetCommits(ctx context.Context, u *model.User, r *model.Repo, sha string, page, perPage int) ([]string, int, error) {
	client := setupClient(ctx, g.API, u)
	return getCommits(ctx, client, r, sha, page, perPage)
}

func getCommits(ctx context.Context, client *github.Client, r *model.Repo, sha string, page, perPage int) ([]string, int, error) {
	var commits []string
	opt := github.CommitsListOptions{
		SHA: sha,
		ListOptions: github.ListOptions{
			Page:    page,
			PerPage: perPage,
		},
	}
	lst, resp, err := client.Repositories.ListCommits(ctx, r.Owner, r.Name, &opt)
	if err != nil {
		return nil, 0, createError(resp, err)
	}
	for _, val := range lst {
		commits = append(commits, val.GetSHA())
	}
	return commits, resp.NextPage, nil
}

func (g *Github) GetContents(ctx context.Context, u *model.User, r *model.Repo, path string) ([]byte, error) {
	client := setupClient(ctx, g.API, u)
	return getContents(ctx, client, r, path)
}

func getContents(ctx context.Context, client *github.Client, r *model.Repo, path string) ([]byte, error) {
	content, _, resp, err := client.Repositories.GetContents(ctx, r.Owner, r.Name, path, nil)
	if err != nil {
		return nil, createError(resp, err)
	}
	body, err := content.GetContent()
	if err != nil {
		return nil, err
	}
	return []byte(body), nil
}

func (g *Github) GetStatus(ctx context.Context, u *model.User, r *model.Repo, sha string) (model.CombinedStatus, error) {
	client := setupClient(ctx, g.API, u)
	return getStatus(ctx, client, r, sha)
}

func getStatus(ctx context.Context, client *github.Client, r *model.Repo, sha string) (model.CombinedStatus, error) {
	result := model.CombinedStatus{}
	status, resp, err := client.Repositories.GetCombinedStatus(ctx, r.Owner, r.Name, sha, nil)
	if err != nil {
		return result, createError(resp, err)
	}
	result.State = status.GetState()
	result.Statuses = make(map[string]model.CommitStatus)
	for _, s := range status.Statuses {
		result.Statuses[s.GetContext()] = model.CommitStatus{
			Context:     s.GetContext(),
			Description: s.GetDescription(),
			State:       s.GetState(),
		}
	}
	return result, nil
}

func getRequiredStatusChecks(ctx context.Context, client *github.Client, r *model.Repo, branch string) ([]string, error) {
	content, resp, err := client.Repositories.GetRequiredStatusChecks(ctx, r.Owner, r.Name, branch)
	if err != nil {
		if resp.StatusCode == http.StatusNotFound {
			return nil, nil
		}
		return nil, createError(resp, err)
	}
	return content.Contexts, nil
}

func (g *Github) HasRequiredStatus(ctx context.Context, u *model.User, r *model.Repo, branch, sha string) (bool, error) {
	client := setupClient(ctx, g.API, u)
	return hasRequiredStatus(ctx, client, r, branch, sha)
}

func hasRequiredStatus(ctx context.Context, client *github.Client, r *model.Repo, branch, sha string) (bool, error) {
	status, err := getStatus(ctx, client, r, sha)
	if err != nil {
		return false, err
	}
	if status.State != "success" {
		return false, nil
	}
	log.Debug("overall status is success -- checking to see if all status checks returned success")
	required, err := getRequiredStatusChecks(ctx, client, r, branch)
	if err != nil {
		return false, err
	}
	for _, r := range required {
		if _, ok := status.Statuses[r]; !ok {
			return false, nil
		}
	}
	for _, s := range status.Statuses {
		if s.State != "success" {
			return false, nil
		}
	}
	return true, nil
}

func (g *Github) SetStatus(ctx context.Context, u *model.User, r *model.Repo, sha, context, status, desc string) error {
	client := setupClient(ctx, g.API, u)
	return setStatus(ctx, client, r, sha, context, status, desc)
}

func setStatus(ctx context.Context, client *github.Client, r *model.Repo, sha, context, status, desc string) error {
	// An undocumented feature of GitHub statuses is a 140 character limit
	if len(desc) > 135 {
		desc = desc[:135] + "..."
	}

	data := github.RepoStatus{
		Context:     github.String(context),
		State:       github.String(status),
		Description: github.String(desc),
	}

	_, resp, err := client.Repositories.CreateStatus(ctx, r.Owner, r.Name, sha, &data)
	if err != nil {
		return createError(resp, err)
	}
	return nil
}

func (g *Github) CreateEmptyCommit(ctx context.Context, u *model.User, r *model.Repo, sha, msg string) (string, error) {
	client := setupClient(ctx, g.API, u)
	return createEmptyCommit(ctx, client, r, sha, msg)
}

func createEmptyCommit(ctx context.Context, client *github.Client, r *model.Repo, sha, msg string) (string, error) {
	prev, resp, err := client.Git.GetCommit(ctx, r.Owner, r.Name, sha)
	if err != nil {
		return "", createError(resp, err)
	}
	commit, resp, err := client.Git.CreateCommit(ctx, r.Owner, r.Name, &github.Commit{
		Message: github.String(msg),
		Tree:    prev.Tree,
		Parents: []github.Commit{{
			SHA: github.String(sha),
		}},
	})
	if err != nil {
		return "", createError(resp, err)
	}
	return commit.GetSHA(), nil
}

func (g *Github) CreateReference(ctx context.Context, u *model.User, r *model.Repo, sha, name string) (string, error) {
	client := setupClient(ctx, g.API, u)
	return createReference(ctx, client, r, sha, name)
}

func createReference(ctx context.Context, client *github.Client, r *model.Repo, sha, name string) (string, error) {
	ref, resp, err := client.Git.CreateRef(ctx, r.Owner, r.Name, &github.Reference{
		Ref: github.String(name),
		Object: &github.GitObject{
			Type: github.String("commit"),
			SHA:  github.String(sha),
		},
	})
	if err != nil {
		return "", createError(resp, err)
	}
	return ref.Object.GetSHA(), nil
}

func (g *Github) CreatePR(ctx context.Context, u *model.User, r *model.Repo, title, head, base, body string) (int, error) {
	client := setupClient(ctx, g.API, u)
	return createPR(ctx, client, r, title, head, base, body)
}

func createPR(ctx context.Context, client *github.Client, r *model.Repo, title, head, base, body string) (int, error) {
	pr := github.NewPullRequest{
		Title: github.String(title),
		Head:  github.String(head),
		Base:  github.String(base),
		Body:  github.String(body),
	}
	pullRequest, resp, err := client.PullRequests.Create(ctx, r.Owner, r.Name, &pr)
	if err != nil {
		return 0, createError(resp, err)
	}
	return pullRequest.GetNumber(), nil
}

func (g *Github) GetPullRequest(ctx context.Context, u *model.User, r *model.Repo, number int) (model.PullRequest, error) {
	client := setupClient(ctx, g.API, u)
	return getPullRequest(ctx, client, r, number)
}

func getPullRequest(ctx context.Context, client *github.Client, r *model.Repo, number int) (model.PullRequest, error) {
	pr, _, err := doGetPullRequest(ctx, client, r, number)
	return pr, err
}

func (g *Github) GetPullRequestFiles(ctx context.Context, u *model.User, r *model.Repo, number int) ([]model.CommitFile, error) {
	client := setupClient(ctx, g.API, u)
	return getPullRequestFiles(ctx, client, r, number)
}

func getPullRequestFiles(ctx context.Context, client *github.Client, r *model.Repo, number int) ([]model.CommitFile, error) {
	var files []*github.CommitFile
	resp, err := buildCompleteList(func(opts *github.ListOptions) (*github.Response, error) {
		newFiles, resp, err := client.PullRequests.ListFiles(ctx, r.Owner, r.Name, number, opts)
		files = append(files, newFiles...)
		return resp, err
	})
	if err != nil {
		return nil, createError(resp, err)
	}
	res := []model.CommitFile{}
	for _, f := range files {
		if f.Filename == nil {
			log.Warnf("Repo %s pr %d has a modified file with no filename: %s",
				r.Name, number, f.String())
			continue
		}
		res = append(res, model.CommitFile{Filename: f.GetFilename()})
	}
	return res, nil
}

func (g *Github) GetPullRequestsForCommit(ctx context.Context, u *model.User, r *model.Repo, sha *string) ([]model.PullRequest, error) {
	client := setupClient(ctx, g.API, u)
	return getPullRequestsForCommit(ctx, client, r, sha)
}

func getPullRequestsForCommit(ctx context.Context, client *github.Client, r *model.Repo, sha *string) ([]model.PullRequest, error) {
	log.Debug("sha == ", *sha)
	issues, resp, err := client.Search.Issues(ctx, fmt.Sprintf("%s&type=pr", *sha), &github.SearchOptions{
		TextMatch: false,
	})
	if err != nil {
		return nil, createError(resp, err)
	}
	out := []model.PullRequest{}
	for _, v := range issues.Issues {
		log.Debugf("got pull request %v", v)
		if v.State != nil && *v.State == "closed" {
			log.Debugf("skipping pull request %s because it's closed", *v.Title)
			continue
		}
		pr, prHeadSha, err := doGetPullRequest(ctx, client, r, *v.Number)

		if err != nil {
			return nil, err
		}

		if *prHeadSha != *sha {
			log.Debugf("Pull Request %d has sha %s at head, not sha %s, so not a pull request for this commit", pr.Number, *prHeadSha, *sha)
			continue
		}

		out = append(out, pr)

	}
	return out, nil
}

func (g *Github) GetIssue(ctx context.Context, u *model.User, r *model.Repo, number int) (model.Issue, error) {
	client := setupClient(ctx, g.API, u)
	return getIssue(ctx, client, r, number)
}

func getIssue(ctx context.Context, client *github.Client, r *model.Repo, number int) (model.Issue, error) {
	issue, resp, err := client.Issues.Get(ctx, r.Owner, r.Name, number)
	if err != nil {
		return model.Issue{}, exterror.Create(resp.StatusCode, err)
	}
	result := model.Issue{
		Number: number,
		Title:  issue.GetTitle(),
		Author: lowercase.Create(issue.User.GetLogin()),
	}
	return result, nil
}

func (g *Github) MergePR(ctx context.Context, u *model.User, r *model.Repo, pullRequest model.PullRequest, approvers []*model.Person, message string, mergeMethod string) (*string, error) {
	client := setupClient(ctx, g.API, u)
	return mergePR(ctx, client, r, pullRequest, approvers, message, mergeMethod)
}

func mergePR(ctx context.Context, client *github.Client, r *model.Repo, pullRequest model.PullRequest, approvers []*model.Person, message string, mergeMethod string) (*string, error) {
	log.Debugf("incoming message: %v", message)
	msg := message
	if len(msg) > 0 {
		msg += "\n"
	}
	msg += fmt.Sprintf("Merged by %s\n", envvars.Env.Branding.ShortName)
	if len(approvers) > 0 {
		apps := "Approved by:\n"
		for _, v := range approvers {
			//Brad Rydzewski <brad.rydzewski@mail.com> (@bradrydzewski)
			if len(v.Name) > 0 {
				apps += fmt.Sprintf("%s", v.Name)
			}
			if len(v.Email) > 0 {
				apps += fmt.Sprintf(" <%s>", v.Email)
			}
			if len(v.Login) > 0 {
				apps += fmt.Sprintf(" (@%s)", v.Login)
			}
			apps += "\n"
		}
		msg += apps
	}
	log.Debugf("Constructed message: %v", msg)
	options := github.PullRequestOptions{MergeMethod: "merge"}
	options.MergeMethod = mergeMethod
	result, resp, err := client.PullRequests.Merge(ctx, r.Owner, r.Name, pullRequest.Number, msg, &options)
	if err != nil {
		return nil, createError(resp, err)
	}

	if !(*result.Merged) {
		return nil, errors.New(*result.Message)
	}
	return result.SHA, nil
}

func (g *Github) CompareBranches(ctx context.Context, u *model.User, repo *model.Repo, base string, head string, owner string) (model.BranchCompare, error) {
	client := setupClient(ctx, g.API, u)
	return compareBranches(ctx, client, repo, base, head, owner)
}

func compareBranches(ctx context.Context, client *github.Client, repo *model.Repo, base string, head string, owner string) (model.BranchCompare, error) {
	var result model.BranchCompare
	if repo.Owner != owner {
		head = owner + ":" + head
	}
	res, resp, err := client.Repositories.CompareCommits(ctx, repo.Owner, repo.Name, base, head)
	if err != nil {
		return result, createError(resp, err)
	}
	result.AheadBy = res.GetAheadBy()
	result.BehindBy = res.GetBehindBy()
	result.Status = res.GetStatus()
	result.TotalCommits = res.GetTotalCommits()
	return result, nil
}

func (g *Github) DeleteBranch(ctx context.Context, u *model.User, repo *model.Repo, name string) error {
	client := setupClient(ctx, g.API, u)
	return deleteBranch(ctx, client, repo, name)
}

func deleteBranch(ctx context.Context, client *github.Client, repo *model.Repo, name string) error {
	resp, err := client.Git.DeleteRef(ctx, repo.Owner, repo.Name, "refs/heads/"+name)
	if err != nil {
		err = fmt.Errorf("Deleting branch %s/%s/%s. %s", repo.Owner, repo.Name, name, err)
		return createError(resp, err)
	}
	return nil
}

func (g *Github) ListTags(ctx context.Context, u *model.User, r *model.Repo) ([]model.Tag, error) {
	client := setupClient(ctx, g.API, u)
	return listTags(ctx, client, r)
}

func listTags(ctx context.Context, client *github.Client, r *model.Repo) ([]model.Tag, error) {
	var tags []*github.RepositoryTag
	resp, err := buildCompleteList(func(opts *github.ListOptions) (*github.Response, error) {
		newTags, resp, err := client.Repositories.ListTags(ctx, r.Owner, r.Name, opts)
		tags = append(tags, newTags...)
		return resp, err
	})

	if err != nil {
		return nil, createError(resp, err)
	}
	out := make([]model.Tag, len(tags))
	for k, v := range tags {
		out[k] = model.Tag(*v.Name)
	}
	return out, nil
}

func (g *Github) Tag(ctx context.Context, u *model.User, r *model.Repo, tag *string, sha *string) error {
	client := setupClient(ctx, g.API, u)
	return doTag(ctx, client, r, tag, sha)
}

func doTag(ctx context.Context, client *github.Client, r *model.Repo, tag *string, sha *string) error {
	t := time.Now()
	gittag, resp, err := client.Git.CreateTag(ctx, r.Owner, r.Name, &github.Tag{
		Tag:     tag,
		SHA:     sha,
		Message: github.String(fmt.Sprintf("Tagged by %s", envvars.Env.Branding.ShortName)),
		Tagger: &github.CommitAuthor{
			Date:  &t,
			Name:  github.String(envvars.Env.Branding.ShortName),
			Email: github.String(envvars.Env.Github.Email),
		},
		Object: &github.GitObject{
			SHA:  sha,
			Type: github.String("commit"),
		},
	})

	if err != nil {
		return createError(resp, err)
	}
	_, resp, err = client.Git.CreateRef(ctx, r.Owner, r.Name, &github.Reference{
		Ref: github.String("refs/tags/" + *tag),
		Object: &github.GitObject{
			SHA: gittag.SHA,
		},
	})

	if err != nil {
		return createError(resp, err)
	}

	return nil
}

func (g *Github) WriteComment(ctx context.Context, u *model.User, r *model.Repo, num int, message string) error {
	client := setupClient(ctx, g.API, u)
	return writeComment(ctx, client, r, num, message)
}

func writeComment(ctx context.Context, client *github.Client, r *model.Repo, num int, message string) error {
	emsg := model.CommentPrefix + " " + message
	_, resp, err := client.Issues.CreateComment(ctx, r.Owner, r.Name, num, &github.IssueComment{
		Body: github.String(emsg),
	})
	if err != nil {
		return createError(resp, err)
	}
	return nil
}

func (g *Github) ScheduleDeployment(ctx context.Context, u *model.User, r *model.Repo, d model.DeploymentInfo) error {
	client := setupClient(ctx, g.API, u)
	return scheduleDeployment(ctx, client, r, d)
}

func scheduleDeployment(ctx context.Context, client *github.Client, r *model.Repo, d model.DeploymentInfo) error {
	_, resp, err := client.Repositories.CreateDeployment(ctx, r.Owner, r.Name, &github.DeploymentRequest{
		Ref:         github.String(d.Ref),
		Task:        github.String(d.Task),
		Environment: github.String(d.Environment),
	})
	if err != nil {
		return createError(resp, err)
	}
	return nil
}

func doGetPullRequest(ctx context.Context, client *github.Client, r *model.Repo, number int) (model.PullRequest, *string, error) {
	pr, resp, err := client.PullRequests.Get(ctx, r.Owner, r.Name, number)
	if err != nil {
		return model.PullRequest{}, nil, createError(resp, err)
	}

	log.Debug("current issue ==", number)
	log.Debug("current pr ==", *pr)
	sha := pr.Head.SHA

	mergeable := true
	if pr.Mergeable != nil {
		mergeable = *pr.Mergeable
	}

	result := model.PullRequest{
		Issue: model.Issue{
			Number: number,
			Title:  pr.GetTitle(),
			Author: lowercase.Create(pr.User.GetLogin()),
		},
		// head branch contains what you like to be applied
		// base branch contains where changes should be applied
		Branch: model.Branch{
			CompareName:    pr.Head.GetRef(),
			CompareSHA:     pr.Head.GetSHA(),
			CompareOwner:   pr.Head.User.GetLogin(),
			Mergeable:      mergeable,
			Merged:         pr.GetMerged(),
			MergeCommitSHA: pr.GetMergeCommitSHA(),
			BaseName:       pr.Base.GetRef(),
			BaseSHA:        pr.Base.GetSHA(),
		},
		Body: pr.GetBody(),
	}
	return result, sha, nil
}
