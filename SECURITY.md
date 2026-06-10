# Security Policy

## Supported versions

This project is pre-1.0. Security fixes are applied to the latest released
`v0.x` minor; there are no long-term support branches yet.

| Version | Supported |
| ------- | --------- |
| latest `v0.x` | ✅ |
| older         | ❌ |

## Reporting a vulnerability

Please report security issues **privately** — do not open a public issue
for anything exploitable.

- Preferred: open a private security advisory on the repository
  ("Security" → "Report a vulnerability").
- If that is unavailable, contact the maintainers privately through the
  repository host.

Please include enough detail to reproduce: affected version, a minimal
proof of concept, and the impact you observed. We'll acknowledge the
report, investigate, and coordinate a fix and disclosure timeline with
you.

## Scope

This is a client library; it makes outbound HTTPS requests to an
Audiobookshelf server you configure. The most relevant areas:

- **Credential handling** — tokens and API keys are sent as bearer
  headers and are never logged by this library. `WithInsecureSkipVerify`
  disables TLS verification and is intended for local testing only; never
  use it against a server reachable over an untrusted network.
- **Response parsing** — untrusted server responses are decoded into Go
  types; report any panic or unbounded-allocation triggered by a
  malformed response.

Vulnerabilities in Audiobookshelf itself should be reported upstream to
the [Audiobookshelf project](https://github.com/advplyr/audiobookshelf).
