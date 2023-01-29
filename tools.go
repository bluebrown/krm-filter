//go:build tools

package krmfilter

import (
	_ "github.com/GoogleContainerTools/kpt/mdtogo"
	_ "github.com/golangci/golangci-lint/cmd/golangci-lint"
	_ "github.com/google/ko"
	_ "github.com/google/yamlfmt/cmd/yamlfmt"
)
