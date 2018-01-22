package model

import "github.com/capitalone/checks-out/strings/lowercase"

type Commit struct {
	Author  lowercase.String
	Message string
	SHA     string
}
