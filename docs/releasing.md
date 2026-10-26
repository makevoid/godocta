# Open-source release readiness

This repository is prepared for a source release under MIT. Publishing the GitHub
repository, creating tags/releases, and shipping desktop binaries are separate
maintainer actions. GitHub Actions is intentionally disabled to avoid starting
hosted jobs before the maintainer has set up a budget.

## Review findings

| Area | Finding and preparation |
| --- | --- |
| License | Added the maintainer-selected MIT license, with project and contributor attribution |
| Onboarding | README now leads with capabilities and offline setup; contributor guide, issue/PR templates, security policy, and credits added |
| Verification | Added `tools/check.py` to import resources, require Godot success summaries, reject script/engine errors, and run Go format/vet/tests |
| Portability | Replaced a hard-coded macOS test path with an OS temporary path; made `bin/run` POSIX-compatible; routed the macOS launcher through it; isolated the Go preview's Unix priority syscall so Windows compiles |
| CI | Saved a disabled workflow for Godot 4.4.1/4.7.2 and Go on Linux/macOS/Windows, including Go 1.22; no hosted jobs are enabled |
| Export | Removed a custom web template path into `.godot/`; standard installed export templates are used; development resources are excluded |
| Credentials | Pattern scan of the reviewed working tree and all 503 unique blobs in 19 reachable commits found no credential matches; this is not a guarantee against every secret format |
| Repository size | About 107 MiB of structures, 3.4 MiB of artwork, and 45 MiB of local Git history at review; largest historical file about 2.7 MiB; no need to migrate these assets to LFS for this release |
| Asset provenance | Added an inventory of generated structures, SVG icons, AI artwork, and optional tooling; normalized image sidecar paths |
| Capture tooling | A failed macOS app launch now fails the wrapper instead of returning success with an empty log |

The `.octa` assets are useful offline examples and cross-language test fixtures.
Keep them versioned. Regenerating unchanged scenes or rewriting existing history
is not part of release preparation.

## Before publishing source

- [ ] Confirm you have redistribution rights for all bundled code and assets,
  particularly `.agents/skills/sunburst-images/` and the AI-generated UI artwork.
  Review the applicable generation-account terms and record any required notices
  in [credits](credits.md). The root license cannot establish third-party ownership.
- [ ] Review the complete commit history and staged changes you intend to publish.
  Repeat a credential scan if additional commits or refs are included; rotate any
  exposed credential before history cleanup. Git author names/emails and historical
  paths are also visible when publishing history.
- [ ] Run `python3 tools/check.py` from a fresh checkout or source archive. Confirm
  all 85 builds and `neighbors.octa` are present, without relying on local imports.
- [ ] Review the current README and controls in the actual desktop app. At minimum,
  exercise opening/saving a file, switching builds, brush/erase, undo/redo, and X-ray.
- [ ] Enable GitHub private vulnerability reporting so the channel in `SECURITY.md`
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
   release. The checked-in preset covers the experimental web preview; desktop
   exports still need platform presets and testing.
2. Include `LICENSE`, asset credits, and the required notices for Godot and its
   bundled components. See [Godot's license guidance](https://docs.godotengine.org/en/stable/about/complying_with_licenses.html).
3. Test exports on each advertised platform, including Save/Load and HiDPI UI.
   For web, test the served export in a browser and state its local-file limitations.
4. Package artifacts under ignored `dist/`, include checksums and release notes,
   and document signing/notarization status for desktop downloads.
5. Publish only the intended source archive and release artifacts. Keep credentials,
   capture logs, local caches, and generated preview directories out of releases.
