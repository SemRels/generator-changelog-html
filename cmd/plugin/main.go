// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The generator-changelog-html Authors

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	plugin "github.com/SemRels/generator-changelog-html/internal/plugin"
)

const pluginSchemaVersion = 1

func main() {
	os.Exit(run(os.Stdout, os.Stderr, os.Getenv))
}

func run(stdout, stderr io.Writer, getenv func(string) string) int {
	_, _ = fmt.Fprintf(stderr, "plugin_schema_version=%d\n", pluginSchemaVersion)
	ctx, err := releaseContextFromEnv(getenv)
	if err != nil {
		fmt.Fprintln(stderr, "generator-changelog-html:", err)
		return 1
	}

	options := plugin.DefaultGenerateOptions()
	options.Signature = envBool(getenv, "SEMREL_PLUGIN_SIGNATURE", false)
	options.NewContributors = envBoolSynonyms(getenv, true, "SEMREL_PLUGIN_FIRST_TIME_CONTRIBUTORS", "SEMREL_PLUGIN_NEW_CONTRIBUTORS")
	options.MVP = envBoolSynonyms(getenv, false, "SEMREL_PLUGIN_RELEASE_MVP", "SEMREL_PLUGIN_MVP")
	options.AIDisclosure = envBool(getenv, "SEMREL_PLUGIN_AI_DISCLOSURE", false)
	options.AIDisclosureSection = envBool(getenv, "SEMREL_PLUGIN_AI_DISCLOSURE_SECTION", false)
	if badge := strings.TrimSpace(getenv("SEMREL_PLUGIN_AI_DISCLOSURE_BADGE")); badge != "" {
		options.AIDisclosureBadge = badge
	}
	if metric := strings.TrimSpace(getenv("SEMREL_PLUGIN_MVP_METRIC")); metric != "" {
		options.MVPMetric = metric
	}
	for _, warning := range contributorWarnings(getenv) {
		fmt.Fprintln(stderr, "generator-changelog-html:", warning)
	}
	options.Contributors = contributorsFromEnv(getenv)

	if _, err := io.WriteString(stdout, plugin.New().Generate(ctx, options)); err != nil {
		fmt.Fprintln(stderr, "generator-changelog-html:", err)
		return 1
	}

	return 0
}

func releaseContextFromEnv(getenv func(string) string) (plugin.ReleaseContext, error) {
	raw := getenv("SEMREL_COMMITS")

	var commits []string
	if raw != "" {
		if err := json.Unmarshal([]byte(raw), &commits); err != nil {
			return plugin.ReleaseContext{}, fmt.Errorf("invalid SEMREL_COMMITS JSON: %w", err)
		}
	}

	return plugin.ReleaseContext{
		Version:        firstNonEmpty(getenv("SEMREL_VERSION"), getenv("SEMREL_TAG_NAME"), getenv("SEMREL_NEXT_VERSION")),
		CurrentVersion: strings.TrimSpace(getenv("SEMREL_CURRENT_VERSION")),
		Branch:         strings.TrimSpace(getenv("SEMREL_BRANCH")),
		RepositoryURL:  strings.TrimSpace(getenv("SEMREL_REPOSITORY_URL")),
		Commits:        commits,
	}, nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}
	return ""
}

func envBool(getenv func(string) string, key string, defaultValue bool) bool {
	value := strings.TrimSpace(getenv(key))
	if value == "" {
		return defaultValue
	}

	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return defaultValue
	}

	return parsed
}

func envBoolSynonyms(getenv func(string) string, defaultValue bool, keys ...string) bool {
	hasExplicit := false
	value := defaultValue
	for _, key := range keys {
		raw := strings.TrimSpace(getenv(key))
		if raw == "" {
			continue
		}
		parsed, err := strconv.ParseBool(raw)
		if err != nil {
			continue
		}
		if parsed {
			return true
		}
		hasExplicit = true
		value = false
	}
	if hasExplicit {
		return value
	}
	return defaultValue
}

func contributorsFromEnv(getenv func(string) string) []plugin.Contributor {
	for _, key := range []string{"SEMREL_CONTRIBUTORS", "SEMREL_PLUGIN_CONTRIBUTORS_JSON"} {
		raw := strings.TrimSpace(getenv(key))
		if raw == "" {
			continue
		}

		var contributors []plugin.Contributor
		if err := json.Unmarshal([]byte(raw), &contributors); err != nil {
			continue
		}
		if key == "SEMREL_PLUGIN_CONTRIBUTORS_JSON" {
			for index := range contributors {
				contributors[index].FirstTime = true
			}
		}
		return contributors
	}
	return nil
}

func contributorWarnings(getenv func(string) string) []string {
	warnings := make([]string, 0, 2)
	for _, key := range []string{"SEMREL_CONTRIBUTORS", "SEMREL_PLUGIN_CONTRIBUTORS_JSON"} {
		raw := strings.TrimSpace(getenv(key))
		if raw == "" {
			continue
		}

		var contributors []plugin.Contributor
		if err := json.Unmarshal([]byte(raw), &contributors); err == nil {
			return warnings
		}
		warnings = append(warnings, fmt.Sprintf("invalid %s JSON: ignored", key))
	}
	return warnings
}
