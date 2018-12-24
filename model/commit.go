package model

import "github.com/capitalone/checks-out/strings/lowercase"

type Commit struct {
	Author    lowercase.String
	Committer string
	Message   string
	SHA       string
	Parents   []string
}

func DefaultCommit() CommitConfig {
	return CommitConfig{
		Range:         Head,
		AntiRange:     Head,
		TagRange:      Head,
		IgnoreUIMerge: true,
	}
}
