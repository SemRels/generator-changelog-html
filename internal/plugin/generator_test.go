// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The generator-changelog-html Authors

package plugin

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestGeneratorGenerate(t *testing.T) {
	t.Parallel()

	generator := &Generator{now: func() time.Time {
		return time.Date(2026, 5, 25, 0, 0, 0, 0, time.UTC)
	}}

	output := generator.Generate(ReleaseContext{
		Version:        "1.3.0",
		CurrentVersion: "1.2.0",
		Branch:         "main",
		Commits: []string{
			"feat: add <search>",
			"fix: escape & encode",
			"refactor!: remove deprecated endpoint",
		},
	})

	require.Equal(t, "<section class=\"changelog-entry\">\n  <h2>v1.3.0 <small>2026-05-25</small></h2>\n  <p>Previous release: v1.2.0 · Branch: main</p>\n  <h3>Breaking Changes</h3>\n  <ul>\n    <li>refactor!: remove deprecated endpoint</li>\n  </ul>\n  <h3>Features</h3>\n  <ul>\n    <li>feat: add &lt;search&gt;</li>\n  </ul>\n  <h3>Bug Fixes</h3>\n  <ul>\n    <li>fix: escape &amp; encode</li>\n  </ul>\n</section>", output)
}

func TestMetadataLine(t *testing.T) {
	t.Parallel()

	require.Equal(t, "", metadataLine(ReleaseContext{}))
	require.Equal(t, "Previous release: v1.2.0", metadataLine(ReleaseContext{CurrentVersion: "1.2.0"}))
	require.Equal(t, "Branch: main", metadataLine(ReleaseContext{Branch: "main"}))
}

func TestClassifyCommit(t *testing.T) {
	t.Parallel()

	section, line := classifyCommit("feat: add endpoint\n\nBREAKING CHANGE: API changed")
	require.Equal(t, breakingChangesSection, section)
	require.Equal(t, "BREAKING CHANGE: API changed", line)

	section, line = classifyCommit("docs: update README")
	require.Equal(t, otherChangesSection, section)
	require.Equal(t, "docs: update README", line)
}
