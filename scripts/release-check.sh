#!/usr/bin/env bash
set -euo pipefail

reported_version="$(go run ./cmd/lean | awk '{print $2}')"
tag="${1:-v${reported_version}}"

if [[ ! "$tag" =~ ^v[0-9]+\.[0-9]+\.[0-9]+([.-][0-9A-Za-z.-]+)?$ ]]; then
  echo "Invalid release tag: $tag" >&2
  exit 1
fi

version="${tag#v}"
if [[ "$reported_version" != "$version" ]]; then
  echo "The CLI reports $reported_version but the release tag is $tag" >&2
  exit 1
fi

if ! grep -Eq "^## \[$version\] - [0-9]{4}-[0-9]{2}-[0-9]{2}$" CHANGELOG.md; then
  echo "CHANGELOG.md has no dated $version release" >&2
  exit 1
fi

notes="docs/releases/$tag.md"
if [[ ! -s "$notes" ]]; then
  echo "Release notes are missing: $notes" >&2
  exit 1
fi
