+++
date = "2015-12-05T16:00:21-08:00"
draft = false
title = "Customize"
weight = 4
menu = "main"
toc = true
+++

# Introduction

Place an `.checks-out` file in the root of your repository to customize your
project's approval process. This file is in Human JSON
([HJSON](http://hjson.org/)) format. JSON is also valid HJSON so
you can write your .checks-out file in JSON if you prefer.

Each property of the .checks-out file has a default value that is used
when you do not provide a value. This document will explain each
section of the .checks-out file. At the beginning of each section an
example of the configuration will be provided. The example will
be populated with the default values for that section.

For example here are all the default values of the .checks-out file:

```json
approvals:
[
  {
    scope:
    {
      branches: []
      paths: []
    }
    name: ""
    match: "all[count=1,self=true]"
    antimatch: "all[count=1,self=true]"
    authormatch: "universe[count=1,self=true]"
    pattern: null
    antipattern: null
    antititle: null
    tag: null
    merge: null
    feedback: null
  }
]
pattern: "(?i)^I approve\\s*(version:\\s*(?P<version>\\S+))?\\s*(comment:\\s*(?P<comment>.*\\S))?\\s*"
antipattern: null
antititle: null
commit:
{
  range: head
  antirange: head
  tagrange: head
  ignoreuimerge: false
}
maintainers:
{
  path: MAINTAINERS
  type: text
}
feedback:
{
  type: ["comment", "review"]
}
merge:
{
  enable: false
  merge: "merge"
  delete: false
  uptodate: true
}
tag:
{
  enable: false
  algorithm: semver
  template: "{{.Version}}"
  increment: "patch"
  docker: false
}
comment:
{
  enable: false
  targets: []
}
deploy:
{
  enable: false
  deployment: DEPLOYMENTS
}
```

If you do not need to change the default values in a section
of the .checks-out then you may leave out that section from
the file. Here is a valid .checks-out file where we are changing
the number of approvers from 1 to 2 and the remaining properties
use the default values:

```json
approvals:
[
  {
    match: "all[count=2]"
  }
]
```

You can inspect the .checks-out file for your repository using the api endpoint
`/api/repos/[orgname]/[reponame]/config`. The output will have any missing
configuration parameters populated with default values. If the configuration
file cannot be parsed then an error message will be returned.

# Approval Policies

checks-out has extensive support for different approval policies. Each approval
policy consists of two sections: the approval scope and the approval match.
The scope determines which policy is applied against a pull request.
The match determines when the pull request is allowed to merge. The
approvals section of .checks-out is an ordered list (an array) of policies.
The policies are traversed in the order specified by the configuration file.
The first scope that matches against the pull request is the policy
that is used (see Policy Scope section).

Each approval policy may optionally declare a tag section. The tag
section is identical in structure to the global tag section. If a policy
has a tag section then its tag section is applied instead of the global
tag section.

Each approval policy may optionally declare a merge section. The merge
section is identical in structure to the global merge section. If a policy
has a merge section then its merge section is applied instead of the global
merge section.

Each approval policy may optionally declare a pattern. If a policy has a pattern
then it is is applied for regular expression matching instead of the global
pattern. Each approval policy may optionally declare an antipattern. If a policy
has a antipattern then it is is applied for regular expression matching instead
of the global antipattern.

An .checks-out configuration file is required to have an empty scope
in the last policy in the policies array. This ensures that every
pull request has a matching scope.

If you have several approval policies, then we recommend enabling the [GitHub
notification channel](#github-comments) and include "open" in the types
field. This will enable checks-out to post a GitHub comment on a pull request
when the request is opened. The comment will describe which approval policy
is being applied to the pull request.

## Policy Scope

```json
scope:
{
  branches: []
  paths: []
}
```

If the 'branches' array is non-empty then the policy is limited
to the specified branches. In GitHub terminology this is a list of base
branches. The base branch is where the changes should be applied.

If the 'paths' array is non-empty then the policy is limited
to the specified file paths. Paths uses a simplified form of wildcard
glob expansion. The sequence `**` will match zero or more instances
of any character. The sequence `*` will match zero or more instances
of any character except the directory separator `/`. All other character
sequences are literal matches.

For example, to match against all java source files use the path
expression `**.java`. To match against recursively against all files
in a subdirectory use `foo/bar/**`.

## Policy Name

```json
name: ""
```

The 'name' field of a policy allows you to optionally specify a human-readable
description of the policy. If the name is empty then the policy is
identified by it's 1-based offset into the policy array. If the name is
nonempty then no two policies can have the same name.

## Policy Match

```json
match: "all[count=1,self=true]"
```

The policy match algorithms are specified with a domain-specific language (DSL).
The DSL is based around the concept of organization approvals. Boolean operators
can be used to combine organizations. Parenthesis may be used to defined
precedence of operations. There are several predefined organizations: "all",
"universe", "us", and "them".

### All Match

```json
match: "all[count=1,self=true]"
```

This matches against anyone in the MAINTAINERS file. The 'count'
field determines the minimum amount of approvals for a pull request
to be accepted. The 'self' field determines whether the author
of a pull request is allowed to approve the pull request.

### Universe Match

```json
match: "universe[count=1,self=true]"
```

This is a variation of "all" matching where anyone with an account
on this GitHub instance can approve the pull request. The 'count'
field determines the minimum amount of approvals for a pull request
to be accepted. The 'self' field determines whether the author
of a pull request is allowed to approve the pull request.

### Us Match

```json
match: "us[count=1,self=true]"
```

This policy will match to any person that has at least one organization
in common with the author of the pull request. The 'count'
field determines the minimum amount of approvals for a pull request
to be accepted. The 'self' field determines whether the author
of a pull request is allowed to approve the pull request.

### Them Match

```json
match: "them[count=1,self=true]"
```

This policy will match to any person that has at least no organizations
in common with the author of the pull request. The 'count'
field determines the minimum amount of approvals for a pull request
to be accepted. The 'self' field determines whether the author
of a pull request is allowed to approve the pull request.

### Org Match

```json
match: "foo[count=1,self=true]"
```

This policy will match against an organization named "foo" in the
MAINTAINERS file. The 'count' field determines the minimum
amount of organizations for a pull request to be accepted.
The 'self' field determines whether the author
of a pull request is allowed to approve the pull request.

### Anonymous Org Match

```json
match: "{john,jane,sally}[count=1,self=true]"
```

This policy will match against an an anonymous organization composed of the
specified people. Anonymous organizations are ignored by the 'us' and 'them'
groups defined above. The 'count' field determines the minimum amount of
organizations for a pull request to be accepted. The 'self' field determines
whether the author of a pull request is allowed to approve the pull request.

### And Match

```json
match: "foo[count=1,self=true] and bar[count=1,self=true]"
```

This policy allows you to combine two or more policies. All
of the policies must be met in order for this policy to be true.

### Or Match

```json
match: "foo[count=1,self=true] or bar[count=1,self=true]"
```

This policy allows you to combine two or more policies. One
or more of the policies must be met in order for this policy
to be true.

### Not Match

```json
match: "not foo[count=1,self=true]"
```

This policy allows you to negate another policy.

### True Match

```json
match: "true"
```

This policy allows all pull requests. It's very likely that you
want to use the "off" approval policy instead of this policy.

### False Match

```json
match: "false"
```

This policy allows no pull requests.

### Disable Match

```json
match: "off"
```

This policy disables the service. Added in version 0.6.10.

This is identical to the approval policy "true" with the addition that
it adds a local merge policy to disable automatic merging of branches.

### Atleast Function

```json
match: "atleast(2, foo[count=3,self=false], bar, baz, {jane,john})"
```

This function requires at least N of the specified groups to match
successfully, where N is the first parameter to the function.

### Author Function

```json
match: "author(foo or bar or not {jane,john})"
```

This function checks to see who is the author for the pull request. It is used to limit an approval policy to a set of authors. The potential author can be specified using an expression built out of the other matching expression terms. Since a pull request can only have a single author, any expression that can only be satisfied by more than one author, such as specifying `foo and bar`, or `foo[count=2]` will create a case that cannot match.

## Pattern

```json
pattern: "(?i)^checks-out\\s*(?P<version>\\S*)"
```

The regular expression that checks-out matches against the comment on a GitHub
pull request. The version capture group must be specified if
the version section of .checks-out is enable.

Pattern matching is performed using Go's regular expressions package. We
recommended testing custom regular expressions in the [Go
playground](http://play.golang.org/p/nQx_jGsLHz).

## Disapproval

```json
antipattern: null
antimatch: "all[count=1,self=true]"
```

checks-out has optional support for allowing the reviewer to raise a concern.
If an 'antipattern' is specified then the disapproval policy is enabled.
The disapproval policy is specified by the 'antimatch' parameter.
The antipattern was introduced in version 0.5.17 and the antimatch
was introduced in version 0.6.8.

Approvals and disapprovals are enforced by the following mechanism. First,
all the comments in the pull request are scanned for the 'antipattern' regular
expression. Any matching comments are applied towards the 'antimatch'
policy. If the antimatch policy becomes true then the pull request is blocked
from merge. If the antimatch policy is false then the match policy is tested
as usual. If the match policy becomes true then the pull request can be merged.

If a reviewer enters a disapproval comment followed by an approval comment,
then the approval comment will cancel out the disapproval comment. The
disapproval comment is forgotten but the approval comment also applies
towards the match policy, assuming the reviewer participates in both the
antimatch policy and the match policy.

The [commit range](#commit) configuration applies the same behavior to
approvals and disapprovals. With the default behavior, "head", only comments
that occur after the timestamp of the HEAD of the branch are used. Therefore
an objection will be erased after a new commit. Using the behavior "all" the
objection will persist and the reviewer must enter another comment with
the approval string.

## Author Restriction

```json
authormatch: "universe[count=1,self=true]"
```

It is possible to restrict the allowed authors with the authormatch approval
policy. The default value allows all users to submit pull requests. Using
the policy "all[count=1,self=true]" will only all users listed in the
MAINTAINERS file to submit a pull request. Since a pull request can only have a single author, any expression that can only be satisfied by more than one author, such as specifying `foo and bar`, or `foo[count=2]` will create a case that cannot match.

The difference between `authormatch` and the `author` function in the `match` is subtle. `author` is used to specify that a match rule applies an approval policy to a set of authors. `authormatch` is used to specify which authors can create pull requests at all.

## Work-In-Progress Pull Requests

```json
antititle: null
```

Optionally specify a regular expression that checks-out matches against the title
of a GitHub pull request. If the title matches then block the pull request
from being merged. A common value is "^WIP:". This feature was added in
version 0.7.6.

Pattern matching is performed using Go's regular expressions package. We
recommended testing custom regular expressions in the [Go
playground](http://play.golang.org/p/nQx_jGsLHz).

## Commit

```json
commit:
{
  range: head
  antirange: head
  tagrange: head
  ignoreuimerge: false
}
```

The range options affect the processing of commits on the pull request. The
range fields can have two possible values: "head" or "all". "head" will use the
comments that occur after the timestamp of the HEAD of the branch. "all" will
use all the comments on that branch. The 'range' parameter affects approval
comments, 'antirange' affects disapproval comments, and 'tagrange' affects
the tagging section. 'range' was introduced in version 0.5.17. 'antirange'
and 'tagrange' was introduced in version 0.7.7.

'ignoreuimerge' will ignore merges from upstream that are made through the
GitHub user interface (by clicking on "update branch").

## Maintainers

```json
maintainers:
{
  path: MAINTAINERS
  type: text
}
```

The path to the file that specifies the [project maintainers](../maintainers).
The type field can be "text", "hjson", "toml", or "legacy".

## Merge

```json
merge:
{
  enable: false
  delete: false
  method: "merge"
  uptodate: true
}
```

checks-out has the ability to automatically merge a pull request when the branch is
mergeable and all status checks (including optional ones) have passed. If
the delete flag is true then the head branch is deleted after a successful
merge. The delete flag was introduced in version 0.5.18. This only works if defined
inside a scoped block of approval policy, example provided:

```json
approvals:
[
  {
    name: "dev"
    scope:
    {
      branches: [ "develop", "qa", "release" ]
    }
    match: "all[count=1,self=true]"
    merge:
    {
      enable: true
      delete: true
      method: "merge"
      uptodate: true
    }
]
```

The final approval policy must also have no scope and be: `match: "off"`

This is to prevent random branch deletions and means that delete is explicitly opt-in.

If you enable automatic merges in the global merge section and you have
an approval policy that is the policy "true", then you are
encouraged to declare local merge section for that policy that is set to false.
Otherwise checks-out will automatically merge pull requests on those
branches as soon as they are opened. Instead of using the policy "true",
use the policy "off" which handles this case automatically.

The parameter 'method' determines the type of merge. It can be "merge",
"squash" or "rebase".

The parameter 'uptodate' determines whether the compare branch must
be zero commits behind the base branch before performing an automatic
merge. This parameter is enabled by default. It was introduced in
version 0.21.0.

## Tag

```json
tag:
{
  enable: false
  algorithm: semver
  template: "{{.Version}}"
  increment: "patch"
  docker: false
}
```

checks-out has the ability to automatically place a tag on a merge.
There are three tagging algorithms: "semver", "timestamp-rfc3339",
and "timestamp-millis". The 'template' field defines the
Golang [text/template](https://golang.org/pkg/text/template/)
that is used to generate the tag. The 'increment' field is used
by the semver algorithm and determines which section of the semantic
version to increment when no explicit version is provided
by the approver: 'major', 'minor', 'patch', or 'none'.
The 'docker' field enables stricter validation of the template to comply with
Docker tag requirements.

### Semantic Versioning

The semantic versioning standard is described at <http://semver.org/>.

Approval comments are used to provide the semantic version for the current pull request. By default, the version is specified after the I approve string:

```comment
I approve 1.0.0
```

will tag the merge commit with the version 1.0.0.

If multiple approvers for a pull commit specify a version, the highest version
specified will be used.

If no approver specifies a version, or if the maximum version specified by an
approver is less than any previously specified version tag, the version will be
set to the highest specified version tag with the patch version incremented.

### Timestamp Versioning

There are currently two options for timestamps, one based on the RFC 3339 standard and one specifying milliseconds since the epoch (Jan 1, 1970, 12:00:00AM UTC).

The type "timestamp-rfc3339" will look like this:

```comment
2016-05-16T19.06.26Z
```

(colons aren't legal in git tags, so the colons in the RFC format have been replaced by periods.)

The type "timestamp-millis" specifies the number of milliseconds since the epoch.

## Feedback

```json
feedback:
{
  type: ["comment", "review"]
}
```

The feedback section customizes the processing of approval events. The feedback
type determines which kinds of events are processed. "comment" accepts
pull request comments. "review" accepts pull request reviews.

## Comment

```json
comment:
{
  enable: false
  targets: []  
}
```

Allows notifications to be sent to output channels when checks-out performs an
action. Legal values for the "types" field are the following:

* "error" Error while processing webhook
* "open" Pull request is opened
* "close" Pull request is closed without being merged
* "accept" Pull request has been closed and merged
* "approve" An approval comment has been added
* "block" A disapproval comment has been added
* "reset" Commit of PR branch has reset previous approvals
* "merge" Pull request was auto-merged after all status checks passed
* "tag" Branch was tagged after merge
* "delete" Branch was auto-deleted after merge
* "deploy" Deployment was triggered after merge
* "author" Pull request is blocked because author is not approved

### GitHub Comments

```json
{
  target: github
  pattern: null
  types: []
}
```

If enabled then checks-out will add comments to the pull request. Pattern is an
optional regular expression that allows comments to be only applied to
some pull requests based on the title of the request. The types field
is described above. If types is empty then all types send notifications.
GitHub comments was introduced in 0.7.9.

### Slack Comments

```json
{
  target: foobar.slack.com
  pattern: null
  types: []
  names: []
}
```

If enabled then checks-out will write messages to Slack. Pattern is an
optional regular expression that allows comments to be only applied to
some pull requests based on the title of the request. The types field
is described above. If types is empty then all types send notifications.
Names is a list of slack channels or users you want to message. Channels are prefixed with #,
users are prefixed with @. Slack comment support was introduced in 0.7.9.

The Slack integration allows custom servers to be specified.

The target field can be the hostname of the Slack target such as "foobar.slack.com".
If that hostname has been previously registered with the checks-out service then you
will be able to use it. Otherwise you need to register a [Slack Webhook](https://api.slack.com/incoming-webhooks) with the checks-out service. Use the
/api/user/slack/:hostname endpoint to register the webhook. This is documented
on the API page.

## Deploy

```json
deploy:
{
  enable: false
  deployment: DEPLOYMENTS
}
```

TODO: add documentation
