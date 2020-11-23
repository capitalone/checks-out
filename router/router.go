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
package router

import (
	"net/http"
	"net/http/pprof"
	rpprof "runtime/pprof"
	"time"

	"github.com/capitalone/checks-out/api"
	"github.com/capitalone/checks-out/envvars"
	"github.com/capitalone/checks-out/router/middleware"
	"github.com/capitalone/checks-out/router/middleware/access"
	"github.com/capitalone/checks-out/router/middleware/header"
	"github.com/capitalone/checks-out/router/middleware/session"
	"github.com/capitalone/checks-out/web"
	"github.com/capitalone/checks-out/web/static"
	"github.com/capitalone/checks-out/web/template"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// Load creates a new HTTP handler
func Load() http.Handler {
	e := gin.New()
	sunlight := envvars.Env.Monitor.Sunlight
	e.Use(middleware.Recovery(sunlight))

	e.SetHTMLTemplate(template.Template())
	e.StaticFS("/static", static.FileSystem())

	e.Use(header.NoCache)
	e.Use(header.Options)
	e.Use(header.Secure)
	e.Use(middleware.Ginrus(logrus.StandardLogger(), time.RFC3339, true))
	e.Use(middleware.Store())
	e.Use(middleware.Remote())
	e.Use(middleware.Cache())
	e.Use(middleware.ExtError())
	if sunlight {
		e.Use(middleware.Version)
	}
	e.Use(session.SetUser)
	e.Use(session.SetCapability)
	e.Use(middleware.BaseURL)

	adminGroup := e.Group("/admin", session.UserMust, api.CheckAdmin)
	adminGroup.GET("config/subtree/*path", api.GetAllConfigurationSubtree)
	adminGroup.POST("slack/:hostname", api.RegisterSlackUrl)
	adminGroup.GET("slack/:hostname", api.GetSlackUrl)
	adminGroup.DELETE("slack/:hostname", api.DeleteSlackUrl)
	adminGroup.PUT("slack/:hostname", api.UpdateSlackUrl)
	adminGroup.DELETE("repos/:owner", api.AdminDeleteOrg)
	adminGroup.DELETE("repos/:owner/:repo", api.AdminDeleteRepo)
	adminGroup.GET("user/:user/repos", api.GetReposForUserLogin)
	adminGroup.GET("stats", api.AdminStats)

	e.GET("/api/user", session.UserMust, api.GetUser)
	e.DELETE("/api/user", session.UserMust, api.DeleteUser)
	e.GET("/api/user/orgs", session.UserMust, api.GetOrgs)
	e.GET("/api/user/repos", session.UserMust, api.GetUserRepos)
	e.GET("/api/user/repos/:org", session.UserMust, api.GetOrgRepos)
	e.GET("/api/user/orgs/enabled", session.UserMust, api.GetEnabledOrgs)

	e.GET("/api/repos/:owner/:repo", session.UserMust, access.RepoPull, api.GetRepo)
	e.POST("/api/repos/:owner/:repo", session.UserMust, access.RepoAdmin, api.PostRepo)
	e.DELETE("/api/repos/:owner/:repo", session.UserMust, access.RepoAdmin, api.DeleteRepo)
	e.GET("/api/repos/:owner/:repo/config", session.UserMust, access.RepoPull, api.GetConfig)
	e.GET("/api/repos/:owner/:repo/maintainers", session.UserMust, access.RepoPull, api.GetMaintainer)
	e.GET("/api/repos/:owner/:repo/validate", session.UserMust, access.RepoPull, api.Validate)
	e.GET("/api/repos/:owner/:repo/lgtm-to-checks-out", session.UserMust, access.RepoPull, api.Convert)

	e.GET("/api/teams/:owner", session.UserMust, api.GetTeams)

	e.GET("/api/pr/:owner/:repo/:id/status", session.UserMust, access.RepoPull, web.ApprovalStatus)

	e.POST("/api/repos/:owner", session.UserMust, access.OwnerAdmin, api.PostOrg)
	e.DELETE("/api/repos/:owner", session.UserMust, access.OwnerAdmin, api.DeleteOrg)

	e.POST("/api/user/slack/:hostname", session.UserMust, api.UserRegisterSlackUrl)
	e.GET("/api/user/slack/:hostname", session.UserMust, api.UserGetSlackUrl)
	e.DELETE("/api/user/slack/:hostname", session.UserMust, api.UserDeleteSlackUrl)
	e.PUT("/api/user/slack/:hostname", session.UserMust, api.UserUpdateSlackUrl)

	if sunlight {
		e.GET("/api/repos", api.GetAllRepos)
		e.GET("/version", web.Version)
		e.GET("/debug/pprof/", gin.WrapF(pprof.Index))
		e.GET("/debug/pprof/cmdline", gin.WrapF(pprof.Cmdline))
		e.GET("/debug/pprof/profile", gin.WrapF(pprof.Profile))
		e.GET("/debug/pprof/symbol", gin.WrapF(pprof.Symbol))
		e.GET("/debug/pprof/trace", gin.WrapF(pprof.Trace))
		for _, p := range rpprof.Profiles() {
			e.GET("/debug/pprof/"+p.Name(), gin.WrapH(pprof.Handler(p.Name())))
		}
	}
	e.GET("/api/count", api.GetAllReposCount)

	e.POST("/hook", web.ProcessHook)
	e.GET("/login", web.Login)
	e.POST("/login", web.LoginToken)
	e.GET("/logout", web.Logout)
	e.NoRoute(web.Index)

	return e
}
