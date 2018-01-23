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
	"bytes"
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/capitalone/checks-out/exterror"
	"github.com/capitalone/checks-out/model"
	"github.com/capitalone/checks-out/shared/httputil"
	"github.com/capitalone/checks-out/strings/lowercase"
	"github.com/capitalone/checks-out/usage"

	log "github.com/Sirupsen/logrus"
	"github.com/google/go-github/github"
)

// TODO: move this into its own package when
// we support backends other than GitHub

func createHook(c context.Context, r *http.Request) (Hook, context.Context, error) {

	// For server requests the Request Body is always non-nil
	// but will return EOF immediately when no body is present.
	// The Server will close the request body. The ServeHTTP
	// Handler does not need to.

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, c, err
	}

	event := r.Header.Get("X-Github-Event")
	usage.RecordIncomingWebHook(event)

	var hook Hook
	switch event {
	case "pull_request_review":
		hook, err = createReviewHook(body)
	case "issue_comment":
		hook, err = createCommentHook(body)
	case "status":
		hook, err = createStatusHook(body)
	case "pull_request":
		hook, err = createPRHook(body)
	case "repository":
		hook, err = createRepoHook(r, body)
	}
	if hook != nil {
		hook.SetEvent(event)
	}
	c2 := usage.AddEventToContext(c, event)
	return hook, c2, err
}

func createError(msg string, body []byte, e error) error {
	if e == io.EOF {
		log.Errorf("Logging request body on eof error: %s", body)
	}
	e = exterror.Create(http.StatusInternalServerError, e)
	return exterror.Append(e, msg)
}

func createReviewHook(body []byte) (Hook, error) {

	data := github.PullRequestReviewEvent{}
	err := json.NewDecoder(bytes.NewReader(body)).Decode(&data)
	if err != nil {
		err = createError("Getting pull request review hook", body, err)
		return nil, err
	}

	log.Infof("repository %s pr %d pull_request_review state %s",
		data.Repo.GetFullName(), data.PullRequest.GetNumber(),
		data.PullRequest.GetState())
	// don't process reviews on closed pull requests
	if data.PullRequest.GetState() == "closed" {
		log.Debugf("PR %s is closed -- not processing comments for it any more", data.PullRequest.Title)
		return nil, nil
	}

	hook := &ReviewHook{
		ApprovalHook: ApprovalHook{
			Issue: &model.Issue{
				Title:  data.PullRequest.GetTitle(),
				Number: data.PullRequest.GetNumber(),
				Author: lowercase.Create(data.PullRequest.User.GetLogin()),
			},
			Repo: &model.Repo{
				Owner: data.Repo.Owner.GetLogin(),
				Name:  data.Repo.GetName(),
				Slug:  data.Repo.GetFullName(),
			},
		},
	}

	return hook, nil
}

func createCommentHook(body []byte) (Hook, error) {

	data := github.IssueCommentEvent{}
	err := json.NewDecoder(bytes.NewReader(body)).Decode(&data)
	if err != nil {
		err = createError("Getting comment hook", body, err)
		return nil, err
	}

	// don't process comments on GitHub issues
	if len(data.Issue.PullRequestLinks.GetURL()) == 0 {
		return nil, nil
	}

	log.Infof("repository %s pr %d issue_comment state %s",
		data.Repo.GetFullName(), data.Issue.GetNumber(),
		data.Issue.GetState())
	// don't process comments on closed pull requests
	if data.Issue.GetState() == "closed" {
		log.Debugf("PR %s is closed -- not processing comments for it any more", data.Issue.Title)
		return nil, nil
	}

	hook := &CommentHook{
		ApprovalHook: ApprovalHook{
			Issue: &model.Issue{
				Title:  data.Issue.GetTitle(),
				Number: data.Issue.GetNumber(),
				Author: lowercase.Create(data.Issue.User.GetLogin()),
			},
			Repo: &model.Repo{
				Owner: data.Repo.Owner.GetLogin(),
				Name:  data.Repo.GetName(),
				Slug:  data.Repo.GetFullName(),
			},
		},
		Comment: data.Comment.GetBody(),
	}

	return hook, nil
}

func createStatusHook(body []byte) (Hook, error) {

	data := github.StatusEvent{}
	err := json.NewDecoder(bytes.NewReader(body)).Decode(&data)
	if err != nil {
		err = createError("Getting status hook", body, err)
		return nil, err
	}

	log.Infof("repository %s status commit %s",
		data.Repo.GetFullName(),
		data.GetSHA())
	log.Debug(data)

	hook := &StatusHook{
		SHA: data.GetSHA(),
		Status: &model.CommitStatus{
			State:       data.GetState(),
			Context:     data.GetContext(),
			Description: data.GetDescription(),
		},
		Repo: &model.Repo{
			Owner: data.Repo.Owner.GetLogin(),
			Name:  data.Repo.GetName(),
			Slug:  data.Repo.GetFullName(),
		},
	}

	return hook, nil
}

func createPRHook(body []byte) (Hook, error) {

	data := github.PullRequestEvent{}
	err := json.NewDecoder(bytes.NewReader(body)).Decode(&data)
	if err != nil {
		err = createError("Getting pull request hook", body, err)
		return nil, err
	}

	log.Debug(data)

	log.Infof("repository %s pr %d pull_request action %s state %s",
		data.Repo.GetFullName(), data.PullRequest.GetNumber(),
		data.GetAction(),
		data.PullRequest.GetState())

	mergeable := true
	if data.PullRequest.Mergeable != nil {
		mergeable = data.PullRequest.GetMergeable()
	}

	hook := &PRHook{
		ApprovalHook: ApprovalHook{
			HookCommon: HookCommon{
				ActionType: data.GetAction(),
			},
			Issue: &model.Issue{
				Title:  data.PullRequest.GetTitle(),
				Number: data.PullRequest.GetNumber(),
				Author: lowercase.Create(data.PullRequest.User.GetLogin()),
			},
			Repo: &model.Repo{
				Owner: data.Repo.Owner.GetLogin(),
				Name:  data.Repo.GetName(),
				Slug:  data.Repo.GetFullName(),
			},
		},
		PullRequest: &model.PullRequest{
			Issue: model.Issue{
				Number: data.PullRequest.GetNumber(),
				Title:  data.PullRequest.GetTitle(),
				Author: lowercase.Create(data.PullRequest.User.GetLogin()),
			},
			// head branch contains what you like to be applied
			// base branch contains where changes should be applied
			Branch: model.Branch{
				CompareName:    data.PullRequest.Head.GetRef(),
				CompareSHA:     data.PullRequest.Head.GetSHA(),
				CompareOwner:   data.PullRequest.Head.User.GetLogin(),
				Mergeable:      mergeable,
				Merged:         data.PullRequest.GetMerged(),
				MergeCommitSHA: data.PullRequest.GetMergeCommitSHA(),
				BaseName:       data.PullRequest.Base.GetRef(),
				BaseSHA:        data.PullRequest.Base.GetSHA(),
			},
			Body: data.PullRequest.GetBody(),
		},
	}

	return hook, nil
}

func createRepoHook(r *http.Request, body []byte) (Hook, error) {

	data := github.RepositoryEvent{}
	err := json.NewDecoder(bytes.NewReader(body)).Decode(&data)
	if err != nil {
		err = createError("Getting repository hook", body, err)
		return nil, err
	}

	log.Debug(data)

	hook := &RepoHook{
		HookCommon: HookCommon{
			ActionType: data.GetAction(),
		},
		Name:    data.Repo.GetName(),
		Owner:   data.Repo.Owner.GetLogin(),
		BaseURL: httputil.GetURL(r),
	}

	return hook, nil
}
