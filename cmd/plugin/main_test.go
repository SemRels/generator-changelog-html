// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The generator-changelog-html Authors

package main

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRunWritesHTML(t *testing.T) {
	t.Parallel()

	getenv := func(key string) string {
		switch key {
		case "SEMREL_VERSION":
			return "1.3.0"
		case "SEMREL_CURRENT_VERSION":
			return "1.2.0"
		case "SEMREL_BRANCH":
			return "main"
		case "SEMREL_COMMITS":
			return `["feat: add search"]`
		}
		return ""
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := run(&stdout, &stderr, getenv)

	require.Equal(t, 0, code)
	require.Empty(t, stderr.String())
	require.Contains(t, stdout.String(), "<section class=\"changelog-entry\">")
	require.Contains(t, stdout.String(), "<h3>Features</h3>")
}

func TestRunRejectsInvalidCommitJSON(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := run(&stdout, &stderr, func(key string) string {
		if key == "SEMREL_COMMITS" {
			return `[`
		}
		return ""
	})

	require.Equal(t, 1, code)
	require.Empty(t, stdout.String())
	require.Contains(t, stderr.String(), "invalid SEMREL_COMMITS JSON")
}

func TestReleaseContextFromEnvUsesVersionFallback(t *testing.T) {
	t.Parallel()

	ctx, err := releaseContextFromEnv(func(key string) string {
		switch key {
		case "SEMREL_NEXT_VERSION":
			return "2.0.0"
		case "SEMREL_CURRENT_VERSION":
			return "1.9.0"
		case "SEMREL_BRANCH":
			return "release/main"
		}
		return ""
	})

	require.NoError(t, err)
	require.Equal(t, "2.0.0", ctx.Version)
	require.Equal(t, "1.9.0", ctx.CurrentVersion)
	require.Equal(t, "release/main", ctx.Branch)
}
