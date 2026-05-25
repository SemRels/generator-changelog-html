# generator-changelog-html

HTML changelog generator plugin for Semantic Release.

Generates HTML changelog output from Semantic Release results.

## Documentation

- Docs (coming soon): <https://github.com/SemRels/semrel/tree/main/docs/plugins/generator-changelog-html>
- Template source: <https://github.com/SemRels/plugin-template>

## Repository Layout

`	ext
cmd/plugin/              Plugin entry point
internal/plugin/         Business logic scaffold
internal/grpc/           gRPC transport scaffold
proto/v1                 Symlink to the SemRel protobuf contract
.github/workflows/       CI, release, and security automation
`

## Development

`ash
go build ./cmd/plugin
go test ./...
`

## Configuration Example

`yaml
plugins:
  - name: generator-changelog-html
    type: generator
    config:
      output: changelog.html
      template: default
      include_styles: true
`

## Status

This repository is bootstrapped from SemRels/plugin-template and is ready for implementation.