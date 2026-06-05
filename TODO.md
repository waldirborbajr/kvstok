# Release TODO

## Priority
For the next release, prioritize a real security/UX improvement rather than a small CLI tweak.

## High-impact feature to add
### Secure multi-device sync with peer authorization — Issue #305
- You already have P2P sync support.
- The next important step is making it safe and robust:
  - authenticated peer pairing
  - conflict resolution for concurrent updates
  - explicit trust / allowlist instead of blind mDNS sync
- This will make `kvstok` much more usable across laptops/devices while keeping local-first encryption.

## Other strong candidates
### Encrypted backup / export / import improvements — Issue #306
- `export/import` already exists, but a dedicated encrypted backup command with versioning and restore would be very useful.
- Example commands:
  - `kvstok backup`
  - `kvstok restore`
- Include a secure passphrase and integrity check.

### Audit / history and secret rotation — Issue #307
- Add access logs or a command to show recent secret activity.
- Add a `rotate` or `rekey` flow for secret values, especially for API tokens.

## Recommended next release focus
1. Pick one major sync/security feature
   - secure peer trust + sync conflict handling
2. Add one smaller productivity feature
   - encrypted backup/restore or audit trail

> This gives the next release a strong theme: **more secure, reliable multi-device secret management**.
