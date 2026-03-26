package cmdline

import (
	"runtime/debug"
	"strings"
)

var (
	// set through .gitattributes when `git archive` is used
	// see https://icinga.com/blog/2022/05/25/embedding-git-commit-information-in-go-binaries/
	gitArchiveVersion = "$Format:%(describe)$"
)

func Version() string {
	// Go 1.24+ automatically embeds VCS version in the binary
	if version := moduleVersionFromBuildInfo(); version != "" {
		return version
	}

	// Fallback to git archive version for GitHub release tarballs
	if !strings.HasPrefix(gitArchiveVersion, "$Format:") {
		return gitArchiveVersion
	}

	return "unknown"
}

func moduleVersionFromBuildInfo() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return ""
	}
	if info.Main.Version == "(devel)" {
		return ""
	}
	return info.Main.Version
}
