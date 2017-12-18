+++
date = "2015-12-05T16:00:21-08:00"
draft = false
title = "Branches"
weight = 3
menu = "main"
toc = true
+++

# Overview

Checks-out automatically enables GitHub protected branches for your repositories' __default branch__. You can further customize or disable this behavior by navigating to your repository branch settings screen in GitHub.

![protected branches](/checks-out/images/protected_branches.png)

# GitHub Enterprise

GitHub Enterprise 2.4 does not have an API for protected branches. As a result you will need to manually configure protected branches after enabling your repository in checks-out. This can be done by navigating to your repository branch settings screen in GitHub Enterprise (pictured above).

# Other Branches

Checks-out automatically enables itself for the __default__ branch. If you would like to enable checks-out for additional branches you can manually configure additional branches in your repository branch settings screen in GitHub.

# Other Checks

Checks-out automatically adds itself as a required status check. If you would like to enable additional status checks, such as continuous integration or code coverage, you can further customize this behavior in your repository branch settings screen in GitHub.
