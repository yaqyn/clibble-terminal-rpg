# Using Clibble

Clibble is an independent Pokemon field guide. Start `clibble` for the interactive
CLI, run `clibble <command>` for a single operation, or `clibble serve` to browse a
saved collection in your browser.

## Finding and catching Pokemon

`map` lists 20 location areas. Repeat it for the next page; `mapb` moves back.
`explore pastoria-city-area` lists the Pokemon found there. Use `catch pidgey` to
attempt a catch. Base experience affects the probability; higher experience means
a lower chance. An escape does not change your collection.

A successful catch saves immediately. `inspect pidgey` reads the stored record;
`collection` lists caught names, and `count` prints the number of distinct catches offline. Interactive records convert height/weight to
metres/kilograms. Plain CLI output keeps PokéAPI's original decimetres/hectograms.

## The local companion

```bash
clibble serve
```

Open the printed address (normally http://127.0.0.1:8787). The interface searches
names and types, selects a record, and shows its dimensions and base stats.
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

Clibble retains the version-1 JSON collection format. It looks under the platform
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

The interactive view shows a brief welcome, grouped help, creature profiles, and
collection summaries. It uses your existing terminal font/background and never
switches to a fullscreen display. On supported terminals, the line editor supports
Tab command completion, up/down history, cursor movement, and editing shortcuts.
Its last 100 commands stay in memory only. Terminal settings are restored on exit. Narrow terminals use linear layouts.

`-color never`, `NO_COLOR`, or `TERM=dumb` disables automatic color. The explicit
`always` option is only for interactive presentation; redirected results stay plain.

```bash
clibble -color never
clibble -timeout 5s catch pidgey
clibble -data ./collection.json inspect pidgey
printf 'collection\nexit\n' | clibble
```

Shell flags precede the command. One-shot failures return status 1. Batches keep
processing commands, then return status 1 if any command failed. Interactive errors
leave the prompt usable. An escaped catch returns success because the attempt ran.

## Recovery

- **Save already in use:** another CLI writer owns that collection. Close it or choose another `-data` file. The browser companion can remain open.
- **Invalid save:** preserve the file and restore a backup or correct its JSON. Clibble refuses to reset a malformed collection silently.
- **HTTP 404:** check the Pokemon or area identifier. Names use letters, digits, and hyphens.
- **Request timeout:** retry later or increase `-timeout`. Map pagination stays unchanged after failures.
- **Port in use:** choose another `-port`. The companion still binds only to 127.0.0.1.
- **No browser records:** confirm CLI and companion use the same collection path; refresh after catching.
