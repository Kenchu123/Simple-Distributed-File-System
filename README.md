# Simple Distributed File System

UIUC CS425, Distributed Systems: Fall 2023 Machine Programming 3

## Description

See [MP Document](./docs/MP3.CS425.FA23.pdf)

## Installation

### Prerequisites

- Go 1.20

### Build

```bash
make build
```

## Usage

### Serve

`serve` command start the command server and wait for commands from clients.

```bash
./bin/sdfs serve [flags]

Flags:
  -c, --config string   path to config file (default ".sdfs/config.yml")
  -h, --help            help for serve
  -l, --log string      path to log file (default "logs/sdfs.log")
```

### Join

`join` command tells the machine to join the group.

```bash
./bin/sdfs join [flags]

Flags:
  -h, --help   help for join
  -c, --config string          path to config file (default ".sdfs/config.yml")
  -m, --machine-regex string   regex for machines to join (e.g. "0[1-9]") (default ".*")
```

### Leave

`leave` command tells the machine to leave the group.

```bash
./bin/sdfs leave [flags]

Flags:
  -h, --help   help for leave
  -c, --config string          path to config file (default ".sdfs/config.yml")
  -m, --machine-regex string   regex for machines to join (e.g. "0[1-9]") (default ".*")
```

### Fail

`fail` command tells the machine's process to fail.

```bash
./bin/sdfs fail [flags]

Flags:
  -h, --help   help for fail
  -c, --config string          path to config file (default ".sdfs/config.yml")
  -m, --machine-regex string   regex for machines to join (e.g. "0[1-9]") (default ".*")
```

### List the Membership List

`list_mem` command lists the membership list.

```bash
./bin/sdfs list_mem [flags]

Flags:
  -h, --help   help for list_mem
  -c, --config string          path to config file (default ".sdfs/config.yml")
  -m, --machine-regex string   regex for machines to join (e.g. "0[1-9]") (default ".*")
```

### List Self's ID

`list_self` command lists self's ID.

```bash
./bin/sdfs list_self [flags]

Flags:
  -h, --help   help for list_self
  -c, --config string          path to config file (default ".sdfs/config.yml")
  -m, --machine-regex string   regex for machines to join (e.g. "0[1-9]") (default ".*")
```

### Enable/Disable Suspicion

`enable suspicion` command enables suspicion.
`disable suspicion` command disables suspicion.

```bash
./bin/sdfs enable suspicion [flags]
./bin/sdfs disable suspicion [flags]

Flags:
  -h, --help     help for suspicion
  -c, --config string          path to config file (default ".sdfs/config.yml")
  -m, --machine-regex string   regex for machines to join (e.g. "0[1-9]") (default ".*")
```

### Config

#### Set DropRate

`config set-droprate` command sets the drop rate.

```bash
./bin/sdfs config set-droprate [flags]

Flags:
  -d, --droprate float32   droprate
  -h, --help               help for droprate

Global Flags:
  -c, --config string          path to config file (default ".sdfs/config.yml")
  -m, --machine-regex string   regex for machines to join (e.g. "0[1-9]") (default ".*")
```

#### Set Verbose

`config set-verbose` command sets the verbose level.

```bash
./bin/sdfs config set-verbose [flags]

Flags:
  -h, --help      help for set-verbose
  -v, --verbose   enable or disable verbose

Global Flags:
  -c, --config string          path to config file (default ".sdfs/config.yml")
  -m, --machine-regex string   regex for machines to join (e.g. "0[1-9]") (default ".*")
```

## Development

### Prerequisites

- Docker
- docker compose

### Set Environment

```bash
# on one session
docker compose -f docker-compose.dev.yml up [-d] [--build]

# on another session
docker exec -it dev /bin/bash
docker exec -it dev-m[1-10] /bin/bash

$ go run main.go [command] [flags]
```

## Running on VMs

### Prerequisites

- `sdfs` binary in each VM home directory
- [tmux](https://github.com/tmux/tmux) installed

### Run Serve

Use tmux to run `sdfs serve` in background on each VM.

```bash
# run all sdfs process
./vm_run_all.sh
```

### Run Commands

```bash
# ssh to one machine
./sdfs [command] [flags]
```

### Kill

```bash
# kill all sdfs process
./vm_kill_all.sh
```

## Contributor

- [Che-Kuang Chu](https://github.com/Kenchu123)
- [Jhih-Wei Lin](https://github.com/williamlin0825)
