# Open-source release readiness

Octa ships source under MIT and desktop downloads for macOS universal and Linux
x86-64. The application version lives in `project.godot`; release tags use `v` plus
that version. GitHub Actions remains intentionally disabled. Desktop packages are
built locally using Docker and published as GitHub Release assets.

## Desktop release procedure

1. Update `config/version` in `project.godot` and prepare release notes.
2. Start Docker and run `python3 tools/release.py`. The Dockerfile in
   `tools/release/` supports arm64 and amd64 hosts, downloads the official Godot
   4.7.2 editor and export templates, and verifies their pinned SHA-512 hashes.
   It runs Godot tests, verifies the exported build list and bytes, and creates
   `dist/<version>/Octa-macos.zip`, `Octa-linux-x86_64.tar.gz`, and `SHA256SUMS.txt`.
3. Extract and smoke-test both packages. `tools/release/smoke.gd` exercises the
   exported scene, gallery, Fable Balloon Festival, primitives, Save/Load and X-ray.
   Run the matching Godot editor with `--main-pack <exported.pck>` and
   `--script tools/release/smoke.gd`, followed by
   `-- --expect-version=0.4.0` (substitute the release version). With a real renderer,
   also pass `--capture=/absolute/path.png` to inspect the UI. Keep macOS captures
   hidden using `open -g -j`; Linux can use Xvfb and Mesa inside an amd64 container.
   Do not describe software-rendered or headless results as physical GPU/HiDPI testing.
   Official release templates disable external script overrides, so also launch the
   actual exported executables with `--quit-after` to verify standalone startup.
   On Apple silicon, x86-64 Mesa under QEMU can fail in LLVM or time out. In that
   case, run the packaged-content check with the native Linux editor and verify
   the x86-64 executable headlessly; report native x86-64 GPU testing as outstanding.
4. Verify the macOS app's universal architectures and ad-hoc signature with `file`
   and `codesign --verify --deep --strict`. It is not Apple-notarized. State this in
   the release notes and include the standard Privacy & Security opening guidance.
5. Commit the intended source changes, tag that commit, and push the branch and tag.
   Create a draft release, upload only the two archives and checksum file, inspect
   its assets, then publish it. The README buttons use stable asset names under
   `releases/latest/download/`, so every release should retain those names.

For example, after checking the artifacts and writing `dist/0.4.0/release-notes.md`:

```sh
git tag -a v0.4.0 -m "Octa 0.4.0"
git push origin main v0.4.0
gh release create v0.4.0 --verify-tag --draft --title "Octa 0.4.0" \
  --notes-file dist/0.4.0/release-notes.md \
  dist/0.4.0/Octa-macos.zip dist/0.4.0/Octa-linux-x86_64.tar.gz \
  dist/0.4.0/SHA256SUMS.txt
gh release edit v0.4.0 --draft=false --latest
```

Release archives include the project license, credits, controls, launch instructions,
and the matching upstream Godot license and third-party notices. Experiment files
stay local under `experiments/`; captures, caches and development tools are excluded.

## Review findings

| Area | Finding and preparation |
| --- | --- |
| License | Added the maintainer-selected MIT license, with project and contributor attribution |
| Onboarding | README now leads with capabilities and offline setup; contributor guide, issue/PR templates, security policy, and credits added |
| Verification | Added `tools/check.py` to import resources, require Godot success summaries, reject script/engine errors, and run Go format/vet/tests |
| Portability | Replaced a hard-coded macOS test path with an OS temporary path; made `bin/run` POSIX-compatible; routed the macOS launcher through it; isolated the Go preview's Unix priority syscall so Windows compiles |
| CI | Saved a disabled workflow for Godot 4.4.1/4.7.2 and Go on Linux/macOS/Windows, including Go 1.22; no hosted jobs are enabled |
| Export | Removed a custom web template path into `.godot/`; standard installed export templates are used; development resources are excluded |
| Credentials | Pattern scans at review and again before publishing (438 text blobs in 22 reachable commits, 2026-10-26) found no credential matches; this is not a guarantee against every secret format |
| Repository size | About 107 MiB of structures, 3.4 MiB of artwork, and 45 MiB of local Git history at review; largest historical file about 2.7 MiB; no need to migrate these assets to LFS for this release |
| Asset provenance | Added an inventory of generated structures, SVG icons, AI artwork, and optional tooling; normalized image sidecar paths |
| Capture tooling | A failed macOS app launch now fails the wrapper instead of returning success with an empty log |

The `.octa` assets are useful offline examples and cross-language test fixtures.
Keep them versioned. Regenerating unchanged scenes or rewriting existing history
is not part of release preparation.

## Before publishing source

- [x] Confirm you have redistribution rights for all bundled code and assets,
  particularly `.agents/skills/sunburst-images/` and the AI-generated UI artwork.
  Confirmed by the maintainer on 2026-10-26 and recorded in [credits](credits.md).
- [x] Review the complete commit history and staged changes you intend to publish.
  Repeat a credential scan if additional commits or refs are included; rotate any
  exposed credential before history cleanup. Git author names/emails and historical
  paths are also visible when publishing history.
- [x] Run `python3 tools/check.py` from a fresh checkout or source archive. Confirm
  all 100 builds and `neighbors.octa` are present, without relying on local imports.
- [ ] Review the current README and [controls](controls.md) in the actual desktop app. At minimum,
  exercise opening/saving a file, switching builds, brush/erase, undo/redo, and X-ray.
- [x] Enable GitHub private vulnerability reporting so the channel in `SECURITY.md`
  is available. Until then its documented fallback does not disclose vulnerability details.
- [ ] Set the repository description and topics, choose an initial version/tag,
  and write release notes describing supported platforms and known limitations.

The locally verified desktop platform is macOS. Do not describe Linux/Windows GUI
behavior or the web preview as verified solely because headless tests or Go
cross-compilation pass. A source release can state these limits explicitly.

## Enabling CI later

No `.yml` or `.yaml` workflow is active under `.github/workflows/`. Local checks
do not require GitHub credits.

When the maintainer is ready to enable hosted jobs:

```sh
mv .github/workflows/ci.yml.disabled .github/workflows/ci.yml
```

The enabled workflow runs on pushes, pull requests, and manual dispatch. Review
runner costs, version pins, and branch protection requirements before committing
that rename. The Godot matrix downloads official release archives and verifies
their published SHA-512 checksums. Action references are pinned to full commit IDs.

Dependabot is also stored as `.github/dependabot.yml.disabled`. Rename it to
`.github/dependabot.yml` separately to opt into monthly action-update pull requests.
Neither template has been executed on hosted runners during this preparation.

## Before shipping binaries or a web preview

1. Use official export templates matching the Godot version used to validate the
   release. The checked-in presets cover macOS universal, Linux x86-64 and the
   experimental web preview.
2. Include `LICENSE`, asset credits, and the required notices for Godot and its
   bundled components. See [Godot's license guidance](https://docs.godotengine.org/en/stable/about/complying_with_licenses.html).
3. Test exports on each advertised platform, including Save/Load and HiDPI UI.
   For web, test the served export in a browser and state its local-file limitations.
4. Package artifacts under ignored `dist/`, include checksums and release notes,
   and document signing/notarization status for desktop downloads.
5. Publish only the intended source archive and release artifacts. Keep credentials,
   capture logs, local caches, and generated preview directories out of releases.
