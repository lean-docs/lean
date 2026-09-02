#!/usr/bin/env bash
set -euo pipefail

tag="${1:?Usage: build-release.sh <tag> [output-directory]}"
output="${2:-dist}"
version="${tag#v}"
work="$(mktemp -d)"
trap 'rm -rf "$work"' EXIT

mkdir -p "$output"

targets=(
  "linux amd64"
  "linux arm64"
  "darwin amd64"
  "darwin arm64"
  "windows amd64"
  "windows arm64"
)

for target in "${targets[@]}"; do
  read -r goos goarch <<< "$target"
  name="lean_${version}_${goos}_${goarch}"
  package="$work/$name"
  binary="lean"
  if [[ "$goos" == "windows" ]]; then
    binary="lean.exe"
  fi

  mkdir -p "$package"
  CGO_ENABLED=0 GOOS="$goos" GOARCH="$goarch" go build -trimpath -ldflags="-s -w" -o "$package/$binary" ./cmd/lean
  cp LICENSE README.md THIRD_PARTY_NOTICES.md "$package/"
  tar -C "$work" -czf "$output/$name.tar.gz" "$name"
done

(
  cd "$output"
  sha256sum ./*.tar.gz > checksums.txt
  sha256sum -c checksums.txt
)
