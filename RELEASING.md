# Releasing ali

Releases are cut by pushing a version tag; the [release workflow](.github/workflows/release.yml)
does the rest ([ADR 0042](docs/adr/0042-release-from-version-tags-with-goreleaser-and-provenance.md)).

## Once per repository

1. The Homebrew tap repository `josh-quinn/homebrew-tap` exists.
2. A fine-grained personal access token with **Contents: read and write** on
   `josh-quinn/homebrew-tap` only is stored as the repository secret
   `HOMEBREW_TAP_TOKEN` (`gh secret set HOMEBREW_TAP_TOKEN`). Without it the release
   is still published, but the Homebrew cask is skipped.

## Each release

1. Make sure `main` is green in CI and your clone is up to date (`git switch main && git pull`).
2. Move the `Unreleased` notes in `CHANGELOG.md` under a new version heading, in a
   small `docs(changelog)` commit through a pull request.
3. Choose the version with [semantic versioning](https://semver.org): before 1.0, a
   minor bump (0.2.0) for new features or anything incompatible, a patch bump
   (0.1.1) for fixes.
4. Tag and push (the pre-push secret scan runs):

   ```sh
   git tag -a v0.2.0 -m "ali v0.2.0"
   git push origin v0.2.0
   ```

5. Watch the release workflow. It refuses tags that are not on `main`, runs the
   tests, publishes the GitHub Release with archives, checksums and generated release
   notes, attests build provenance, and updates the Homebrew cask.
6. Check the release:

   ```sh
   gh release view v0.2.0
   gh attestation verify ali_0.2.0_darwin_arm64.tar.gz --repo josh-quinn/ali
   brew update && brew upgrade --cask ali   # or: brew install --cask josh-quinn/tap/ali
   ali --version
   ```

If a release goes wrong, delete the GitHub Release and the tag, fix the problem on
`main`, and release the next patch version rather than reusing the number.
