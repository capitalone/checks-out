Checks-out is a fork of [LGTM](https://github.com/lgtmco/lgtm). It was forked
at 4cbae5. Checks-out uses [semantic versioning](http://semver.org/). The
Checks-out configuration format is incompatible with the LGTM configuration
format but the legacy format can be parsed.

# 0.28.0

* Apply fix for https://github.com/google/go-github/issues/664.
Allows users to onboard to checks-out when GitHub Reviews is
enabled on the repository.

# 0.27.0

* Change default behavior of 'ignoreuimerge' from false to true.

# 0.26.0

* Reduce scope of 'authoraffirm' feature. The message will only appear
when the pull request has multiple committers AND there are approvers
who are also committers. If the approvers are not committers then no
error message appears. When the error message appears, we have changed
the wording of the message posted to the pull request.
* Improvement to branch deletion protected that was added in 0.25.0.
If deletion of the compare branch is enabled then the final approval
policy in the policy array must either have a match of "off" or have
deletion disabled in the policy.
* Fix several 500-level http responses that should be 400-level
responses.

# 0.25.0

* Add 'authoraffirm' feature when multiple committers on pull request.
Pull request author must approve the pull request when there are committers
on the pull request other than the pull request author. This is
enabled by default.
* Enhanced logging of GitHub webhook events
* Lazy loading of teams for github-team repo-self. Teams are only loaded
when they are needed by the approval policy in the .checks-out file.
This will fix the triggering of GitHub rate limiters for organizations
with many teams.
* github-team repo-self [orgname] support. It is now possible to load
all the teams from another organization.
* Fix docker tag validation

# 0.23.0

* Improved error message when configuration or MAINTAINERS file is missing.
* Improved logging when GitHub hook sends an empty request body.
* Enabling a repository requires successful validation of the configuration files.
* Bug fix for ignoring GitHub pull request comments created by the service.
* Bug fix for organizations in the user interface. If another admin
  of the organization has enabled it, then you will see the organization
  enabled. Previously a user only saw the organizations they had 
  enabled themselves.

# 0.22.0

* Bug fix for case-insensitive team name matching for repositories
  using the HJSON format for MAINTAINERS file.
* Log request headers on 500-level errors. Will allow us to
  debug intermittent GitHub status hooks with empty request body.

# 0.21.0

* Case-insensitive team name matching. Previously only user names
  were case-insensitive.
* Allow users and groups to be specified in anonymous match policy.
  So "{foo,bar}\[count=2\]" foo and bar can either be a group or
  a user name. Previously only usernames were accepted. **Breaking
  change**: it is now a configuration error to have a user name
  and group name with the same name.
* Eliminate default approval policy. Previously you could
  omit the approvals list from your configuration file and a default
  policy would be applied. This behavior can lead to silent
  misconfiguration of the configuration file. It is now a configuration
  error if no approval policy is specified. 
  **Breaking change**: If your configuration file has no approval policy
  it is now a configuration error. You must specify an approval policy.
* Add 'uptodate' option to merge configuration. When auto-merge is enabled,
  the uptodate feature will require the compare branch to have merged in
  all commits from the base branch before automatic merge is performed.
  Default value for this option is true.

# 0.20.0

* New admin endpoints to manually deregister github repositories.
* Allow multiple Slack servers to be used as endpoints.

# 0.13.0

* Use "github-collab \[parameter\]" notation in MAINTAINERS file.
  parameter is either the literal "repo-self" or the "org/repo" slug
  of another repository.
* A repository named “checks-out-configuration” in your organization
  is special. The files template.checks-out and template.MAINTAINERS
  in that repository will be used by any other repository that
  is missing those files.
* user interface has the ability to register or de-register
  an entire organization. If an organization is registered then
  any new repositories created in the organization are automatically
  added to the service.
* Approval scope has new parameters 'regexpaths', 'regexbase',
  and 'regexcompare'. These are regular expression matches against
  the file paths of the pull request, the base branch name,
  or the compare branch name, respectively.

# 0.12.0

* Skipped by accident

# 0.11.0

* GitHub Reviews integration. Existing repositories must de-register and re-register
* Bugfix for GitHub API library calls that return an error without a HTTP response

# 0.10.0

* github-org repo-self when used in a personal repository evaluates to the owner
* Implement "ignoreuimerge" feature to ignore merges from the user interface
* Allow "comment:" to provide a comment on the git commit when a pull request passes status checks
* Add shadowed variable checking to 'go vet' Makefile directive
* Bugfix for users with no user repositories and no group repositories
* Truncate GitHub status messages that are too long (140 character limit)
* Fix typo in reading environment variables
* Change github-team [name] to case insensitive comparison of slugs

# 0.9.4

* Bugfix for users with no user repositories and no group repositories

# 0.9.3

* Truncate GitHub status messages that are too long (140 character limit)

# 0.9.2

* Fix typo in reading environment variables

# 0.9.1

* Change github-team [name] to case insensitive comparison of slugs

# 0.9.0

* Pull request commit status message shows error when configuration file is misconfigured
* Update dependencies to ensure that go-yaml has Apache 2.0 license
* Add integration end-to-end testing of checks-out service
* Bug fix for logging severity level of login errors
* Improved error reporting when OAuth capabilities have insufficient privileges
* Remove unused email address column from the database
* Bug fix to eliminate spurious warning level log message
* Refactoring of environment variables into one package
* Add "./checks-out -help" and "./checks-out -env" command-line options
* Parse legacy .lgtm configuration file format
* Add button in user interface for a user to deregister from service
* Ability to use a GitHub team from another organization "github-team {team} {org}"
* Ability to change the location of the documentation via an environment variable
* Ability to limit which users can create accounts via environment variables and user/org lists in the database
* Ability to parse the legacy MAINTAINERS toml file format
* Bug fix generate ERROR severity log message when recovering from panic

# 0.8.0

* Http error response codes that are less than 500 use the warning log level.
* Bug fix. Generate 404 response when repo is not found in database.
* Add 'authormatch' section to approval policies. Can be used for Contributer License Agreements.
* Internal user agent blacklist when logging http requests.
* Update user interface with on/off toggle for activation.
* Add validation button to user interface.
* User case insensitive matching for GitHub usernames. ([#85](https://github.com/capitalone/checks-out/issues/85))
* Ability to change commit ranges for disapproval and tag selection. ([#80](https://github.com/capitalone/checks-out/issues/80))
* 'antititle' optional regex to prevent merges based on pull request title ([#39](https://github.com/capitalone/checks-out/issues/39))
* Finegrain notification types as GitHub comments or Slack messages
* Postgres database backend support

# 0.7.0

* Add new approval policy 'off' to disable service
* Add optional 'merge' section for each approval scope. ([#66](https://github.com/capitalone/checks-out/issues/66))
* Simplify disapproval configuration. Add 'antimatch' to configuration.
* Add 'github-team repo-self' notation. ([#60](https://github.com/capitalone/checks-out/issues/60))
* Show all errors in response when parsing configuration/MAINTAINERS files
* When possible collapse multierrors to a single HTTP response code.
* Pagination for GitHub API requests. Fixes bug for getting team members.
* Status messages with approval information.
* Populate name and email address for GitHub orgs and teams. ([#57](https://github.com/capitalone/checks-out/issues/57))
* Set log level to Info on release build.
* Set GIN_MODE to release mode on release build.

# 0.6.0

* Add 'antipattern' to configuration. Specifies disapproval regular expression.
* Added 'range' option to commit section of configuration.
* Added 'delete' option to merge section of configuration.
* Remove private repositories from /api/repos endpoint. Add /api/count endpoint that returns a count of all public and private repos.

# 0.5.16

The following is a summary of the changes between the LGTM project
and this project.

* The configuration file has changed from a .lgtm file in TOML format to
a configuration file in Human JSON ([HJSON](http://hjson.org/)) format.
* The MAINTAINERS file can either be specified in the original text
format or in HJSON format.
* The default behavior when reading the comments on a pull request has
been changed to only consider all comments since the HEAD commit on the
branch. This behavior can be changed to consider all comments on the
pull request with a configuration parameter. Fixes [#40](https://github.com/lgtmco/lgtm/issues/40).
* Add custom approval policies based on a grammar that tests groups of people against specified thresholds. Fixes [#32](https://github.com/lgtmco/lgtm/issues/32), [#35](https://github.com/lgtmco/lgtm/issues/35).
* Add ability to auto-merge a pull request after status checks have passed.
* Add ability to tag the base branch after an auto-merge has completed.
* Add custom approval scopes that can be based on either file paths in the
repository or against a list of base branches. Fixes [#20](https://github.com/lgtmco/lgtm/issues/20).
* Added LGTM_SUNLIGHT environment variable. Can be enabled to provide APIs that might have security or privacy concerns when publicly accessible.
* Improved error messages in API responses for misconfigured git repositories.
