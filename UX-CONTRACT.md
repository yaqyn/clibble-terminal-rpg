# Dexrove interaction contract

## Sources and canonical owners

| Capability | Owner | Source | Verification |
| --- | --- | --- | --- |
| Command dispatch | repl.go | Existing CLI command/argument contracts | repl_test.go, cli_test.go |
| Save lifecycle | storage.go, storage_lock.go | Version-1 JSON format and single writer | storage_test.go |
| Terminal presentation | presentation.go | DESIGN.md and brand identity | presentation_test.go |
| Local collection API | companion.go | Read-only saved collection | companion_test.go |
| Search and selection | internal/webui/assets/app.js | Local array filtering and selected creature | Browser and static checks |
| Scrollbar and focus | internal/webui/assets/styles.css | DESIGN.md | Browser and static checks |

The app is single-user. The web companion binds to loopback, rejects foreign Host
and Origin headers, and provides no write endpoint. It reads the same version-1
collection as the CLI. It does not expose the save path, analytics, or credentials.
There are no payments, accounts, deletion actions, or permission settings.

## Flows

- Startup: restore a collection, show concise orientation, and accept commands.
- Catch: fetch, determine outcome, persist success, then offer inspect. Save failure
  rolls back the new in-memory catch. Escape is an outcome, not an error.
- Inspect: show a caught record without a network request; an uncaught name yields
  a correction hint. `collection` is the primary label; `pokedex` remains an alias.
- Browse: load saved records, filter locally, choose a record, and show its details.
  Filtering updates the visible count; selection moves to a visible record when
  its previous record no longer matches. Clear restores the list and search focus.
- Refresh: retain current records until success. On failure, explain the problem,
  retain the last readable state, and enable retry.

## Modes and accessibility

Terminal output is linear; color is optional and disabled by NO_COLOR, dumb
terminals, or the explicit never mode. Pipes have no banner, colors, or prompts.
Web inputs have labels, controls are native buttons, selection uses aria-pressed,
and status uses an aria-live region. Details are a section with a named heading.
No modal, custom listbox, toast, or gesture-only interaction is used. All content
remains available without color, animation, or pointer hover.
