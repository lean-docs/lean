# Repository rules

Use `Type` for owned type discriminators and `type` in JSON. Use descriptive names ending in `Type` for private variables where `type` is a language keyword. Do not introduce `Kind` or `kind` synonyms. Preserve externally defined contracts.

Document every new or changed public engine capability under `docs/`, including supported operations, preservation behavior, export behavior, and limitations. Review those claims against implementation and acceptance evidence before publishing. Existing guides may be extended instead of duplicated.

The long-term goal is complete documentation of every public engine feature. Keep missing coverage explicit. Documentation review is a repository rule, not a separate CI existence check.
