+++
date = "2015-12-05T16:00:21-08:00"
draft = false
title = "Approval Policies"
weight = 5
menu = "main"
toc = true
+++

Here are some templates you can use to write the approval policy section
of your .checks-out configuration.

```
match: "all[count=1,self=false]"
```

This policy requires one approval from anyone in the built-in 'all'
group. 'all' expands to all the people in the MAINTAINERS file.
Populating the MAINTAINERS file with the macro "github-org repo-self"
will expand to everyone within the organization of the GitHub repository.
self=false forbids the author of the pull request to approve their own
request. If you want self-approval then use "all[count=1,self=true]"

```
match: "sharks[count=1,self=false] and jets[count=1,self=false]"
```

You have defined the groups 'sharks' and 'jets' in the MAINTAINERS file.
This policy requires one approval from the sharks and one approval
from the jets. If someone belongs to both the sharks and the jets, then
their approval will count for both of the clauses and therefore they
do not require a second reviewer. In general, the teams work best when
they are non-overlapping. But the next example takes advantage of
overlapping groups.

```
match: "all[count=2,self=false] and reviewers[count=1,self=false]"
```

This policy requires two approvals and at least one of the approvals
must be from the group 'reviewers' in the MAINTAINERS file. The policy
takes advantage of the fact that reviewers is a subset of all. If
you only have 1 person in the reviewers group and you want the reviewer
to submit pull requests and it is OK to have only a single approval
against those pull request, then use 
"all[count=2,self=false] and reviewers[count=1,self=true]" and the
reviewer must self-approve their pull request.