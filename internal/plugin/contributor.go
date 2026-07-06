// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The generator-changelog-html Authors

package plugin

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

type Contributor struct {
	Name                   string
	Login                  string
	ProfileURL             string
	FirstContributionLabel string
	FirstContributionURL   string
	FirstContributionPR    int
	FirstContributionSHA   string
	CommitCount            int
	FirstTime              bool
}

func (c *Contributor) UnmarshalJSON(data []byte) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	c.Name = firstJSONString(raw,
		"name",
		"displayName",
		"author",
		"authorName",
	)
	c.Login = normalizeLogin(firstJSONString(raw,
		"login",
		"username",
		"handle",
		"github",
		"githubLogin",
		"github_login",
		"mention",
	))
	c.ProfileURL = firstJSONString(raw,
		"profileURL",
		"profileUrl",
		"profile",
		"htmlURL",
		"htmlUrl",
		"url",
	)
	c.FirstContributionLabel = firstJSONString(raw,
		"firstContributionLabel",
		"firstContributionRef",
		"firstContributionReference",
		"reference",
		"ref",
	)
	c.FirstContributionURL = firstJSONString(raw,
		"firstContributionURL",
		"firstContributionUrl",
		"firstContributionLink",
	)
	c.CommitCount = firstJSONInt(raw, "commitCount", "commits")
	c.FirstTime = firstJSONBool(raw, "firstTime", "isFirstTimeContributor", "newContributor")

	if pr := firstJSONInt(raw, "pr", "pullRequest", "pull_request", "firstPR", "firstPr"); pr > 0 {
		c.FirstContributionPR = pr
	}
	if sha := firstJSONString(raw, "firstCommit", "commit", "sha"); sha != "" {
		c.FirstContributionSHA = sha
	}

	if firstContributionRaw, ok := raw["firstContribution"]; ok && len(firstContributionRaw) > 0 {
		var firstContribution map[string]json.RawMessage
		if err := json.Unmarshal(firstContributionRaw, &firstContribution); err == nil {
			c.FirstTime = true
			if label := firstJSONString(firstContribution, "label", "reference", "ref", "title"); label != "" {
				c.FirstContributionLabel = label
			}
			if url := firstJSONString(firstContribution, "url", "htmlURL", "htmlUrl", "link"); url != "" {
				c.FirstContributionURL = url
			}
			if pr := firstJSONInt(firstContribution, "number", "pr", "pullRequest", "pull_request"); pr > 0 {
				c.FirstContributionPR = pr
			}
			if sha := firstJSONString(firstContribution, "commit", "sha"); sha != "" {
				c.FirstContributionSHA = sha
			}
		}
	}

	if c.Name == "" && c.Login != "" {
		c.Name = strings.TrimPrefix(c.Login, "@")
	}
	if c.FirstContributionLabel == "" {
		switch {
		case c.FirstContributionPR > 0:
			c.FirstContributionLabel = fmt.Sprintf("#%d", c.FirstContributionPR)
		case c.FirstContributionSHA != "":
			c.FirstContributionLabel = shortReference(c.FirstContributionSHA)
		}
	}
	if !c.FirstTime {
		c.FirstTime = c.FirstContributionLabel != "" || c.FirstContributionURL != "" || c.FirstContributionPR > 0 || c.FirstContributionSHA != ""
	}

	return nil
}

func normalizeLogin(value string) string {
	return strings.TrimSpace(strings.TrimPrefix(value, "@"))
}

func firstJSONString(raw map[string]json.RawMessage, keys ...string) string {
	for _, key := range keys {
		value, ok := raw[key]
		if !ok || len(value) == 0 {
			continue
		}

		var str string
		if err := json.Unmarshal(value, &str); err == nil {
			str = strings.TrimSpace(str)
			if str != "" {
				return str
			}
		}
	}

	return ""
}

func firstJSONBool(raw map[string]json.RawMessage, keys ...string) bool {
	for _, key := range keys {
		value, ok := raw[key]
		if !ok || len(value) == 0 {
			continue
		}

		var boolean bool
		if err := json.Unmarshal(value, &boolean); err == nil {
			return boolean
		}

		var str string
		if err := json.Unmarshal(value, &str); err == nil {
			parsed, err := strconv.ParseBool(strings.TrimSpace(str))
			if err == nil {
				return parsed
			}
		}
	}

	return false
}

func firstJSONInt(raw map[string]json.RawMessage, keys ...string) int {
	for _, key := range keys {
		value, ok := raw[key]
		if !ok || len(value) == 0 {
			continue
		}

		if count, ok := jsonInt(value); ok {
			return count
		}
	}

	return 0
}

func jsonInt(value json.RawMessage) (int, bool) {
	var integer int
	if err := json.Unmarshal(value, &integer); err == nil {
		return integer, true
	}

	var str string
	if err := json.Unmarshal(value, &str); err == nil {
		parsed, err := strconv.Atoi(strings.TrimSpace(str))
		if err == nil {
			return parsed, true
		}
	}

	var array []json.RawMessage
	if err := json.Unmarshal(value, &array); err == nil {
		return len(array), true
	}

	return 0, false
}

func shortReference(reference string) string {
	reference = strings.TrimSpace(reference)
	if len(reference) > 7 {
		return reference[:7]
	}
	return reference
}
