# Clibble interaction contract

## Sources and canonical owners

| Capability | Canonical owner | Source of truth | Allowed variants | Verification |
| --- | --- | --- | --- | --- |
| Command dispatch | repl.go | Existing CLI command contract | interactive / one-shot / batch | repl_test.go, cli_test.go |
| Save lifecycle | storage.go | Version-2 game schema; version-1 import | default / custom / legacy | storage_test.go, presentation_test.go |
| RPG screen and keyboard | game.go, game_view.go, forest.go | DESIGN.md | narrow / wide / loading / retry / record | game_test.go |
| Command presentation | presentation.go | DESIGN.md | narrow / wide / plain | presentation_test.go |
| Catch transaction | catch.go | progress.go rules and atomic version-2 saves | success / escape / save failure / cancellation | repl_test.go, game_test.go |
| Collection API | companion.go | Read-only local save | empty / populated / unavailable | companion_test.go |
| Search and selection | internal/webui/assets/app.js | UX-CONTRACT.md | filtered / empty / selected | tests/companion.test.mjs |
| Scrollbar | internal/webui/assets/styles.css | DESIGN.md | normal / forced-colors | browser and static checks |


The app is single-user. The web companion binds to loopback, rejects foreign Host
and Origin headers, and provides no write endpoint. It reads the same version-1 or version-2
collection as the CLI. It does not expose the save path, analytics, or credentials.
There are no payments, accounts, deletion actions, or permission settings.

## Flows

- Startup: restore a collection and enter the RPG title screen, then Start exploring in a supported TTY.
  Use `-plain` for command entry; dumb terminals and pipes use command mode.
- RPG navigation: arrows/j/k move, Enter/Space selects, Esc/Backspace goes back.
  Travel opens paged destinations; selecting one explores its encounter list.
  Select an encounter, attempt a catch, then inspect the saved record. Returning
  restores the relevant selection. Tab pages through long narration/errors; text
  is never available only in a truncated message. Empty areas and collections offer a next step.
- Background work: show narration while pending, disable repeated actions, allow
  cancelling travel/exploration, and reject stale response IDs. Catch/save settles
  before another action. Failures retain prior location/collection and offer retry.
  Shutdown cancels requests and drains workers before releasing the writer lock.
- Catch: look up an original wildling, determine outcome, persist the complete game,
  then offer inspect. Escapes also persist consumed supplies/tries. Save failure
  rolls back the journal and budget together. `progress.go` owns game balance;
  `docs/GUIDE.md` explains the rules. Reopening menus or switching trails does not
  replenish tries. Escape is an outcome, not an error.
- Forest settings: motion preference saves in the same transaction format. The UI
  waits for the save before applying it; failed saves retain the previous setting.
  Pending catches and settings cannot be cancelled halfway through a write.
- Inspect: show a caught record without a network request; an uncaught name yields
  a correction hint. `collection` is the primary label; `pokedex` remains an alias.
- Browse: load saved records, filter locally, choose a record, and show its details.
  Filtering updates the visible count; selection moves to a visible record when
  its previous record no longer matches. Clear restores the list and search focus.
- Refresh: retain current records until success. On failure, explain the problem,
  retain the last readable state, and enable retry.

## Modes and accessibility

The RPG screen is fully keyboard operated, resizes without hiding selection, and
restores the normal terminal on exit. `-plain` and script output stay linear; the RPG opens in color unless `-color never` is selected. Command output honors
NO_COLOR and dumb terminals. Pipes have no banner, colors, or prompts.
Web inputs have labels, controls are native buttons, selection uses aria-pressed,
and status uses an aria-live region. Details are a section with a named heading.
No modal, custom listbox, toast, or gesture-only interaction is used. All content
remains available without color, animation, or pointer hover.
