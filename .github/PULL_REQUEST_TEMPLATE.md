## What and why

<!-- One change per pull request. Say why, not only what. Link the issue if there is one. -->

## How it was checked

- [ ] `gofmt -l .` prints nothing, `go vet ./...` and `make test` pass (or this change touches no Go code)
- [ ] Docs, manifest or version changes: `python3 scripts/sync_docs_from_manifest.py --check`, `python3 scripts/check_doc_links.py` and `python3 scripts/check_marketplace_manifests.py` pass
- [ ] Behavior changes come with a test that fails before the fix and passes after
- [ ] Any new claim in a public doc is backed by a reproducible test or names its corpus

## Not checked

<!-- Be explicit about anything you could not run (for example: a specific OS). -->
