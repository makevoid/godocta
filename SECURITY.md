# Security

Octa is a local desktop editor. Its runtime has no network client or telemetry.
The optional artwork-generation tool sends prompts to fal.ai and uses a credential
from the environment or a home-directory key file. It is separate from the app and
is not run by the tests or CI.

## Reporting a vulnerability

Please do not post exploit details, credentials, or sensitive files in public issues.
Use GitHub's **Security → Report a vulnerability** for this repository when available:

https://github.com/makevoid/godocta/security/advisories/new

If private reporting is not yet enabled, open an issue requesting a private contact
without disclosing the vulnerability. Include affected versions/commits, impact, and
a minimal reproducer in the eventual private report. Response times are best effort.

Security fixes target the current default branch. There is no separate maintained
release series yet; use the latest revision and a maintained Godot release.

## Files and credentials

`.octa` loading validates payload length, count, coordinate limits, lattice parity,
span, and palette indices. It reads data rather than executable Godot resources.
Large valid structures can still consume significant memory and CPU.

Never commit API keys, `.env` files, signing credentials, or export credentials.
If a credential reaches Git history, revoke it before considering history cleanup;
deleting the current file alone does not remove previous revisions.
