+++
date = "2015-12-05T16:00:21-08:00"
draft = false
title = "Quick Start"
weight = 1
menu = "main"
toc = true
+++

# Overview

Checks-out is a fork of [LGTM](https://github.com/lgtmco/lgtm).

Checks-out will register itself as a required status check. When you enable your repository on public GitHub it is automatically configured to use [protected branches](https://github.com/blog/2051-protected-branches-and-required-status-checks). Protected branches prevent pull requests from being merged until required status checks are passing. If you are using Enterprise GitHub then you might be required to manually enable protected branches on the master branch.

You can customize many options using a .checks-out file in your repository. Please see the [customization documentation](../customize) for more detail.

# Setup

The simplest way to configure checks-out is to create a .checks-out file the
default branch of your GitHub repository with the contents:

```
approvals:
[
  {
    # count determines how many approvals are needed
    # self determines whether you can approve your own pull requests
    match: "all[count=1,self=true]"
  }
]
```

Create a MAINTAINERS file in the default branch with the contents:

```
github-org repo-self
```

Login to the checks-out service (this will provide the bot
with credentials to monitor your repository) and enable the service
for your repo. Click on the button that says "OFF" to enable the service.

## Organization Setup

Create a repository named "checks-out-configuration" in your organization.
The files template.checks-out and template.MAINTAINERS in the
checks-out-configuration repository will be used for any repository
that is missing .checks-out or MAINTAINERS files. You can also register
an entire GitHub organization using the checks-out user interface. Select
the "on/off" toggle next to the name of an organization. Registering
an organization will register any existing repositories and will auto-register
any new repositories. Added in version 0.12.0.

# Approvers

Define a list of individuals that can approve pull requests. This list should be store in a MAINTAINERS file to the root of your repository. Please see the [maintainers](../maintainers) documentation for supported file formats.

**Please note** that checks-out pulls the MAINTAINERS file from your repository default branch (typically master). Changes to the MAINTAINERS file are not recognized until present in the default branch. Changes to the MAINTAINERS file in a pull request are therefore not recognized until merged.

# Approvals

Pull requests are locked and cannot be merged until the minimum number of approvals are received. Project maintainers can indicate their approval by commenting on the pull request and including "I approve" in their approval text.

The service also provides integration with GitHub Reviews. An accepted GitHub
Review is counted as an approval. GitHub Review that requests additional
changes blocks the pull request from merging.
