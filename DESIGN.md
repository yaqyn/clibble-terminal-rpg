---
version: alpha
name: Dexrove
description: A Pokemon field guide shaped by exploration, clear records, and local ownership.
colors:
  primary: "#315dcd"
  text: "#192b46"
  background: "#edf2f9"
  surface: "#ffffff"
  muted: "#53637a"
  terminal-accent: "#88baff"
typography:
  sans:
    fontFamily: "Manrope, system-ui, sans-serif"
  mono:
    fontFamily: "ui-monospace, monospace"
rounded:
  DEFAULT: "0.75rem"
  control: "0.5rem"
spacing:
  section-gap: "2rem"
  page-max: "76rem"
components:
  navigation:
    backgroundColor: "{colors.background}"
    textColor: "{colors.text}"
  creature-record:
    backgroundColor: "{colors.surface}"
    textColor: "{colors.text}"
  caption:
    textColor: "{colors.muted}"
  terminal:
    textColor: "{colors.terminal-accent}"
---

# Dexrove design system

## Creative direction

The reference is a contemporary field atlas: a route, a clear species record,
and space to notice what matters. The memorable feature is the open-route D mark.
The terminal is the primary product; the read-only browser view is a complement.
The interface supports that motif with calm cobalt and mist surfaces, rather
than reproducing a game console or turning the app into a dashboard of metrics.

Dexrove means dex + rove. “Explore. Catch. Keep.” gives the three actions equal
weight. The audience is terminal-comfortable Pokemon fans; the web companion
makes saved records easier to browse. English is the supported interface language;
this is not a Japan-market product merely because its data comes from Pokemon.

## Ownership and mapping

[internal/brand/identity.json](internal/brand/identity.json) is the canonical
identity and palette. This file describes its roles. The Go brand package embeds
those tokens. `scripts/brand_assets.py` generates web CSS variables and vector
marks from the same source. The terminal renderer maps Azure to headings on a
terminal's existing background; normal text inherits the terminal foreground.
The README cover is generated raster artwork, with the exact name visible.

## Typography, layout, and states

The web view uses locally bundled Manrope with a system sans fallback. Headlines
are sentence case; numbers align with tabular figures. Body lines stay below
80 characters. The page uses an asymmetric field-guide composition: a collection
index and an adjacent creature record on wide screens, stacked on narrow screens.
Records stay readable at 320px; details have no fixed viewport height.

The CLI uses the user's terminal font. It does not clear the screen automatically, animate, or hide the cursor.
The line editor temporarily uses raw input and restores the previous terminal
state on exit. It provides Tab completion and a bounded in-memory history. Narrow terminals get linear records.
Color is optional and never carries meaning alone. Piped output remains plain.

Web controls have explicit hover, focus, selected, disabled, and loading states.
Search has a clear action; filtering distinguishes an empty collection from no
matches. Failed refreshes preserve the last readable collection. Status is textual
and announced. The web app is read-only; collecting stays in the CLI.

## Shape and restraint

The D mark is a curved route with an open spine. Use it once in the header, not
as decoration on every record. Rounded control corners belong to interactive
controls; records are separated by spacing and purposeful borders. No mascot,
third-party creature artwork, neon effects, or ornamental terminal frames.

## Verification

Use the Go test suite for save compatibility, command modes, and local web API
boundaries. Run the design-context linter and premium static audit. Inspect the
terminal at wide/narrow sizes, color off, and piped output. Browser verification
covers filtering, detail selection, keyboard focus, empty/error/loading states,
narrow layout, and reduced motion when a browser connection is available.
