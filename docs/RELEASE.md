# Preparing a release

1. Run `make check` and `make vuln` from the repository root.
2. Commit the intended release scope. The packaging script requires a clean tree.
3. Run `python3 scripts/release.py vMAJOR.MINOR.PATCH`.
4. From `dist/vMAJOR.MINOR.PATCH`, verify `sha256sum -c SHA256SUMS`.
5. Extract the matching platform archive and run `pokedex -version` and
   `pokedex -ephemeral help`. Verify save/restart behavior in an isolated directory.

The script builds Linux, macOS, and Windows binaries for amd64 and arm64 with
CGO disabled. Archives include the README, vector artwork, application license,
and dependency notices. Downloaded binaries have no runtime Go or Python dependency.

The build version comes from the release argument. Packaging does not create a
Git tag, publish to GitHub, or install binaries system-wide. Native tests run locally on Linux and in GitHub Actions on Linux, macOS, and
Windows. Verify the workflow for the exact release commit before publishing;
cross-compilation is not a substitute for native runtime tests.

For a schema change, retain compatibility with existing `version: 1` collections
or add an explicit migration. Never silently reset a collection when loading fails.
