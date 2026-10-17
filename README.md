![Clibble — A little wild in your terminal.](readme-assets/cover-original-world.webp)

# Clibble — a little wilderness RPG in your terminal

Explore a rain-soaked forest, meet wildlings, and fill a field journal of your own.

Clibble brings an arrow-driven adventure to your terminal. Travel between locations,
inspect your catches offline, or browse saved records in the optional, read-only web companion.
Six trails and ten original wildlings are bundled into the app. Play offline, with no account or API key.

[Published releases](https://github.com/yaqyn/clibble-terminal-rpg/releases/latest) · [Start from source](#-installation) · [User guide](docs/GUIDE.md)

## <img src="readme-assets/icons/features.svg" width="24" height="24" alt=""> Features

- **Play in the terminal.** Choose with arrows and Enter, follow the narration, and inspect your catches.
- **Follow a new trail.** Browse forest trails and discover the wildlings that live there.
- **Make the catch.** Try your luck, save successful catches, and return to their records offline.
- **See your collection.** Search by name or habitat in the read-only web companion; compare stats and measurements.
- **Use it your way.** Run one-shot commands or pipe a batch; local saves keep your collection between sessions.

## <img src="readme-assets/icons/installation.svg" width="24" height="24" alt=""> Installation

The original-world forest build is currently available from source. Published
v1.x archives predate this redesign. For those older versions, download the archive
for your operating system and architecture from the
[releases page](https://github.com/yaqyn/clibble-terminal-rpg/releases/latest), verify it against
`SHA256SUMS`, and extract it. Run `clibble` on Linux/macOS or `clibble.exe` on Windows.
The compiled app includes the web interface and font; it needs no Go, Node, or Python runtime.
The world, terminal scenery, companion, and font are bundled. Playing requires no internet connection.

To build from source, install the Go version declared in [go.mod](go.mod) or newer:

```bash
git clone https://github.com/yaqyn/clibble-terminal-rpg.git
cd clibble-terminal-rpg
go build -trimpath -o clibble .
```

On Windows, use `-o clibble.exe`. Go downloads pinned build dependencies on the first build.

## <img src="readme-assets/icons/launch.svg" width="24" height="24" alt=""> Quick start

```bash
./clibble
```

Choose **Start exploring**, then **Travel**. Visit Mosslight Woods, approach
a wildling, and choose **Offer a catch seed**. After a successful catch, choose
**Inspect your catch** to open its saved record. Your catches save automatically.

| Key | Action |
| --- | --- |
| Up / Down or `j` / `k` | Move through choices; scroll records and help |
| Enter / Space | Choose the highlighted action |
| Esc / Backspace | Go back; cancel travel or exploration |
| Left / Right | Browse destination pages |
| `t` / `c` / `?` | Open travel / collection / field guide |
| Tab | Read the next page of a longer message |
| `r` | Retry a failed action |
| `q` / Ctrl+C | Leave safely |

The RPG screen needs a terminal at least 40 columns × 18 rows. It adapts to resizing
and restores your terminal when you leave. Location and menu position last for the
session; your journal, supplies, encounter tries, and motion preference remain saved.

You begin with **12 catch seeds**. Run `./clibble supplies` to inspect the refreshed pouch and next refill time without spending seeds or saving changes. Each resolved offer costs one seed and one try,
including escapes. You get **three tries per wildling in a 30-minute window**,
shared across trails and restarts. One seed replenishes every five minutes, up
to 12. API/data errors and failed saves spend nothing.

The green pixel forest has gentle rain, canopy movement, and drifting motes at
12 fps. Choose **Forest settings → Motion: off** to save a still-forest preference,
or launch with `-no-motion` for this session. `./clibble` opens the colored forest directly, even when `NO_COLOR` is inherited.
Use `-color never` for a static character backdrop. `-plain` keeps the linear
prompt and honors `NO_COLOR`.

For the command prompt, use `./clibble -plain`. It keeps grouped help, Tab command
completion, and the last 100 commands in memory. Scripts and one-shot commands
continue to work directly.

## <img src="readme-assets/icons/usage.svg" width="24" height="24" alt=""> Usage

| Command | What it does |
| --- | --- |
| `help` | Show the field guide |
| `map` / `mapb` | Browse the next / previous page of location areas |
| `explore <area>` | List wildlings in a location area |
| `catch <wildling>` | Attempt a catch; shyer wildlings are harder to catch |
| `inspect <wildling>` | Read a caught wildling's record without a network request |
| `collection` | List your caught wildlings |
| `count` | Print the number of distinct caught wildlings offline |
| `serve` | Start the local web companion; run this from your shell |
| `exit` | Leave the interactive session |

Commands and identifiers are case-insensitive. `collection` opens the journal; the historical command alias remains supported. Use `where <wildling>` to list its bundled trails offline without spending seeds. Use single commands or pipe a batch:

```bash
./clibble inspect mosskin
./clibble collection
printf 'collection\nexit\n' | ./clibble
```

Pipes omit banners, color, and prompts. Results go to stdout; errors go to stderr.
A failed command produces a nonzero exit status. Batches continue after errors
and report failure at the end; an escaped catch is a valid outcome, not a command error.

### Optional browser companion

To open the collection in a browser, run this in another terminal:

```bash
./clibble serve
```

Open **http://127.0.0.1:8787**. Search your saved wildlings by name or habitat, then choose
one to read its record. Catch in the CLI and use **Refresh collection** to see new
records. The companion is local and read-only; it does not expose writes or share your collection online.

See the [user guide](docs/GUIDE.md) for flags, storage locations, recovery, and examples.

## <img src="readme-assets/icons/configuration.svg" width="24" height="24" alt=""> Configuration

| Flag | Default | Purpose |
| --- | --- | --- |
| `-data <path>` | Platform user directory | Choose a collection file |
| `-plain` | Off | Use the command prompt instead of the RPG screen |
| `-ephemeral` | Off | Use a temporary CLI collection without saves |
| `-timeout <duration>` | `10s` | Bound each API request |
| `-color auto\|always\|never` | `auto` | Control interactive terminal color |
| `-port <number>` | `8787` | Select the local companion's port |
| `-version` / `-h` | — | Print version / flags |

Flags precede commands: `./clibble -data ./collection.json serve`.
Existing Pokedex saves remain accessible automatically; no collection is moved or reset.
See [save compatibility](docs/GUIDE.md#save-compatibility).

## <img src="readme-assets/icons/architecture.svg" width="24" height="24" alt=""> Architecture

The same Go binary runs the CLI and serves the companion. A bundled original world
feeds both terminal modes; collections and gameplay progress use version-2 local JSON. The web
view reads that collection and filters records locally, without an external service.

Read the [architecture guide](docs/ARCHITECTURE.md), [design system](DESIGN.md), and
[interaction contract](UX-CONTRACT.md). Brand tokens live in one
[identity file](internal/brand/identity.json) and generate the marks and palette adapters.

## <img src="readme-assets/icons/verification.svg" width="24" height="24" alt=""> Verification

```bash
go test -race ./...
go vet ./...
go run golang.org/x/vuln/cmd/govulncheck@v1.9.0 ./...
```

For web interaction tests, install the development-only Node dependencies:

```bash
npm ci
npm test
```

Tests cover RPG navigation, cancellation, stale results, shutdown during saves,
small-screen layouts, catch/escape outcomes, pagination, cached requests, malformed responses,
timeouts, save recovery and locking, legacy-save discovery, plain CLI output,
local web API boundaries, search, selection, empty states, and refresh failures.
[CI](https://github.com/yaqyn/clibble-terminal-rpg/actions/workflows/ci.yml) runs native Go checks
on Linux, macOS, and Windows. See [release preparation](docs/RELEASE.md) for packaging.

## <img src="readme-assets/icons/security.svg" width="24" height="24" alt=""> Security and limitations

- This is an independent, single-user nature adventure. There are no accounts, sync service, battle simulation, or automatic backups.
- The companion binds to loopback only and rejects foreign hosts/origins. Its content policy allows bundled assets only; it has no write endpoint or third-party tracking.
- Saves are validated, bounded to 4 MiB, flushed before replacement, and protected by cooperating-writer locks. This does not replace backups or a database durability guarantee.
- Catch limits discourage normal spamming, including across restarts. They use the local clock and save file; the device owner can edit both. This is a single-player game, not a tamper-proof competitive service.
- Wildlings can escape. All exploration, catching, and saved records work offline.

## <img src="readme-assets/icons/license.svg" width="24" height="24" alt=""> License

[MIT](LICENSE). Dependency and font licenses are included in
[third-party notices](THIRD_PARTY_NOTICES.md). The current trails and wildlings are
created for Clibble; legacy imported records remain readable.

> Based on Boot.dev's [Build a Pokedex in Go](https://www.boot.dev/courses/build-pokedex-cli-golang),
> extended with persistent saves, scriptable commands, a local collection companion,
> and the Clibble identity.
