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
package web

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/capitalone/checks-out/envvars"
	"github.com/capitalone/checks-out/exterror"
	"github.com/capitalone/checks-out/model"
	"github.com/capitalone/checks-out/remote"
	"github.com/capitalone/checks-out/shared/httputil"
	"github.com/capitalone/checks-out/shared/token"
	"github.com/capitalone/checks-out/store"
	"github.com/gin-gonic/gin"
)

// Login attempts to authorize a user via GitHub oauth2. If the user does not
// yet exist, and new account is created. Upon successful login the user is
// redirected to the main screen.
func Login(c *gin.Context) {
	// render the error page if the login fails. Without this block
	// we would encounter an infinite number of redirects.
	if err := c.Query("error"); len(err) != 0 {
		c.HTML(500, "error.html", gin.H{"error": err})
		return
	}

	// when dealing with redirects we may need
	// to adjust the content type. I cannot, however,
	// remember why, so need to revisit this line.
	c.Writer.Header().Del("Content-Type")

	tmpuser, err := remote.GetUser(c, c.Writer, c.Request)
	if err != nil {
		log.Warnf("cannot authenticate user. %s", err)
		c.Redirect(303, "/login?error=oauth_error")
		return
	}
	// this will happen when the user is redirected by
	// the remote provide as part of the oauth dance.
	if tmpuser == nil {
		return
	}

	// get the user from the database
	u, err := store.GetUserLogin(c, tmpuser.Login)
	if err != nil && err != sql.ErrNoRows {
		c.HTML(500, "error.html", gin.H{"error": err})
		return
	} else if err == sql.ErrNoRows {
		err = validateUserAccess(c, tmpuser)
		if err != nil {
			log.Warnf("cannot create account for user. %s", tmpuser.Login, err)
			c.Redirect(303, "/login?error=no_access_error")
			return
		}

		// create the user account
		u = &model.User{}
		u.Login = tmpuser.Login
		u.Token = tmpuser.Token
		u.Avatar = tmpuser.Avatar
		u.Scopes = tmpuser.Scopes
		u.Secret = model.Rand()

		// insert the user into the database
		if err = store.CreateUser(c, u); err != nil {
			log.Errorf("cannot insert %s. %s", u.Login, err)
			c.Redirect(303, "/login?error=internal_error")
			return
		}
	}

	// update the user meta data and authorization
	// data and cache in the datastore.
	u.Token = tmpuser.Token
	u.Avatar = tmpuser.Avatar

	if err = store.UpdateUser(c, u); err != nil {
		log.Errorf("cannot update %s. %s", u.Login, err)
		c.Redirect(303, "/login?error=internal_error")
		return
	}

	exp := time.Now().Add(time.Hour * 72).Unix()
	sessToken := token.New(token.SessToken, u.Login)
	tokenstr, err := sessToken.SignExpires(u.Secret, exp)
	if err != nil {
		log.Errorf("cannot create token for %s. %s", u.Login, err)
		c.Redirect(303, "/login?error=internal_error")
		return
	}

	httputil.SetCookie(c.Writer, c.Request, "user_sess", tokenstr)
	c.Redirect(303, "/")
}

func validateUserAccess(c context.Context, u *model.User) error {
	if !envvars.Env.Access.LimitUsers && !envvars.Env.Access.LimitOrgs {
		return nil
	}
	checkedUsers := false
	if envvars.Env.Access.LimitUsers {
		checkedUsers = true
		valid, err := store.CheckValidUser(c, u.Login)
		if err != nil || valid {
			return err
		}
	}
	checkedOrgs := false
	if envvars.Env.Access.LimitOrgs {
		checkedOrgs = true
		orgs, err := remote.GetOrgs(c, u)
		if err != nil {
			return err
		}
		validOrgs, err := store.GetValidOrgs(c)
		if err != nil {
			return err
		}
		for _, v := range orgs {
			if validOrgs.Contains(v.Login) {
				return nil
			}
		}
	}
	msg := "User %s is not on "
	if checkedUsers {
		msg += "the allowed users "
		if checkedOrgs {
			msg += " or "
		}
	}
	if checkedOrgs {
		msg += "the allowed orgs "
	}
	msg += "list."
	return fmt.Errorf(msg, u.Login)
}

// LoginToken authenticates a user with their GitHub token and
// returns a service API token in the response.
func LoginToken(c *gin.Context) {
	access := c.Query("access_token")
	login, err := remote.GetUserToken(c, access)
	if err != nil {
		c.Error(exterror.Append(err, "Unable to authenticate user"))
		return
	}
	user, err := store.GetUserLogin(c, login)
	if err != nil {
		c.Error(exterror.Create(http.StatusUnauthorized, errors.New("Unable to authenticate user")))
		return
	}
	exp := time.Now().Add(time.Hour * 72).Unix()
	userToken := token.New(token.UserToken, user.Login)
	tokenstr, err := userToken.SignExpires(user.Secret, exp)
	if err != nil {
		log.Errorf("cannot create token for %s. %s", user.Login, err)
		err = errors.New("Internal error attempting to authenticate user")
		c.Error(exterror.Create(http.StatusInternalServerError, err))
		return
	}
	c.IndentedJSON(http.StatusOK, &tokenPayload{
		Access:  tokenstr,
		Expires: exp - time.Now().Unix(),
	})
}

type tokenPayload struct {
	Access  string `json:"access_token,omitempty"`
	Refresh string `json:"refresh_token,omitempty"`
	Expires int64  `json:"expires_in,omitempty"`
}

// Logout terminates the session for the currently authenticated user,
// deleting all session cookies, and redirecting back to the main page.
func Logout(c *gin.Context) {
	httputil.DelCookie(c.Writer, c.Request, "user_sess")
	c.HTML(200, "logout.html", gin.H{})
}
