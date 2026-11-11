# 0008. Deep links run only allow-listed read-only commands

Status: accepted

## Context

Every tappable command is a link `/#<command>`, so a page URL can carry a command. Running arbitrary commands from a URL would let a crafted link change the theme, clear the screen, open windows or start live feeds without the visitor typing anything.

## Decision

On load the host runs a hash line only when it is a single command made of letters, digits, spaces and `. _ / ~ @ : = , + -`, whose first word is in a fixed list of read-only commands. Anything else is placed in the input for the visitor to confirm. Effects of a link-run are ignored. The list and character set are in `docs/render-contract.md`; the host and its tests are the check.

## Consequences

- A link cannot trigger `theme`, `clear`, `open`, `repos`, `packages` or `camera`.
- Adding a command means deciding whether it is read-only before adding it to the list.
