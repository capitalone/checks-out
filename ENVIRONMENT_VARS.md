# Environment Variables for Checks-Out

This document contains information about all the environment variables that can or
must be defined for Checks-Out to function.

## Server configuration

### Server IP Address and Port
- Format: `SERVER_ADDR="_ip_address_:_port_"` but more specifically see the documentation for
golang's `http.ListenAndServe` and `http.ListanAndServeTLS`
- Default: `:8000`, which means port 8000 on all interfaces
- Required: No

See https://golang.org/pkg/net/http/#ListenAndServe for detailed documentation on the format.

    
### Server SSL Certificate
- Format: `SERVER_CERT="_full_path_and_filename_of_ssl_server_cert_file_"`
- Default: None
- Required: No

If SERVER_CERT is not specified, Checks-Out runs without SSL.  See https://golang.org/pkg/net/http/#ListenAndServeTLS
for detailed documentation on the format.

### Server SSL Key
- Format: `SERVER_KEY="_full_path_and_filename_of_ssl_server_key_file_"`
- Default: None
- Required: _Only if `SERVER_CERT` is also specified_

If SERVER_CERT is specified, then SERVER_KEY must be specified as well.  See https://golang.org/pkg/net/http/#ListenAndServeTLS
for detailed documentation on the format.

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

### URL for Github API
- Format: `GITHUB_URL="_protocol_plus_hostname_plus_path_prefix_of_url_"`
- Default: `https://github.com`
- Required: No, defaults to `https://github.com` which is fine unless you are using your own
enterprise github

### Github OAuth2 Client ID
- Format: `GITHUB_CLIENT="_your_OAuth2_client_id_"`
- Default: None
- Required: Yes.  You must supply the github OAuth2 client ID issued by the github server you are
connecting to (github.com or enterprise-hosted github server)

### Github OAuth2 Secret
- Format: `GITHUB_SECRET="_your_OAuth2_secret_"`
- Default: None
- Required: Yes.  You must supply the github OAuth2 secret issued by the github server you are
connecting to (github.com or enterprise-hosted github server)

### Github Scope To Use
- Format: `GITHUB_SCOPE="_valid_github_scope_specification_"`.  Please see github documentation for
specifics on the specification format
- Default: `read:org,repo:status,admin:repo_hook`
- Required: No

### Github testing-only settings

#### Enable Github Integration Tests
- Format: `GITHUB_TEST_ENABLE=true|false`.  Enable Github integration tests to run
- Default: False
- Required: No

#### Github Integration Tests OAuth2 Token
- Format: `GITHUB_TEST_TOKEN=__OAUTH2_TOKEN_TO_USE`.  This is an OAuth2 token to be sent to the
github endpoint for github integration test API requests
- Default: None
- Required: _only if GITHUB_TEST_ENABLE is true_

## Logging/Debug

### Debug Logging
- Format: `DEBUG=true|false`
- Default: false
- Required: No

If set to `true`, debug logging will be enabled and output to logging.

### Checks-Out Sunlight
- Format: `CHECKS_OUT_SUNLIGHT=true|false`
- Default: false
- Required: No

If set to true, exposes endpoints and data that might not be suitable for a live site.  Specifically,
the following behaviors are enabled:
- `/api/repos` endpoint is available that will show all repositories managed under Checks-Out
- `/version` endpoint is available that will show the Checks-Out version
- Outputs stack trace information over HTTP when errors happen
- Puts the `X-CHECKS-OUT-VERSION` HTTP header with the Checks-Out version in every response

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
of Checks-Out such as number of commits, approvers, and disapprovers in the specified time period.

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

Determines the length of time Checks-Out will cache github artifacts like user information,
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
- Default: `https://capitalone.github.com/checks-out/docs`
- Required: No

Provides the base URL for links to the documentation in the UI.
