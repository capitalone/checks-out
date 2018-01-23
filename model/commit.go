package model

import "github.com/capitalone/checks-out/strings/lowercase"

type Commit struct {
	Author    lowercase.String
	Committer string
	Message   string
	SHA       string
}
