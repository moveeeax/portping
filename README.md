# portping

`portping` is a small, concurrent TCP reachability scanner. It probes
`host:port` targets using a bounded worker pool, measures connect latency, and
reports the results as an aligned table or a JSON array.

It is handy for quick network checks: confirming a firewall change, verifying a
service is listening across a subnet, or sweeping a range of ports on a host.

## Install

```sh
go install github.com/moveeeax/portping@latest
```

Or build from a checkout:

```sh
go build -o portping .
```

Requires Go 1.26 or newer. No third-party dependencies.

## Usage

```
usage: portping [flags] target [target ...]

A target is HOST:PORTS, where HOST may be a hostname, IP, or IPv4 CIDR
and PORTS may be a list and/or range, e.g. 10.0.0.0/28:22,80,8000-8010
Append =open or =closed to assert the expected state, e.g. 10.0.0.5:3306=closed
```

### Targets

A target is written as `HOST:PORTS`:

- **HOST** — a hostname (`example.com`), an IP literal (`10.0.0.5`,
  `[::1]`), or an IPv4 CIDR block (`10.0.0.0/28`).
- **PORTS** — a comma-separated list of ports and/or ranges, such as
  `22`, `22,80,443`, or `8000-8010`.

A target may end with `=open` or `=closed` to state what it is expected to be,
for example `10.0.0.5:22=open 10.0.0.5:3306=closed`. The suffix applies to every
pair the target expands to (`10.0.0.0/28:3306=closed` asserts the port is closed
on all 14 hosts). Without a suffix a target is expected to be open. Any other
value after `=` is rejected with exit code `2`.

If the same `host:port` pair is given more than once, an explicit suffix wins
over the default, so `10.0.0.0/28:22 10.0.0.5:22=closed` expects SSH open
everywhere except on `10.0.0.5`. Giving the same pair both `=open` and
`=closed` is an error (exit code `2`).

CIDR expansion excludes the network and broadcast addresses for prefixes of
`/30` and shorter. A `/31` yields both addresses (RFC 3021 point-to-point) and a
`/32` yields the single host.

Expansion is bounded so that a mistyped prefix cannot exhaust memory before the
first dial. A single CIDR may cover at most 65536 addresses (a `/16`), and one
invocation may expand to at most 1048576 distinct `host:port` pairs in total.
Anything larger is rejected with exit code `2`; split the sweep into narrower
targets.

Targets may be passed as arguments, read from a file with `--file`, or piped on
stdin (`--file -`, or simply no arguments). In files and on stdin, blank lines
and lines beginning with `#` are ignored, and whitespace-separated targets on a
line are all read.

### Flags

| Flag | Default | Description |
| --- | --- | --- |
| `--timeout` | `2s` | Per-dial timeout. Must be greater than zero. |
| `--concurrency` | `64` | Number of concurrent workers. |
| `--count` | `1` | Dial attempts before a target is declared unreachable. |
| `--json` | `false` | Emit results as a JSON array. |
| `--open-only` | `false` | Only report reachable targets. |
| `--file` | | Read targets from a file (`-` for stdin). |
| `--fail-on-unreachable` | `true` | Exit non-zero if any target without an explicit `=open`/`=closed` is unreachable. |

A zero or negative `--timeout` is rejected rather than accepted as "no
deadline": without a positive per-dial timeout, a single target that never
responds (a firewall silently dropping the SYN, for example) would hang the
whole scan indefinitely instead of being reported as closed.

`--fail-on-unreachable=false` only waives targets without a suffix. A target
with an explicit `=open` or `=closed` is an assertion and fails the run when it
does not match, whatever this flag says.

### Exit codes

- `0` — every target is in its expected state (or the only mismatches are
  unreachable targets without a suffix and `--fail-on-unreachable=false` is
  set). Also returned for `-h`/`--help`.
- `1` — one or more targets are not in their expected state: an unreachable
  target expected open, or a reachable target expected closed.
- `2` — invalid arguments or targets.

## Examples

Scan a couple of ports on localhost:

```sh
$ portping 127.0.0.1:22,80,443
TARGET         STATE   EXPECT  MATCH  LATENCY  ATTEMPTS  DETAIL
127.0.0.1:22   open    open    yes    0.3ms    1
127.0.0.1:80   closed  open    no     -        1         dial tcp 127.0.0.1:80: connect: connection refused
127.0.0.1:443  closed  open    no     -        1         dial tcp 127.0.0.1:443: connect: connection refused
```

Check a firewall change in CI: SSH must answer, MySQL must not.

```sh
$ portping --timeout 500ms 10.0.0.5:22=open 10.0.0.5:3306=closed
TARGET         STATE   EXPECT  MATCH  LATENCY  ATTEMPTS  DETAIL
10.0.0.5:22    open    open    yes    1.2ms    1
10.0.0.5:3306  closed  closed  yes    -        1         dial tcp 10.0.0.5:3306: i/o timeout
```

The same assertions work in a `--file` list, one per line.

Sweep an SSH port across a /28 with a tight timeout, showing only what answers:

```sh
$ portping --timeout 300ms --open-only 10.0.0.0/28:22
```

Emit JSON for further processing. Each object carries `expected` (`"open"` or
`"closed"`) and `ok` (whether the observed state matches it):

```sh
$ portping --json 10.0.0.0/30:22,80
```

Read targets from stdin:

```sh
$ printf '10.0.0.1:22\n10.0.0.2:80\n' | portping --file -
```

## Layout

- `internal/scan` — pure target parsing, CIDR/port expansion, result
  aggregation, and the worker pool.
- `internal/probe` — TCP dialing behind an injectable `Dialer` interface for
  testability.
- `main.go` — the command-line interface.

## License

MIT. See [LICENSE](LICENSE).
