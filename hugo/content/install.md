+++
date = "2015-12-05T16:00:21-08:00"
draft = false
title = "Self Hosting"
weight = 6
menu = "main"
toc = true
+++

# Requirements

checks-out ships as a single binary file. If you are planning on integrating
with GitHub Enterprise it requires version 2.4 or higher.

# Configuration

This is a full list of configuration options. Please note that many of these
options use default configuration value that should work for the majority of
installations.

# Environment Variables for checks-out

This document contains information about all the environment variables that can or
must be defined for checks-out to function.

## Server configuration

### Server IP Address and Port

- Format: `SERVER_ADDR="_ip_address_:_port_"` but more specifically see the documentation for golang's `http.ListenAndServe` and `http.ListanAndServeTLS`
- Default: `:8000`, which means port 8000 on all interfaces
- Required: No

See <https://golang.org/pkg/net/http/#ListenAndServe> for detailed documentation on the format.

### Server SSL Certificate

- Format: `SERVER_CERT="_full_path_and_filename_of_ssl_server_cert_file_"`
- Default: None
- Required: No

If SERVER_CERT is not specified, checks-out runs without SSL.  See <https://golang.org/pkg/net/http/#ListenAndServeTLS>
for detailed documentation on the format.

### Server SSL Key

- Format: `SERVER_KEY="_full_path_and_filename_of_ssl_server_key_file_"`
- Default: None
- Required: _Only if `SERVER_CERT` is also specified_

If SERVER_CERT is specified, then SERVER_KEY must be specified as well.  See <https://golang.org/pkg/net/http/#ListenAndServeTLS> for detailed documentation on the format.

## Database configuration

### Database Driver

- Format: `DB_DRIVER=sqlite3|postgres|mysql`
- Default: None
- Required: Yes

### Database Datasource

- Format: `DB_SOURCE="_db_driver_specific_datasource_spec_"`
- Default: None
- Required: Yes

Please refer to the datasource specifications for the sqlite3, lib/pq, and mysql drivers for their respective
specifications for this environment variable.

## Slack integration

- Format: `SLACK_TARGET_URL="_slack_integration_url"`
- Default: None
- Required: No

If the `SLACK_TARGET_URL` is not defined, then no logging into slack will happen, however it will also
currently cause logging that Slack is not configured to get generated every time a slackable event happens.

## Github integration

### Email Address To Use for Github

- Format: `GITHUB_EMAIL="_valid_email_address_"`
- Default: None
- Required: Yes

Email address for git commits.

### URL for Github API

- Format: `GITHUB_URL="_protocol_plus_hostname_plus_path_prefix_of_url_"`
- Default: `https://github.com`
- Required: No, defaults to `https://github.com` which is fine unless you are using your own enterprise github

### Github OAuth2 Client ID

- Format: `GITHUB_CLIENT="_your_OAuth2_client_id_"`
- Default: None
- Required: Yes.  You must supply the github OAuth2 client ID issued by the github server you are connecting to (github.com or enterprise-hosted github server)

### Github OAuth2 Secret

- Format: `GITHUB_SECRET="_your_OAuth2_secret_"`
- Default: None
- Required: Yes.  You must supply the github OAuth2 secret issued by the github server you are connecting to (github.com or enterprise-hosted github server)

### Github Scope To Use

- Format: `GITHUB_SCOPE="_valid_github_scope_specification_"`.  Please see github documentation for specifics on the specification format
- Default: `read:org,repo:status,admin:repo_hook,admin:org_hook`
- Required: No

Permissions granted to checks-out by Github. The minimum
required permissions are the default ones.

### Github rate limiting

- Format: `GITHUB_BATCH_PER_SECOND=int`
- Default: 10
- Required: no

GitHub batch access rate limiter. For certain calls that might happen in rapid succession, this limits how
fast those calls are made to the server. The value is in Hertz; the default value of 10 means that you cannot
send more than 10 calls a second to Github.

### Github testing-only settings

#### Enable Github Integration Tests

- Format: `GITHUB_TEST_ENABLE=true|false`
- Default: False
- Required: No

Enable Github integration tests to run

#### Github Integration Tests OAuth2 Token

- Format: `GITHUB_TEST_TOKEN=__OAUTH2_TOKEN_TO_USE`
- Default: None
- Required: _only if GITHUB_TEST_ENABLE is true_

 This is an OAuth2 token to be sent to the github endpoint for github integration test API requests

## Logging/Debug

### Debug Logging

- Format: `LOG_LEVEL=debug|info|warn|error|fatal|panic`
- Default: info
- Required: No

Specifies the default logging level used in checks-out.

### Checks-out Sunlight

- Format: `CHECKS_OUT_SUNLIGHT=true|false`
- Default: false
- Required: No

If set to true, exposes endpoints and data that might not be suitable for a live site.  Specifically,
the following behaviors are enabled:

- `/api/repos` endpoint is available that will show all repositories managed under checks-out
- `/version` endpoint is available that will show the checks-out version
- Outputs stack trace information over HTTP when errors happen
- Puts the `X-CHECKS-OUT-VERSION` HTTP header with the checks-out version in every response

### Blacklist User Agents from Logging

- Format: `BLACKLIST_USER_AGENTS=user_agent1[]:user_agent2...:user_agent_N]`
- Default: false
- Required: No

Specify a colon-separated list of user agent strings for the middleware that does request logging
to skip logging on.  This allows suppressing logging of user agents like the health checker from aws
that are generally operational noise.

### How Frequently To Log Operational Statistics

- Format: `LOG_STATS_PERIOD=_valid_time.ParseDuration()_string_`
- Default: 0 (Do not periodically log)
- Required: No

Specify a time duration in a valid format understood by the
[time.ParseDuraction() method](https://golang.org/pkg/time/#ParseDuration) to periodically log activity
of checks-out such as number of commits, approvers, and disapprovers in the specified time period.

## Caching

### Response caching

- Format: `CACHE_TTL=_time_specified_in_time.Duration_format_`
- Default: 15 minutes
- Required: No

Determines the length of time the gin middleware will cache.  Default is 15 minutes.

### Github Repo Artifact Caching

- Format: `LONG_CACHE_TTL=_time_specified_in_time.Duration_format_`
- Default: 24 hours
- Required: No

Determines the length of time checks-out will cache github artifacts like user information,
organization members, and team members in memory before going back to the server.  Default
is 24 hours.

## Account Creation Control

### Limit User Access

- Format: `LIMIT_USERS=true|false`
- Default: false
- Required: No

If enabled, only users listed in the `limit_users` table in the database are 
allowed to create accounts. Any existing accounts will still function, even 
if the user's name is not in the `limit_users` table. 

### Limit Organizational Access

- Format: `LIMIT_ORGS=true|false`
- Default: false
- Required: No

If enabled, only users in the organizations listed in the `limit_orgs` table 
in the database are allowed to create accounts. Any existing accounts will 
still function, even if the user is not any of the orgs named in the 
`limit_orgs` table. 

If both `LIMIT_ORGS` and `LIMIT_USERS` are set to `true` then a user can
either be explicitly named in `limit_users` or can belong to an org named
in `limit_orgs`.

## Miscellaneous Properties

### Documentation Location

- Format: `CHECKS_OUT_DOCS_URL=_URL_for_docs_no_closing_slash_`
- Default: `https://www.capitalone.io/checks-out`
- Required: No

Provides the base URL for links to the documentation in the UI.

### Slack Integration

- Format: `SLACK_TARGET_URL=url`
- Default: none
- Required: no

Provides the default Slack notification url. This is the URL that's used by defaut for slack integration when the
target specified in the .checks-out file is "slack".

### Admin management

- Format: `GITHUB_ADMIN_ORG=_github_org_name_`
- Default: none
- Required: no

Specifies the Github organization whose members have admin privleges in checks-out. If this is not specified,
there are no admin users in checks-out. Admin users have access to certain REST API endpoints that other users
do not.

### Template Repo Name

- Format: `ORG_REPOSITORY_NAME=_name_for_template_repo_per_org_`
- Default: checks-out-configuration
- Required: No

Org management repo name. This is the name of the repo in an org that is used to hold the default .checks-out and
MAINTAINERS files for all repos in the org.

## Legacy Properties

The following environment variables exist to provide system-wide defaults for repos that are still using a 
.lgtm file. They are considered obsolete

- Format: `CHECKS_OUT_APPROVALS=_number_of_approvals_`
- Default: 2
- Required: No

Legacy default number of approvals

- Format: `CHECKS_OUT_PATTERN=_regex_for_approval_comments_`
- Default: (?i)LGTM
- Required: No

Legacy matching pattern

- Format: `CHECKS_OUT_SELF_APPROVAL_OFF=true|false`
- Default: false
- Required: No

Legacy self-approval behavior. Set to true to disable self-approvals.

# Registration

Register your application with GitHub (or GitHub Enterprise) to create your client
id and secret. It is very import that the redirect URL matches your http(s) scheme
and hostname exactly with `/login` as the path.

Please use this screenshot for reference:

![github registration](/images/app_registration.png)

# Reverse Proxies

If you are running behind a reverse proxy please ensure the `X-Forwarded-For`
and `X-Forwarded-Proto` variables are configured.

This is an example nginx configuration:

```nginx
location / {
    proxy_set_header X-Forwarded-For $remote_addr;
    proxy_set_header X-Forwarded-Proto $scheme;
    proxy_set_header Host $http_host;
	proxy_pass http://127.0.0.1:8000;
}
```

This is an example caddy server configuration:

```nginx
checks-out.mycomopany.com {
        proxy / localhost:8000 {
                proxy_header X-Forwarded-Proto {scheme}
                proxy_header X-Forwarded-For {host}
                proxy_header Host {host}
        }
}
```

Note that when running behind a reverse proxy you should change the recommended
port mappings from `--publish=80:8000` to something like `--publish=8000:8000`.
