# Changelog

## [Unreleased]

### Added
- Project structure following Go conventions
- IR type definitions (`pkg/ir/`) with JSON serialization
- Style resolution with `basedOn` inheritance chain
- Test suite: Cluster 0 (bootstrap) and Cluster 1 (IR fundamentals) — passing
- Test suite: Cluster 2 (Markdown parser) — failing, awaiting implementation
- CI workflow with build, vet, test, and lint steps
- Architecture documentation (`ARCHITECTURE.md`)
- Requirements and test specification (`SRS.md`)
