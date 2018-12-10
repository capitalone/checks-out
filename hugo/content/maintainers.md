+++
date = "2015-12-05T16:00:21-08:00"
draft = false
title = "Maintainers"
weight = 2
menu = "main"
+++

The list of approvers is stored in a `MAINTAINERS` text file in the root
of your repository. The maintainers file can be in either a text or HJSON
format. The format is specified in the [.checks-out file](../customize) file
and the default format is text.

You can inspect the maintainers file for your repository using the api endpoint
`/api/repos/[orgname]/[reponame]/maintainers`. The output will have any
github-org and github-team macros expanded (see below). If the maintainers
file cannot be parsed then an error message will be returned.

# Text format

The text format can be any of the following types.

Username, separated by newline:

```json
foo
bar
baz
```

Username and email address, separated by newline:

```json
foo <foo@mail.com>
bar <bar@mail.com>
baz <baz@mail.com>
```

FullName, email address and username, separated by newline:

```json
Fooshure Jones <foo@mail.com> (@foo)
Bar None <bar@mail.com> (@bar)
Bazinga Smith <baz@mail.com> (@baz)
```

Directives for importing GitHub organizations and GitHub teams.
These directives populate both the 'people' and 'org' fields
with the correct values (see below).

```json
github-org foo # loads organization foo
github-team bar # loads team bar within the organization of the repository
github-team bar foo # loads team bar from organization foo
github-org repo-self # loads the organization of the repository
github-team repo-self # loads all the teams of the repository
```

In the examples above the groups are automatically assigned the following names:

```json
foo
bar
foo-bar
[name of organization]
[names of teams]
```

# HJSON format

[Human JSON](http://hjson.org) format inspired by the [Docker
project](https://github.com/docker/opensource/blob/master/MAINTAINERS).
Organizations can be specified using this format.

The github-org and github-team directives can be used in the 'group' section
but not in the 'people' section. These directives will populate the correct
values into the 'people' section.

```json
{
  people:
  {
    bob:
    {
      name: Bob Bobson
      email: bob@email.co
    }
    fred:
    {
      name: Fred Fredson
      email: fred@email.co
    }
    jon:
    {
      name: Jon Jonson
      email: jon@email.co
    }
    ralph:
    {
      name: Ralph Ralphington
      email: ralph@email.co
    }
    george:
    {
      name: George McGeorge
      email: george@email.co
    }
  }
  org:
  {
    cap:
    {
      people: [ "bob", "fred", "jon", "github-org cap" ]
    }
    iron:
    {
      people: [ "ralph", "george", "github-team iron" ]
    }
  }
}
```
