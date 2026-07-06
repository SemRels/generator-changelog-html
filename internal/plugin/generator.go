// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The generator-changelog-html Authors

package plugin

import (
	"fmt"
	"html"
	"regexp"
	"strings"
	"time"
)

// aiTrailerPatterns matches known AI co-author trailers in commit messages.
var aiTrailerPatterns = []struct {
	pattern *regexp.Regexp
	label   string
}{
	{regexp.MustCompile(`(?i)co-authored-by:[^\n]*copilot`), "GitHub Copilot"},
	{regexp.MustCompile(`(?i)co-authored-by:[^\n]*github-copilot`), "GitHub Copilot"},
	{regexp.MustCompile(`(?i)co-authored-by:[^\n]*claude`), "Claude (Anthropic)"},
	{regexp.MustCompile(`(?i)co-authored-by:[^\n]*chatgpt`), "ChatGPT (OpenAI)"},
	{regexp.MustCompile(`(?i)(?m)^ai-assisted:\s*true`), "AI"},
	{regexp.MustCompile(`(?i)(?m)^generated-by:`), "AI"},
}

// detectAIAuthors returns deduplicated AI tool labels found in commit trailers.
func detectAIAuthors(commitMsg string) []string {
	seen := map[string]struct{}{}
	var labels []string
	for _, pat := range aiTrailerPatterns {
		if pat.pattern.MatchString(commitMsg) {
			if _, ok := seen[pat.label]; !ok {
				seen[pat.label] = struct{}{}
				labels = append(labels, pat.label)
			}
		}
	}
	return labels
}

const (
	breakingChangesSection = "Breaking Changes"
	featuresSection        = "Features"
	bugFixesSection        = "Bug Fixes"
	otherChangesSection    = "Other Changes"
)

type ReleaseContext struct {
	Version        string
	CurrentVersion string
	Branch         string
	RepositoryURL  string
	Commits        []string
}

type GenerateOptions struct {
	Signature           bool
	NewContributors     bool
	MVP                 bool
	MVPMetric           string
	Contributors        []Contributor
	AIDisclosure        bool
	AIDisclosureBadge   string
	AIDisclosureSection bool
}

type Generator struct {
	now func() time.Time
}

var (
	conventionalHeaderPattern = regexp.MustCompile(`^(\w+)(\([\w-]+\))?(!)?:(.+)$`)
	pullRequestPattern        = regexp.MustCompile(`\(#(\d+)\)`)
)

func New() *Generator {
	return &Generator{now: time.Now}
}

func DefaultGenerateOptions() GenerateOptions {
	return GenerateOptions{
		Signature:           false,
		NewContributors:     true,
		MVP:                 false,
		MVPMetric:           "commits",
		AIDisclosure:        false,
		AIDisclosureBadge:   "🤖",
		AIDisclosureSection: false,
	}
}

func (g *Generator) Generate(ctx ReleaseContext, options ...GenerateOptions) string {
	generateOptions := DefaultGenerateOptions()
	if len(options) > 0 {
		generateOptions = options[0]
	}

	sections := map[string][]string{}
	type aiEntry struct {
		header string
		labels []string
	}
	var aiEntries []aiEntry

	for _, commit := range ctx.Commits {
		section, line := classifyCommit(commit)
		if section == "" || line == "" {
			continue
		}
		if generateOptions.AIDisclosure {
			if labels := detectAIAuthors(commit); len(labels) > 0 {
				badge := generateOptions.AIDisclosureBadge
				if badge == "" {
					badge = "🤖"
				}
				line = line + " " + badge
				if generateOptions.AIDisclosureSection {
					aiEntries = append(aiEntries, aiEntry{firstLine(commit), labels})
				}
			}
		}
		sections[section] = append(sections[section], line)
	}

	var builder strings.Builder
	builder.WriteString("<section class=\"changelog-entry\">\n")
	fmt.Fprintf(&builder, "  <h2>%s <small>%s</small></h2>\n", html.EscapeString(displayVersion(ctx.Version)), g.currentDate().Format("2006-01-02"))

	if meta := metadataLine(ctx); meta != "" {
		fmt.Fprintf(&builder, "  <p>%s</p>\n", html.EscapeString(meta))
	}

	for _, section := range []string{breakingChangesSection, featuresSection, bugFixesSection, otherChangesSection} {
		lines := sections[section]
		if len(lines) == 0 {
			continue
		}

		fmt.Fprintf(&builder, "  <h3>%s</h3>\n", html.EscapeString(section))
		builder.WriteString("  <ul>\n")
		for _, line := range lines {
			fmt.Fprintf(&builder, "    <li>%s</li>\n", html.EscapeString(line))
		}
		builder.WriteString("  </ul>\n")
	}

	firstTimers := firstTimeContributors(generateOptions.Contributors)
	if generateOptions.NewContributors && len(firstTimers) > 0 {
		builder.WriteString("  <h3>New Contributors</h3>\n")
		builder.WriteString("  <ul>\n")
		for _, contributor := range firstTimers {
			builder.WriteString("    <li>")
			builder.WriteString(formatContributorHTML(contributor, ctx.RepositoryURL))
			builder.WriteString(" made their first contribution")
			if reference := contributorFirstContributionHTML(contributor, ctx.RepositoryURL); reference != "" {
				builder.WriteString(" in ")
				builder.WriteString(reference)
			}
			builder.WriteString("</li>\n")
		}
		builder.WriteString("  </ul>\n")
	}

	if generateOptions.MVP {
		if mvp := pickMVP(generateOptions.Contributors, ctx.Commits, generateOptions.MVPMetric); mvp != nil {
			builder.WriteString("  <h3>🏆 Release MVP</h3>\n")
			fmt.Fprintf(&builder, "  <p>%s led the contributors this release.</p>\n", formatContributorHTML(*mvp, ctx.RepositoryURL))
		}
	}

	if generateOptions.AIDisclosure && generateOptions.AIDisclosureSection && len(aiEntries) > 0 {
		builder.WriteString("  <details>\n    <summary>🤖 AI-Assisted Contributions</summary>\n")
		builder.WriteString("    <p>The following changes were co-authored with an AI assistant:</p>\n    <ul>\n")
		for _, e := range aiEntries {
			fmt.Fprintf(&builder, "      <li>%s &mdash; Co-authored with <strong>%s</strong></li>\n",
				html.EscapeString(e.header), html.EscapeString(strings.Join(e.labels, ", ")))
		}
		builder.WriteString("    </ul>\n    <p><em>Disclosed in accordance with project AI-usage policy (L-08 §8).</em></p>\n  </details>\n")
	}

	if generateOptions.Signature {
		builder.WriteString("  <footer><small>Generated by <a href=\"https://semrel.io\">semrel.io</a></small></footer>\n")
	}

	builder.WriteString("</section>")
	return builder.String()
}

func (g *Generator) currentDate() time.Time {
	if g != nil && g.now != nil {
		return g.now()
	}
	return time.Now()
}

func metadataLine(ctx ReleaseContext) string {
	parts := make([]string, 0, 2)
	if ctx.CurrentVersion != "" {
		parts = append(parts, "Previous release: "+displayVersion(ctx.CurrentVersion))
	}
	if strings.TrimSpace(ctx.Branch) != "" {
		parts = append(parts, "Branch: "+strings.TrimSpace(ctx.Branch))
	}
	return strings.Join(parts, " · ")
}

func classifyCommit(commit string) (string, string) {
	trimmed := strings.TrimSpace(commit)
	if trimmed == "" {
		return "", ""
	}

	if breaking, ok := breakingChangeText(trimmed); ok {
		return breakingChangesSection, breaking
	}

	header := firstLine(trimmed)
	matches := conventionalHeaderPattern.FindStringSubmatch(header)
	if len(matches) == 0 {
		return otherChangesSection, header
	}

	if matches[3] == "!" {
		return breakingChangesSection, header
	}

	switch strings.ToLower(matches[1]) {
	case "feat":
		return featuresSection, header
	case "fix", "perf", "revert":
		return bugFixesSection, header
	default:
		return otherChangesSection, header
	}
}

func breakingChangeText(message string) (string, bool) {
	for _, line := range strings.Split(message, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "BREAKING CHANGE:") {
			return trimmed, true
		}
	}
	return "", false
}

func firstLine(message string) string {
	message = strings.TrimSpace(message)
	if message == "" {
		return ""
	}
	parts := strings.SplitN(message, "\n", 2)
	return strings.TrimSpace(parts[0])
}

func displayVersion(version string) string {
	version = strings.TrimSpace(version)
	if version == "" {
		return "Unreleased"
	}
	if strings.HasPrefix(strings.ToLower(version), "v") {
		return version
	}
	return "v" + version
}

func firstTimeContributors(contributors []Contributor) []Contributor {
	firstTimers := make([]Contributor, 0, len(contributors))
	for _, contributor := range contributors {
		if contributor.FirstTime {
			firstTimers = append(firstTimers, contributor)
		}
	}
	return firstTimers
}

func formatContributorHTML(contributor Contributor, repositoryURL string) string {
	label := contributorLabel(contributor)
	if profileURL := contributorProfileURL(contributor, repositoryURL); profileURL != "" {
		return fmt.Sprintf("<a href=\"%s\">%s</a>", html.EscapeString(profileURL), html.EscapeString(label))
	}
	return html.EscapeString(label)
}

func contributorFirstContributionHTML(contributor Contributor, repositoryURL string) string {
	label := strings.TrimSpace(contributor.FirstContributionLabel)
	if label == "" {
		switch {
		case contributor.FirstContributionPR > 0:
			label = fmt.Sprintf("#%d", contributor.FirstContributionPR)
		case strings.TrimSpace(contributor.FirstContributionSHA) != "":
			label = shortReference(contributor.FirstContributionSHA)
		}
	}
	url := contributorFirstContributionURL(contributor, repositoryURL)
	if label == "" {
		if url == "" {
			return ""
		}
		label = "link"
	}
	if url == "" {
		return html.EscapeString(label)
	}
	return fmt.Sprintf("<a href=\"%s\">%s</a>", html.EscapeString(url), html.EscapeString(label))
}

func contributorLabel(contributor Contributor) string {
	if login := strings.TrimSpace(contributor.Login); login != "" {
		return "@" + strings.TrimPrefix(login, "@")
	}
	if name := strings.TrimSpace(contributor.Name); name != "" {
		return name
	}
	return "unknown"
}

func contributorProfileURL(contributor Contributor, repositoryURL string) string {
	if profileURL := strings.TrimSpace(contributor.ProfileURL); profileURL != "" {
		return profileURL
	}
	login := strings.TrimPrefix(strings.TrimSpace(contributor.Login), "@")
	if login == "" {
		return ""
	}
	baseURL := hostRoot(repositoryURL)
	if baseURL == "" {
		return ""
	}
	return baseURL + "/" + login
}

func contributorFirstContributionURL(contributor Contributor, repositoryURL string) string {
	if contributionURL := strings.TrimSpace(contributor.FirstContributionURL); contributionURL != "" {
		return contributionURL
	}
	repositoryURL = strings.TrimRight(strings.TrimSpace(repositoryURL), "/")
	if repositoryURL == "" {
		return ""
	}
	if contributor.FirstContributionPR > 0 {
		return fmt.Sprintf("%s/pull/%d", repositoryURL, contributor.FirstContributionPR)
	}
	if sha := strings.TrimSpace(contributor.FirstContributionSHA); sha != "" {
		return fmt.Sprintf("%s/commit/%s", repositoryURL, sha)
	}
	return ""
}

func hostRoot(repositoryURL string) string {
	repositoryURL = strings.TrimSpace(repositoryURL)
	if repositoryURL == "" {
		return ""
	}
	if idx := strings.Index(repositoryURL, "//"); idx >= 0 {
		rest := repositoryURL[idx+2:]
		if slash := strings.Index(rest, "/"); slash >= 0 {
			return repositoryURL[:idx+2+slash]
		}
	}
	return strings.TrimRight(repositoryURL, "/")
}

func pickMVP(contributors []Contributor, commits []string, metric string) *Contributor {
	if len(contributors) == 0 {
		return nil
	}
	if len(contributors) == 1 {
		return &contributors[0]
	}

	best := &contributors[0]
	bestScore := contributorScore(contributors[0], commits, metric)
	for i := range contributors[1:] {
		contributor := &contributors[i+1]
		if score := contributorScore(*contributor, commits, metric); score > bestScore {
			best = contributor
			bestScore = score
		}
	}
	if bestScore <= 0 {
		return nil
	}
	return best
}

func contributorScore(contributor Contributor, commits []string, metric string) int {
	if contributor.CommitCount > 0 {
		return contributor.CommitCount
	}
	if contributor.FirstContributionPR <= 0 {
		return 0
	}

	score := 0
	for _, commit := range commits {
		matches := pullRequestPattern.FindAllStringSubmatch(commit, -1)
		for _, match := range matches {
			if len(match) < 2 || match[1] != fmt.Sprintf("%d", contributor.FirstContributionPR) {
				continue
			}
			if strings.EqualFold(strings.TrimSpace(metric), "impact") {
				score += impactWeight(commit)
			} else {
				score++
			}
		}
	}
	return score
}

func impactWeight(commit string) int {
	lower := strings.ToLower(commit)
	if strings.Contains(lower, "breaking change") {
		return 3
	}
	matches := conventionalHeaderPattern.FindStringSubmatch(firstLine(commit))
	if len(matches) > 3 && matches[3] == "!" {
		return 3
	}
	if len(matches) > 1 && strings.EqualFold(matches[1], "feat") {
		return 2
	}
	return 1
}
