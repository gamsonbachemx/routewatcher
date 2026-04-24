# routewatcher

A lightweight CLI tool to monitor and diff routing table changes on Linux hosts.

---

## Installation

```bash
go install github.com/yourusername/routewatcher@latest
```

Or build from source:

```bash
git clone https://github.com/yourusername/routewatcher.git
cd routewatcher
go build -o routewatcher .
```

---

## Usage

Start watching for routing table changes:

```bash
routewatcher watch
```

Take a one-time snapshot and diff against a previous one:

```bash
routewatcher snapshot --output routes.json
routewatcher diff --from routes.json
```

Available flags:

| Flag | Description | Default |
|------|-------------|---------|
| `--interval` | Polling interval in seconds | `5` |
| `--output` | File path for snapshot output | stdout |
| `--format` | Output format (`text`, `json`) | `text` |

Example output:

```
[+] 192.168.10.0/24 via 10.0.0.1 dev eth0
[-] 10.8.0.0/16 via 10.0.0.254 dev eth1
```

---

## Requirements

- Linux (uses `/proc/net/route` and `ip route`)
- Go 1.21+

---

## License

MIT © 2024 yourusername