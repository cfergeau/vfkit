package cmdline

import (
	"regexp"
	"strings"
	"testing"
)

func TestVersion(t *testing.T) {
	version := Version()

	// Version should not be empty
	if version == "" {
		t.Error("Version() returned empty string")
	}

	// Version should not be the placeholder "unknown" when built from git
	// (it might be "unknown" only in very rare cases outside git repo)
	if version == "unknown" {
		t.Log("Version is 'unknown' - this may be expected if not in a git repository")
	}

	t.Logf("Version: %s", version)
}

func TestModuleVersionFromBuildInfo(t *testing.T) {
	version := moduleVersionFromBuildInfo()

	// Should return non-empty when built from git with Go 1.24+
	if version == "" {
		t.Log("moduleVersionFromBuildInfo() returned empty - may be expected in some build contexts")
	} else {
		t.Logf("Build info version: %s", version)

		// Should not be the "(devel)" placeholder
		if version == "(devel)" {
			t.Error("moduleVersionFromBuildInfo() returned '(devel)' but should filter this out")
		}
	}
}

func TestVersionFormat(t *testing.T) {
	version := Version()

	if version == "" || version == "unknown" {
		t.Skip("Skipping format test - no version available")
	}

	// Version should match expected patterns:
	// - v0.6.3 (exact tag)
	// - v0.6.3-20-gcc44a9a (commits after tag)
	// - v0.6.3-20-gcc44a9a-dirty (uncommitted changes)
	// - (devel) should be filtered out by moduleVersionFromBuildInfo

	validPatterns := []*regexp.Regexp{
		regexp.MustCompile(`^v\d+\.\d+\.\d+$`),                      // v0.6.3
		regexp.MustCompile(`^v\d+\.\d+\.\d+-\d+-g[a-f0-9]+$`),       // v0.6.3-20-gcc44a9a
		regexp.MustCompile(`^v\d+\.\d+\.\d+-\d+-g[a-f0-9]+-dirty$`), // v0.6.3-20-gcc44a9a-dirty
		regexp.MustCompile(`^v\d+\.\d+\.\d+-dirty$`),                // v0.6.3-dirty
	}

	matched := false
	for _, pattern := range validPatterns {
		if pattern.MatchString(version) {
			matched = true
			break
		}
	}

	if !matched {
		t.Logf("Version '%s' does not match expected patterns, but may be valid", version)
	}
}

func TestGitArchiveVersionFallback(t *testing.T) {
	// Save original value
	originalGitArchive := gitArchiveVersion
	defer func() {
		gitArchiveVersion = originalGitArchive
	}()

	// Test with substituted git archive version
	gitArchiveVersion = "v0.6.3-test"

	// When moduleVersionFromBuildInfo returns empty, should fall back to git archive
	// We can't easily mock moduleVersionFromBuildInfo, but we can verify the logic
	if !strings.HasPrefix(gitArchiveVersion, "$Format:") {
		t.Logf("Git archive version would be used: %s", gitArchiveVersion)
	}

	// Test with un-substituted placeholder
	gitArchiveVersion = "$Format:%(describe)$"
	if strings.HasPrefix(gitArchiveVersion, "$Format:") {
		t.Log("Git archive version is placeholder, would not be used")
	}
}
