# Releasing Lean

Lean releases come from a versioned commit on `main`. The version must agree in
the CLI, changelog, release notes, and Git tag.

## Prepare

1. Update the version returned by `lean.Version`.
2. Move completed changes from `Unreleased` into a dated changelog section.
3. Add `docs/releases/v<version>.md`.
4. Run `./scripts/release-check.sh v<version>`.
5. Merge the release change after CI passes.

## Publish

Create and push the tag from the released commit on `main`:

```bash
git switch main
git pull --ff-only
git tag -a v0.1.0-alpha.1 -m "Lean v0.1.0-alpha.1"
git push origin v0.1.0-alpha.1
```

The release workflow confirms that the tag points into `main`, runs the full
test suite, builds CLI archives for Linux, macOS, and Windows, writes checksums,
and publishes the GitHub release from the matching release notes.
