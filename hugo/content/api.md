+++
date = "2017-06-21T10:56:00-04:00"
draft = false
title = "REST API"
weight = 7
menu = "main"
+++

# REST API

The following REST APIs are exposed by checks-out.

## Logging in

checks-out relies on Github's OAuth2 support for authentication. To log in to the REST API, a user must go to /settings/tokens on their Github server (for public Github, this 
would be <https://github.com/settings/tokens>.) and create a Personal Access Token. The token should have the permissions admin:org_hook, read:org, repo:status. Store this
token in a safe place; you cannot get it again from Github.

Endpoint: /login
Method: POST
Query parameters: access_token=PERSONAL_ACCESS_TOKEN
Body: none

Success: returns an AuthToken JSON structure, status = 200
Failure: returns status 401 (unauthorized)
Tokens are currently set to expire after 72 hours.

### AuthToken JSON structure

```json
{
    "access_token": "JWT_TOKEN",
    "expires_in": SECONDS_UNTIL_EXPIRY
}
```

## Authorization

After calling the initial login, you will get back a JWT Token. This token must be supplied to all other authenticated requests using one of the following methods:

- a header with key`Authorization` and value `Bearer JWT_TOKEN`
- a query parameter with key `access_token` and value `JWT_TOKEN`
- a cookie with key `user_sess` and value `JWT_TOKEN`

## Get Current User Info

Endpoint: /api/user
Method: GET

Success: returns a GetCurrentUser JSON structure

### GetCurrentUser JSON Structure

```json
{
    "id": ID,
    "login": "USER_NAME",
    "avatar": "GITHUB_AVATAR_URL_FOR_USER"
}
```

## Delete Current User

This call removes the currently logged-in user from checks-out. If you do this, you will need to recreate your account.

Endpoint: /api/user
Method: DELETE

Success: returns a 204 (deleted) status code

## Get Orgs for Current User

Returns all orgs that the current user belongs to.

Endpoint: /api/user/orgs
Method: GET

Success: returns a JSON list of GetOrg JSON structures

### GetOrg JSON Structure

```json
    {
        "login": "NAME_OF_ORG",
        "avatar": "GITHUB_AVATAR_URL_FOR_ORG",
        "enabled": true|false,
        "admin": true|false
    },
```

If the current user is an admin for the org, `admin` is set to true.
If the org has been enabled in checks-out, `enabled` is set to true.

## Get Personal Repos for Current User

Returns all repos in the user's account (not repos in orgs that the user belongs to)

Endpoint: /api/user/repos
Method: GET

Success: returns a 200 (ok) status code and a JSON list of GetRepo JSON structures

### GetRepo JSON Structure

```json
    {
        "id": ID_IN_checks-out,
        "owner": "USER_NAME",
        "name": "REPO_NAME",
        "slug": "USER_NAME/REPO_NAME",
        "link_url": URL_IN_GITHUB",
        "private": true|false,
        "org": true|false
    },
```

If `id` is present, the repo is managed by checks-out.
If the repo is enabled because its org is managed by checks-out, `org` is set to `true`.
If the repo is private to the user, `private` is set to `true`.

## Get Repos in Specified Org

Returns all repos in the specified org

Endpoint: /api/user/repos/:org
Method: GET

:org is replaced with the name of the org whose repos are being requested.

Success: returns a JSON list of GetRepo JSON structures

## Get Enabled Orgs for Current User

Returns all orgs that have been enabled by the current user

Endpoint: /api/user/orgs/enabled
Method: GET

Success: return a JSON list of GetOrg JSON structures

note: the `avatar` and `admin` fields will always be set to `""` and `false` respectively. the `enabled` field will always be set to `true`.

## Enable Repo

Turns on checks-out monitoring for the specified repo.

Endpoint: /api/repos/:owner/:repo
Method: POST
Body: none

:owner is the name of the org or the name of the user, for a personal repo
:repo is the name of the repo

Success: returns 200 (ok) and a GetRepo JSON structure
Failure: returns 409 (Conflict) if the repo is already enabled
Failure: returns 404 (not found) if the repo does not exist or is not available to the user

## Disable Repo

Turns off checks-out monitoring for the specified repo

Endpoint: /api/repos/:owner/:repo
Method: DELETE

:owner is the name of the org or the name of the user, for a personal repo
:repo is the name of the repo

Success: returns 200 (ok)
Failure: returns 409 (Conflict) if the repo is already disabled
Failure: returns 404 (not found) if the repo does not exist or is not available to the user

## Get Enabled Repo

Returns information on the specified enabled repo

Endpoint: /api/repos/:owner/:repo
Method: GET

:owner is the name of the org or the name of the user, for a personal repo
:repo is the name of the repo

Success: returns 200 (ok) and a GetRepo JSON structure
Failure: returns 404 (not found) if the repo does not exist or is not available to the user

## Enable Org

Turns on checks-out monitoring for the specified org. All repos currently in the org are enabled and all repos added to the org in the future will be enabled on creation.

Endpoint: /api/repos/:owner
Method: POST
Body: none

:owner is the name of the org

Success: returns 200 (ok) and an OrgInfo JSON structure
Failure: returns 409 (Conflict) if the org is already enabled
Failure: returns 404 (not found) if the org does not exist, is not available to the user, or if it is the user's personal repos (owner == user name)

### OrgInfo JSON Structure

```json
{
    "id": ID_IN_checks-out,
    "owner": "ORG_NAME",
    "link_url": "GITHUB_URL",
    "private": true|false
}
```

If the org is private, `private` is set to `true`.

## Disable Org

Turns on checks-out monitoring for the specified org. All repos currently in the org are enabled and all repos added to the org in the future will be enabled on creation.

Endpoint: /api/repos/:owner
Method: DELETE

:owner is the name of the org

Success: returns 200 (ok)
Failure: returns 409 (Conflict) if the org is already disabled
Failure: returns 404 (not found) if the org does not exist, is not available to the user, or if it is the user's personal repos (owner == user name)

## Get Config for Repo

Returns the .lgtm or .checks-out file for the repo

Endpoint: /api/repos/:owner/:repo/config
Method: GET

:owner is the name of the org or the name of the user, for a personal repo
:repo is the name of the repo

Success: returns 200 (ok) and a .checks-out config file in JSON
Failure: returns 404 (not found) if the file does not exist or is not available to the user

## Get Maintainers for Repo

Returns the MAINTAINERS file for the repo

Endpoint: /api/repos/:owner/:repo/maintainers
Method: GET

:owner is the name of the org or the name of the user, for a personal repo
:repo is the name of the repo

Success: returns 200 (ok) and a MAINTAINERS config file
Failure: returns 404 (not found) if the file does not exist or is not available to the user

## Validate Config and Maintainers for Repo

Validates that the .lgtm/.checks-out and MAINTAINERS files are present and valid

Endpoint: /api/repos/:owner/:repo/validate
Method: GET

:owner is the name of the org or the name of the user, for a personal repo
:repo is the name of the repo

Success: returns 200 (ok)
Failure: returns 404 (not found) if the file does not exist or is not available to the user
Failure: returns 400 (bad request) if the .lgtm/.checks-out or MAINTAINERS file is not valid. Error messages returned as text in body of response.

## Convert .lgtm to .checks-out

Converts a .lgtm file in a repo into the equivalent .checks-out file

Endpoint: /api/repos/:owner/:repo/lgtm-to-checks-out
Method: GET

:owner is the name of the org or the name of the user, for a personal repo
:repo is the name of the repo

Success: returns 200 (ok) and a .checks-out config file in JSON
Failure: returns 404 (not found) if the file does not exist or is not available to the user
Failure: returns 400 (bad request) if the .lgtm file is not valid. Error messages returned as text in body of response.

## Get Teams in Org

Returns the teams defined in an org

Endpoint: /api/teams/:owner
Method: GET

:owner is the name of the org

Success: returns 200 (ok) and a JSON array of strings. Each string is the name of a team in the org
Failure: returns 404 (not found) if the org does not exist, is not available to the user, or is the user's personal repos (owner == user name)

## Get checks-out Status for Pull Request

Returns the status of a pull request and the configuration used to process it

Endpoint: /api/pr/:owner/:repo/:id/status
Method: GET

:owner is the name of the org or the name of the user, for a personal repo
:repo is the name of the repo
:id is the id of the pull request in Github

Success: returns 200 (ok) and a Status JSON structure
Failure: returns 404 (not found) if the org, repo, or id does not exist or is not available to the user

### Status JSON Structure

```json
{
    "approved": true|false,
    "approvers": ["USER_NAMES_OF_APPROVERS"],
    "disapprovers": ["USER_NAMES_OF_DISAPPROVERS"],
    "policy": {
    },
    "settings": {
    }
}
```

`approved` is `true` if the pull request has been approved by checks-out, `false` otherwise.
`approvers` is an array of user names of the people who have approved the pull request.
`disapprovers` is an array of user names of the people who have disapproved the pull request.
`policy` is the policy in the .checks-out configuration file that is used for this pull request.
`settings` is the .checks-out configuration file. All optional sections are filled in with their default values.


## User Slack URL Management

### Register new User-Level Slack Target

Registers a new Slack target for all repos managed by the current user. The slack target's name is the hostname for the Slack instance.
A User-Level Slack Target will override an Admin-Level Slack Target for the same hostname. It is not visible by any other users in checks-out.

Endpoint: /api/user/slack/:hostname
Method: POST
Body: URL JSON Structure

:hostname is the hostname for the slack group

Success: returns a 201 (created) status code

#### URL JSON Structure

```json
{
    "url":"_URL_FOR_INCOMING_NOTIFICATIONS_WEBHOOKS_"
}```

### Get URL for User-Level Slack Target

Returns the URL for a User-Level Slack target for all repos managed by the current user. 

Endpoint: /api/user/slack/:hostname
Method: GET

:hostname is the hostname for the slack group

Success: returns a 200 (ok) status code and a URL JSON structure
Failure: returns a 404 (not found) status code if no URL is registered by the current user for the specified slack target

### Update URL for User-Level Slack Target

Updates the specified Slack target for all repos managed by the current user.

Endpoint: /api/user/slack/:hostname
Method: PUT
Body: URL JSON Structure

:hostname is the hostname for the slack group

Success: returns a 200 (created) status code

### Delete User-Level Slack Target

Removes the specified Slack target for all repos managed by the current user. If there is an admin-level Slack target specified for
the same hostname, it will be used instead.

Endpoint: /api/user/slack/:hostname
Method: DELETE

:hostname is the hostname for the slack group

Success: returns a 204 (deleted) status code
Failure: returns a 404 (not found) status code if no URL is registered by the current user for the specified slack target

## Admin

### Get All Enabled Repos

Returns a list of all repos that have been enabled in checks-out across all users

Endpoint: /api/repos
Method: GET

Success: returns a 200 (ok) status code and a JSON list of GetRepo JSON structures


### Get Count of Enabled Repos

Returns a count of all repos that have been enabled in checks-out across all users

Endpoint: /api/count
Method: GET

Success: returns a 200 (ok) status code and a number as text

### Admin Slack URL Management

#### Register New Admin-Level Slack Target

Registers a new Slack target for all repos. The slack target's name is the hostname for the Slack instance.
A User-Level Slack Target will override an Admin-Level Slack Target for the same hostname.

Endpoint: /api/admin/slack/:hostname
Method: POST
Body: URL JSON Structure

:hostname is the hostname for the slack group

Success: returns a 201 (created) status code

### Get URL for Admin-Level Slack Target

Returns the URL for an Admin-Level Slack target for all repos.

Endpoint: /api/admin/slack/:hostname
Method: GET

:hostname is the hostname for the slack group

Success: returns a 200 (ok) status code and a URL JSON structure
Failure: returns a 404 (not found) status code if no URL is registered by the admin for the specified slack target

### Update URL for Admin-Level Slack Target

Updates the specified Slack target for all repos.

Endpoint: /api/admin/slack/:hostname
Method: PUT
Body: URL JSON Structure

:hostname is the hostname for the slack group

Success: returns a 200 (created) status code

### Delete Admin-Level Slack Target

Removes the specified Slack target for all repos. If there are user-level Slack targets specified for
the same hostname, they will be used instead for those users.

Endpoint: /api/admin/slack/:hostname
Method: DELETE

:hostname is the hostname for the slack group

Success: returns a 204 (deleted) status code
Failure: returns a 404 (not found) status code if no URL is registered by the current user for the specified slack target
