# Checks-Out

Checks-Out is a simple pull request approval system using GitHub
protected branches and maintainers files. Pull requests are locked and cannot be
merged until the minimum number of approvals are received. Project maintainers
can indicate their approval by commenting on the pull request and including
"I approve" in their approval text. Checks-Out also provides integration
with GitHub Reviews. An accepted GitHub Review is counted as an approval.
GitHub Review that requests additional changes blocks the pull request from merging.

Read the [online documentation](https://capitalone.github.com/checks-out/docs) to find out more about Checks-Out.

### Development

Checks-Out is a fork of [LGTM](https://github.com/lgtmco/lgtm). Our git repository
contains the commit history from the upstream project. We are actively seeking
contributions from the community. If you'd like to contribute we recommend
taking a look at the issues page. You can pick up an open issue and work on it,
submit a bug, or submit a new feature request for feedback.

### Features

Checks-Out has several features that distinguish itself from the parent LGTM project.

The most popular feature is the ability to specify multiple approval policies.
Policies are based around the concept of organizations. An organization is
a set of project maintainers. Various types of thresholds can be configured
for organizations and boolean conditions can be used to combine policies.
Policies can be configured to apply to specific file paths and/or git branches.
Refer to the customization documentation for more information about policies.

Checks-Out has optional support for automatic tagging of merges. Tags can configured
based on timestamp or semantic versioning.

Checks-Out has optional support for automatic merging of pull requests when
all status checks have passed.

Checks-Out has changed the default behavior when new commits are added to a pull
request. By default only comments that have a later timestamp than the
latest commit are processed by Checks-Out. There is a configuration property to use
the original LGTM behavior which is to consider all comments on a pull request. 

### Usage

#### .checks-out file

Each repository managed by Checks-Out must have a .checks-out file in the root of the
repository. This file provides the configuration that Checks-Out uses for the
repository. The configuration file is described in detail in the
customization section of the online documentation.

This repository has an .checks-out file that you can use as an example.
It is likely that you will need a simple .checks-out file, so you can use
the following template:

```
approvals:
[
  {
    match: "all[count=1,self=false]"
  }
]
```

#### MAINTAINERS file

Each repository managed by Checks-Out should have a MAINTAINERS file that specifies
who is allowed to approve pull requests. The format of the file
is described in the maintainers section of the online
documentation. Here is a sample MAINTAINERS file to get you started:

```
github-org repo-self
```

### Build

Checks-Out uses the Go [dep](https://github.com/golang/dep) dependency management tool.
Dependencies are not stored in the repository. Run `dep ensure` to install dependencies.

Commands to build from source:

```sh
make build   # Build the binary
```

## Contributors

We welcome your interest in Capital One’s Open Source Projects (the “Project”). Any Contributor to the Project must accept and sign a CLA indicating agreement to the license terms. Except for the license granted in this CLA to Capital One and to recipients of software distributed by Capital One, You reserve all right, title, and interest in and to your Contributions; this CLA does not impact your rights to use your own contributions for any other purpose.

[Link to Individual CLA](https://docs.google.com/forms/d/19LpBBjykHPox18vrZvBbZUcK6gQTj7qv1O5hCduAZFU/viewform)

[Link to Corporate CLA](https://docs.google.com/forms/d/e/1FAIpQLSeAbobIPLCVZD_ccgtMWBDAcN68oqbAJBQyDTSAQ1AkYuCp_g/viewform)

This project adheres to the Capital One [Open Source Code of Conduct](http://www.capitalone.io/codeofconduct/). By participating, you are expected to honor this code.

### Contribution Guidelines
We encourage any contributions that align with the intent of this project and add more functionality or languages that other developers can make use of. To contribute to the project, please submit a PR for our review. Before contributing any source code, familiarize yourself with the [Apache License 2.0](LICENSE), which controls the licensing for this project.

## License

Checks-Out is available under the Apache License 2.0.

This distribution has a binary dependency on errwrap, which is available under
the Mozilla Public License 2.0 License. The source code of errwrap can be found at
https://github.com/hashicorp/errwrap.

This distribution has a binary dependency on go-version, which is available under
the Mozilla Public License 2.0 License. The source code of go-version can be found at
https://github.com/hashicorp/go-version.

This distribution has a binary dependency on go-multierror, which is available under
the Mozilla Public License 2.0 License. The source code of go-multierror can be found
at https://github.com/mspiegel/go-multierror.

This distribution has a binary dependency on go-sql-driver/mysql, which is available under
the Mozilla Public License 2.0 License. The source code of go-sql-driver/mysql can be found
at https://github.com/go-sql-driver/mysql

## FAQ

1\. How is this different from GitHub Reviews?

Please use [GitHub Reviews](https://help.github.com/articles/about-pull-request-reviews/) if it meets all your requirements. Some significant features in Checks-Out that are not (yet) in GitHub Reviews are: custom
approval policies, different approval policies for different branches and/or file paths, optional auto-merge
when all status checks have passed, optional auto-tagging of merges.
