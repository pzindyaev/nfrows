# nfrows

A terminal UI for managing nftables — browse, create, edit, and delete tables, chains, rules, sets, maps, flowtables, and counters interactively.

Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea), [Bubbles](https://github.com/charmbracelet/bubbles), and [Lip Gloss](https://github.com/charmbracelet/lipgloss).

---

## Requirements

- Go 1.24.2+
- `nft` (nftables) installed and on `$PATH`
- Root privileges or `CAP_NET_ADMIN` capability (nftables requires it)

---

## Build

```sh
go build -o nfrows .
```

Or install directly:

```sh
go install github.com/pzindyaev/nfrows@latest
```

---

## Run

```sh
sudo ./nfrows
```

---

## Interface

### Layout

```
┌─ header ───────────────────────────────────────────┐
│ nfrows                             nftables TUI    │
├─ breadcrumb ───────────────────────────────────────┤
│ tables › inet:filter                               │
├─ tabs (table detail only) ─────────────────────────┤
│  Chains   Sets   Maps   Flowtables   Counters      │
├─ content ──────────────────────────────────────────┤
│ ┌──────────────────────────────────────────────┐   │
│ │ Handle │ Name      │ Type   │ Hook  │ Policy │   │
│ │      1 │ input     │ filter │ input │ drop   │   │
│ │      2 │ forward   │ filter │ fwd   │ drop   │   │
│ └──────────────────────────────────────────────┘   │
├─ flash message ────────────────────────────────────┤
│ Done.                                              │
├─ status bar ───────────────────────────────────────┤
│ r refresh  a/o add  j/k down/up  gg/G top/bottom   │
└────────────────────────────────────────────────────┘
```

### Views

| View | Description |
|---|---|
| **Tables** | Top-level list of all nftables tables |
| **Table detail** | Tabbed view of everything inside a table |
| **Chain rules** | All rules inside a chain, with human-readable text |

---

## Keybindings

### Navigation

| Key | Action |
|---|---|
| `j` / `↓` | Move down |
| `k` / `↑` | Move up |
| `gg` | Jump to first row |
| `G` | Jump to last row |
| `ctrl+u` | Half-page up |
| `ctrl+d` | Half-page down |
| `h` / `←` | Previous tab (table detail) |
| `l` / `→` | Next tab (table detail) |
| `tab` | Next tab |
| `shift+tab` | Previous tab |
| `enter` | Open selected item |
| `esc` | Go back / close modal |

### Actions

| Key | Action |
|---|---|
| `a` / `o` | Add new item |
| `d` | Delete selected item |
| `e` / `i` | Edit selected rule |
| `f` | Flush (remove all rules from table or chain) |
| `r` | Refresh ruleset from kernel |
| `q` / `ctrl+c` | Quit (from tables view) |

### Modals

| Key | Action |
|---|---|
| `enter` | Confirm form |
| `esc` | Cancel |
| `tab` / `↓` | Next field |
| `shift+tab` / `↑` | Previous field |
| `h` / `l` | Cycle option selector left/right |
| `y` | Confirm deletion |
| `n` / `esc` | Cancel deletion |

---

## CRUD reference

### Tables

| Action | How |
|---|---|
| List | Shown on startup |
| Add | `a` → choose family + name |
| Delete | `d` → confirm |
| Flush | `f` → confirm (removes all rules, keeps structure) |
| Open | `enter` |

### Chains

Navigate into a table, then the **Chains** tab is shown by default.

| Action | How |
|---|---|
| Add | `a` → name, type (`filter`/`nat`/`route`), hook, policy, priority |
| Delete | `d` → confirm |
| Flush | `f` → confirm |
| Open (view rules) | `enter` |

Chains are created as base chains (with a hook) via the interactive form. To create a regular chain (no hook), leave Type/Hook as their defaults and nft will treat it as one — alternatively add it via a custom `nft` command and refresh with `r`.

### Rules

Navigate into a table → chain.

| Action | How |
|---|---|
| Add | `a` / `o` → type raw rule text, e.g. `ip protocol tcp accept` |
| Edit | `e` / `i` → pre-filled with current rule text |
| Delete | `d` → confirm |

Rules are displayed as human-readable text (fetched via `nft -a list chain`), not as JSON handles.

### Sets

Navigate into a table → **Sets** tab.

| Action | How |
|---|---|
| Add | `a` → name, element type (e.g. `ipv4_addr`), optional flags (csv, e.g. `interval,timeout`) |
| Delete | `d` → confirm |

### Maps

Navigate into a table → **Maps** tab.

| Action | How |
|---|---|
| Add | `a` → name, key type, value type |
| Delete | `d` → confirm |

### Flowtables

Navigate into a table → **Flowtables** tab.

| Action | How |
|---|---|
| Add | `a` → name, device list (csv, e.g. `eth0,eth1`) |
| Delete | `d` → confirm |

### Counters / Quotas / Limits

Navigate into a table → **Counters** tab. Named counters, quotas, and limits all appear here.

| Action | How |
|---|---|
| Add counter | `a` → name |
| Delete | `d` → confirm |

---

## Project structure

```
nfrows/
├── main.go                       # Entry point
└── internal/
    ├── nft/
    │   ├── types.go              # Ruleset data types (JSON-mapped)
    │   └── client.go             # nft CLI wrapper — full CRUD + rule text parsing
    └── ui/
        ├── app.go                # Root Bubble Tea model and navigation state machine
        ├── table_widget.go       # Scrollable table renderer with keyboard selection
        ├── form.go               # Multi-field modal form (text inputs + option selectors)
        ├── confirm.go            # Yes/no confirmation dialog
        ├── styles.go             # Lip Gloss colour palette and component styles
        ├── keys.go               # Key binding definitions
        ├── msgs.go               # Bubble Tea message types
        └── cmds.go               # Async commands (load ruleset, mutations)
```

---

## How it works

nfrows shells out to the `nft` binary for all operations:

- **Reading** — `nft -j list ruleset` returns the full ruleset as JSON, which is parsed into typed Go structs and organised per-table.
- **Rule display** — `nft -a list chain <family> <table> <chain>` returns human-readable rule text annotated with `# handle N` comments, which are parsed into a handle→text map.
- **Writing** — individual `nft add/delete/replace/flush` commands are issued for each mutation. The full ruleset is reloaded from the kernel after every successful operation.

No direct netlink calls are made; the `nft` binary is the only dependency at runtime.

---

## License

MIT
