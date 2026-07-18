# portping

`portping` is a small, concurrent TCP reachability scanner. It probes
`host:port` targets using a bounded worker pool, measures connect latency, and
reports the results as an aligned table or a JSON array.

It is handy for quick network checks: confirming a firewall change, verifying a
service is listening across a subnet, or sweeping a range of ports on a host.

## Install

```sh
go install github.com/cybercapybara/portping@latest
```

Or build from a checkout:

```sh
go build -o portping .
```

Requires Go 1.17 or newer. No third-party dependencies.

## Usage

```
usage: portping [flags] target [target ...]

A target is HOST:PORTS, where HOST may be a hostname, IP, or IPv4 CIDR
and PORTS may be a list and/or range, e.g. 10.0.0.0/28:22,80,8000-8010
```

### Targets

A target is written as `HOST:PORTS`:

- **HOST** — a hostname (`example.com`), an IP literal (`10.0.0.5`,
  `[::1]`), or an IPv4 CIDR block (`10.0.0.0/28`).
- **PORTS** — a comma-separated list of ports and/or ranges, such as
  `22`, `22,80,443`, or `8000-8010`.

CIDR expansion excludes the network and broadcast addresses for prefixes of
`/30` and shorter. A `/31` yields both addresses (RFC 3021 point-to-point) and a
`/32` yields the single host.

Targets may be passed as arguments, read from a file with `--file`, or piped on
stdin (`--file -`, or simply no arguments). In files and on stdin, blank lines
and lines beginning with `#` are ignored, and whitespace-separated targets on a
line are all read.

### Flags

| Flag | Default | Description |
| --- | --- | --- |
| `--timeout` | `2s` | Per-dial timeout. |
| `--concurrency` | `64` | Number of concurrent workers. |
| `--count` | `1` | Dial attempts before a target is declared unreachable. |
| `--json` | `false` | Emit results as a JSON array. |
| `--open-only` | `false` | Only report reachable targets. |
| `--file` | | Read targets from a file (`-` for stdin). |
| `--fail-on-unreachable` | `true` | Exit non-zero if any target is unreachable. |

### Exit codes

- `0` — all probed targets were reachable (or `--fail-on-unreachable=false`).
- `1` — one or more targets were unreachable.
- `2` — invalid arguments or targets.

## Examples

Scan a couple of ports on localhost:

```sh
$ portping 127.0.0.1:22,80,443
TARGET          STATE   LATENCY  ATTEMPTS  DETAIL
127.0.0.1:22    open    0.3ms    1
127.0.0.1:80    closed  -        1         dial tcp 127.0.0.1:80: connect: connection refused
127.0.0.1:443   closed  -        1         dial tcp 127.0.0.1:443: connect: connection refused
```

Sweep an SSH port across a /28 with a tight timeout, showing only what answers:

```sh
$ portping --timeout 300ms --open-only 10.0.0.0/28:22
```

Emit JSON for further processing:

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
