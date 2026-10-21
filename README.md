# generator-changelog-html

[![Latest Release](https://img.shields.io/github/v/release/SemRels/generator-changelog-html?label=version&color=blue)](https://github.com/SemRels/generator-changelog-html/releases/latest)
[![CI](https://github.com/SemRels/generator-changelog-html/actions/workflows/ci.yml/badge.svg)](https://github.com/SemRels/generator-changelog-html/actions/workflows/ci.yml)
[![License](https://img.shields.io/github/license/SemRels/generator-changelog-html)](LICENSE)
[![OpenSSF Scorecard](https://api.scorecard.dev/projects/github.com/SemRels/generator-changelog-html/badge)](https://scorecard.dev/viewer/?uri=github.com/SemRels/generator-changelog-html)
[![REUSE status](https://api.reuse.software/badge/github.com/SemRels/generator-changelog-html)](https://api.reuse.software/info/github.com/SemRels/generator-changelog-html)

Generates an HTML changelog for the current release.

This plugin is distributed as the standalone Go binary `semrel-plugin-generator-changelog-html`. Semrel executes the binary as a subprocess, provides plugin configuration through `SEMREL_PLUGIN_*` environment variables, provides release context through `SEMREL_*` environment variables, reads standard output, and treats exit code `0` as success and any non-zero exit code as failure. Install the binary in `~/.semrel/plugins/` or anywhere on your `$PATH`.

## Installation

### Binary

```bash
go install github.com/SemRels/generator-changelog-html/cmd/plugin@latest
```

### Docker

Pre-built, multi-platform images (linux/amd64, linux/arm64) are published to the GitHub Container Registry on every release:

```bash
docker pull ghcr.io/semrels/generator-changelog-html:latest
```

Images are signed with [cosign](https://github.com/sigstore/cosign) and include a full SBOM attestation. Verify the signature:

```bash
cosign verify ghcr.io/semrels/generator-changelog-html:latest \
  --certificate-identity-regexp 'https://github.com/SemRels/generator-changelog-html/.github/workflows/container.yml.*' \
  --certificate-oidc-issuer https://token.actions.githubusercontent.com
```


## Configuration

```yaml
plugins:
  - uses: "@semrel/generator-changelog-html"
    phase: generator
    args:
      template: ".semrel/templates/changelog.html.tmpl"
      css_file: ".semrel/templates/changelog.css"
      max_commits: "100"
      first_time_contributors: "true"
      release_mvp: "true"
```

## Options

Every option can be set as an entry under `args` in `.semrel.yaml`. semrel passes each entry to the plugin as the environment variable shown in the table, so you can also set that variable directly (for example in your CI environment) instead of using `args`.

| Argument | Environment variable | Required | Description | Default |
| --- | --- | --- | --- | --- |
| `template` | `SEMREL_PLUGIN_TEMPLATE` | Optional | Path to a custom Go HTML template file. | Built-in template |
| `css_file` | `SEMREL_PLUGIN_CSS_FILE` | Optional | Path to a CSS file included by the HTML template. | None |
| `max_commits` | `SEMREL_PLUGIN_MAX_COMMITS` | Optional | Maximum number of commits to include. | 100 |
| `signature` | `SEMREL_PLUGIN_SIGNATURE` | Optional | Append an opt-in footer element linking to `semrel.io` at the end of the generated HTML entry. | `false` |
| `first_time_contributors` | `SEMREL_PLUGIN_FIRST_TIME_CONTRIBUTORS` | Optional | Append a **New Contributors** section when contributor data is available. Supports the legacy alias `SEMREL_PLUGIN_NEW_CONTRIBUTORS`. | `true` |
| `release_mvp` | `SEMREL_PLUGIN_RELEASE_MVP` | Optional | Append a **🏆 Release MVP** section highlighting the top contributor for the release. Supports the legacy alias `SEMREL_PLUGIN_MVP`. | `false` |
| `contributors_json` | `SEMREL_PLUGIN_CONTRIBUTORS_JSON` | Optional | Legacy contributor JSON payload. `SEMREL_CONTRIBUTORS` is preferred when semrel core provides it. | — |
| `mvp_metric` | `SEMREL_PLUGIN_MVP_METRIC` | Optional | Compatibility toggle for MVP scoring. `commits` picks the highest commit count; `impact` falls back to weighted commit heuristics when counts are unavailable. | `commits` |

## `SEMREL_*` release context used

| Variable | Description |
| --- | --- |
| `SEMREL_VERSION` | Resolved release version for the current run. |
| `SEMREL_TAG_NAME` | Git tag name semrel will create or publish. |
| `SEMREL_NEXT_VERSION` | Next version computed by semrel for the release. |
| `SEMREL_CURRENT_VERSION` | Current version before the new release is applied. |
| `SEMREL_BRANCH` | Git branch associated with the current release run. |
| `SEMREL_REPOSITORY_URL` | Repository URL used to build contributor profile and PR/commit links. |
| `SEMREL_CONTRIBUTORS` | Preferred contributor payload for the current release. |

## Example behavior

The plugin renders a complete HTML changelog using the release metadata and recent commits, then writes the generated document to standard output for semrel to consume.

Set `SEMREL_PLUGIN_SIGNATURE=true` to append this footer element:

```html
<footer><small>Generated by <a href="https://semrel.io">semrel.io</a></small></footer>
```

### New Contributors

When `SEMREL_PLUGIN_FIRST_TIME_CONTRIBUTORS=true` (the default) and contributor data is available, the plugin appends a **New Contributors** section:

```html
<h3>New Contributors</h3>
<ul>
  <li><a href="https://github.com/alice">@alice</a> made their first contribution in <a href="https://github.com/SemRels/generator-changelog-html/pull/42">#42</a></li>
</ul>
```

The preferred data source is `SEMREL_CONTRIBUTORS`. For compatibility, the plugin also accepts `SEMREL_PLUGIN_CONTRIBUTORS_JSON` using the sibling changelog plugin format:

```json
[
  { "name": "Alice", "login": "alice", "pr": 42 }
]
```

When semrel core provides `SEMREL_CONTRIBUTORS`, the plugin also understands richer contributor entries such as:

```json
[
  {
    "name": "Alice",
    "login": "alice",
    "commits": 3,
    "firstContribution": {
      "number": 42,
      "url": "https://github.com/SemRels/generator-changelog-html/pull/42"
    }
  }
]
```

If the contributor payload is missing or invalid JSON, the section is skipped without failing the release.

### 🏆 Release MVP

Set `SEMREL_PLUGIN_RELEASE_MVP=true` to append a **🏆 Release MVP** section for the top contributor in the release:

```html
<h3>🏆 Release MVP</h3>
<p><a href="https://github.com/alice">@alice</a> led the contributors this release.</p>
```

## License

Apache-2.0
