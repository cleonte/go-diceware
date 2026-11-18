# Security policy

## Supported versions

Only the latest release receives fixes.

## Reporting a vulnerability

Please report privately using GitHub's "Report a vulnerability" button on the
[Security tab](https://github.com/cleonte/go-diceware/security/advisories/new)
rather than opening a public issue. Include the command line and input that
trigger the problem. Expect a first reply within a week.

go-diceware generates passphrases from `crypto/rand`; any weakness in the
randomness, a bias in word selection, or a flaw in the embedded wordlists is
in scope. `govulncheck` runs in CI.
