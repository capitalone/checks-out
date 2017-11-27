+++
date = "2015-12-05T16:00:21-08:00"
draft = false
title = "FAQ"
weight = 8
menu = "main"
toc = true
+++

# Why don't I see my repository?

Here are some tips for troubleshooting:

* Is your GitHub repository public?
* Do you have Admin access to the repository?
* Did you grant access to your GitHub organization?
* Did you recently create the repository? There may be up to a 15 minute delay due to internal caching.
* Does your organization restrict integrations? [Learn more](https://github.com/blog/1941-organization-approved-applications).

# Can I use with GitHub Enterprise?

Yes, In order to use with GitHub Enterprise you will need to install and run your own instance. See the [the documentation](../install).

GitHub Enterprise 2.4 has partial support for protected branches and requires [manual configuration](../branches/#github-enterprise:84e35ebc125ab31a6f85cb9b5e08d8c9). GitHub Enterprise 2.3 and below are not supported.

# Why are all the configuration files in HJSON format?

When we began to write complex approval policies in the .checks-out
configuration file, it became apparent that TOML was too verbose for
the nested configuration sections. We looked for a golang text format
serde library that supported the following features: actually parses
whatever format without crashing on corner cases, can defer
unmarshal of specific fields (json.rawMessage), and can write custom
marshal and unmarshal functions.

The encoding/json library is the only library that met these requirements.
Human JSON is a small amount of syntactic sugar for JSON and adding
support for HJSON was trivial.

# Can I install on my own server?

Yes. See the [install documentation](../install).

# Can I use a central Maintainers file?

Sort of. You can create a maintainers team in your GitHub organization. See the [maintainers documentation](../maintainers).

# Why aren't approvals being processed?!

Verify that you have receive the correct number of approvals. The default configuration requires a minimum of two approvals. See the [documentation](../customize) to learn more about custom configurations.

Verify that hooks are being sent correctly. You can see an audit log of all hooks in the **Webhooks & Services** section or your repository settings.

Verify the message is being processed successful. An unsuccessful message will be flagged accordingly in GitHub. Error messages from the service are written to the response body.

![hook error](/images/hook_error.png)

Verify the response from a successful hook. The approval settings, approval status, and list of individuals that approved the pull request are included in the payload for auditing purposes.

![hook success](/images/hook_success.png)

If the payload indicates the pull request was approved, and this is not reflected in the status you can click the re-deliver button to re-deliver the payload.

# Why can't I merge my pull request?

Please remember that checks-out uses GitHub [protected branches](https://github.com/blog/2051-protected-branches-and-required-status-checks) which prevents merging a branch that has failed to meet all required status checks. The ability to merge a pull request is completely governed by GitHub.

Please also note that when protected branches prevent you from merging a pull request that has fallen behind the target branch. That is to say, if a pull request is merged, all outstanding pull requests will need to re-sync before they can be merged.

# This thing is busted!

Please be sure you have read the documentation, including this FAQ, and fully understand how GitHub [protected branches](https://github.com/blog/2051-protected-branches-and-required-status-checks) work. If you are still experiencing issues please [contact support](../support) and we'll be happy to help!
