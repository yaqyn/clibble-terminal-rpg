# Using Clibble

Clibble is an original, offline wilderness adventure. Start `clibble` for the interactive
RPG screen, use `clibble -plain` for the command prompt, run `clibble <command>` for a single operation, or `clibble serve` to browse a
saved collection in your browser.

## Playing at the terminal

Launch `clibble` in a supported terminal. Use up/down (or j/k) and Enter/Space to
choose **Start exploring**, **Travel**, a trail, a wildling, then **Offer a catch seed**. The narration
addresses you as the explorer. You can retry after an escape; after a catch, choose
**Inspect your catch**. Your collection and its field records also work offline.

Esc/Backspace goes back. Left/right changes destination pages. `t`, `c`, and `?`
open travel, collection, and help. `r` retries a failed action. Up/down scrolls
long records/help; Home/End jump to the first/last item. Menus scroll to keep the
selection visible. Return from an encounter or record restores your selection.
Tab pages through narration or errors that do not fit the message box.
`q` or Ctrl+C leaves. During a catch, additional actions are disabled until the
outcome is settled. Shutdown cancels requests and waits for any active save before
releasing the collection lock.

The minimum screen is 40 × 18. Resize a smaller terminal or use `-plain`. The RPG
screen uses the alternate terminal buffer and restores it on exit. It has no
mouse requirement, battles, or tile-based movement. Gentle rain, canopy shimmer,
and particles run at 12 fps. **Forest settings** saves a motion-off preference;
`-no-motion` disables motion for one session. Travel means choosing a bundled trail; your current location lasts this session only.
`-ephemeral` keeps catches only until you leave. Dumb terminals automatically use
the command interface; pipes remain plain text.

## Command prompt and one-shot commands

`map` lists six trails. `explore mosslight-woods` lists the wildlings living there.
Use `catch mosskin` to offer a catch seed. Shyer wildlings are harder to catch.
An escape leaves your journal unchanged, but uses one seed and one try.

You start with 12 seeds and three tries per wildling per 30-minute encounter
window. Tries are shared across trails, command mode, RPG mode, and restarts.
One seed regenerates every five minutes up to 12, including while away.
The first offer starts the encounter window; once it expires, three tries become
available again. Failed data lookup or save spends nothing.
`-ephemeral` intentionally starts a fresh, temporary game. These are local
single-player rules; editing the save or system clock can bypass them.

A successful catch saves immediately. `inspect mosskin` reads the stored record;
`collection` lists caught names, and `count` prints the number of distinct catches offline. Interactive records convert height/weight to
metres/kilograms. Plain CLI output uses decimetres/hectograms for save compatibility.

## The local companion

```bash
clibble serve
```

Open the printed address (normally http://127.0.0.1:8787). The interface searches
names and habitats, selects a record, and shows its dimensions and traits.
**Clear** removes the filter and returns focus to search. **Refresh collection**
loads catches saved since the last refresh. An unsuccessful refresh retains the
last readable records and allows retry.

The view is read-only. Run catches in a separate CLI session, then refresh the
browser. The companion does not hold a writer lock or modify your JSON file.
Use Ctrl+C in its terminal to stop the local server. There is no network binding
option: this interface is not designed for public or shared hosting.

```bash
clibble -port 9000 serve
clibble -data ./collections/first.json serve
```

A missing collection shows an empty state; a corrupt save produces an actionable
error. The web view never returns the save path or detailed filesystem errors.

## Collection locations

New collections use:

- Linux: `$XDG_CONFIG_HOME/clibble/collection.json`, or `~/.config/clibble/collection.json`.
- macOS: `~/Library/Application Support/clibble/collection.json`.
- Windows: `%AppData%\clibble\collection.json`.

`-data` selects another file; parent directories are created automatically by the
CLI. `-ephemeral` starts a temporary CLI collection, and cannot be combined with
`-data` or `serve`. New Unix directories/files use owner-only permissions; Windows
uses the current user's filesystem controls.

## Save compatibility

Clibble reads version-1 collections and version-2 game saves. The first successful
catch or settings change upgrades a legacy file atomically, preserving every
record and adding supplies once. Version 2 stores the journal, supply refill
time, per-wildling encounter windows, and motion preference together. It looks under the platform
configuration directory in this order: `clibble/collection.json`, then the previous
`dexrove/collection.json`, then `pokedex/pokedex.json`. The first existing file is
used with its original lock. Nothing is moved, merged, or overwritten during
discovery. Both prior executable names were renamed; existing saves still work.
Both CLI and companion resolve the same path.

To change the location deliberately, close CLI writers, back up the old JSON, copy
it to the desired location, and use `-data` with both CLI and companion. The old
`pokedex` command remains a collection alias. The executable is now `clibble`; v1.0.0 (Pokedex) and v1.1.0 (Dexrove) remain
available as historical software.

## Terminal and scripts

The `-plain` interactive view shows a brief welcome, grouped help, creature profiles, and
collection summaries. It uses your existing terminal font/background and never
switches to a fullscreen display. On supported terminals, the line editor supports
Tab command completion, up/down history, cursor movement, and editing shortcuts.
Its last 100 commands stay in memory only. Terminal settings are restored on exit. Narrow terminals use linear layouts.

`clibble` opens the colored forest directly. Use `-color never` for a static,
color-free RPG. Command mode honors `NO_COLOR`; `TERM=dumb` uses command mode.
The explicit `always` option is only for interactive presentation; redirected
results stay plain.

```bash
clibble -plain -color never
clibble -no-motion
clibble -data ./collection.json inspect mosskin
printf 'collection\nexit\n' | clibble
```

Shell flags precede the command. One-shot failures return status 1. Batches keep
processing commands, then return status 1 if any command failed. Interactive errors
leave the prompt usable. An escaped catch returns success because the attempt ran.

## Recovery

- **Save already in use:** another CLI writer owns that collection. Close it or choose another `-data` file. The browser companion can remain open.
- **Invalid save:** preserve the file and restore a backup or correct its JSON. Clibble refuses to reset a malformed collection silently.
- **Unknown wildling or trail:** use `map`, then `explore mosslight-woods`. Names use letters, digits, and hyphens.
- **No seeds or tries:** the message shows when you can return. Reopening a menu does not reset the budget.
- **Port in use:** choose another `-port`. The companion still binds only to 127.0.0.1.
- **No browser records:** confirm CLI and companion use the same collection path; refresh after catching.
