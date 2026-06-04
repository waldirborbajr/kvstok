## KVStok

---

<p align="center">
  <img width="256" height="256" src="./assets/kvstok-logo.png" />
</p>

[![Code Scanning](https://github.com/waldirborbajr/kvstok/actions/workflows/codeql.yml/badge.svg)](https://github.com/waldirborbajr/kvstok/actions/workflows/codeql.yml)
[![Dependabot Updates](https://github.com/waldirborbajr/kvstok/actions/workflows/dependabot/dependabot-updates/badge.svg)](https://github.com/waldirborbajr/kvstok/actions/workflows/dependabot/dependabot-updates)
[![Dependency Graph](https://github.com/waldirborbajr/kvstok/actions/workflows/dependabot/update-graph/badge.svg)](https://github.com/waldirborbajr/kvstok/actions/workflows/dependabot/update-graph)
[![Lint](https://github.com/waldirborbajr/kvstok/actions/workflows/lint.yaml/badge.svg)](https://github.com/waldirborbajr/kvstok/actions/workflows/lint.yaml)
[![Test, Build and Publish](https://github.com/waldirborbajr/kvstok/actions/workflows/ci-cd.yaml/badge.svg)](https://github.com/waldirborbajr/kvstok/actions/workflows/ci-cd.yaml)
[![Typo Check](https://github.com/waldirborbajr/kvstok/actions/workflows/typo-check.yaml/badge.svg)](https://github.com/waldirborbajr/kvstok/actions/workflows/typo-check.yaml)
[![Update CONTRIBUTORS file](https://github.com/waldirborbajr/kvstok/actions/workflows/update_contributors.yml/badge.svg)](https://github.com/waldirborbajr/kvstok/actions/workflows/update_contributors.yml)
[![CI](https://github.com/waldirborbajr/kvstok/actions/workflows/ci.yaml/badge.svg)](https://github.com/waldirborbajr/kvstok/actions/workflows/ci.yaml)

---

<p align="center">
  <img width="256" height="256" src="./assets/kvstok-logo.png" />
</p>

## About

KVStoK is a secure, local key-value store for managing secrets, credentials, and configuration variables from the command line. Store sensitive data in one place without exporting to environment variables or cluttering your shell history.

### Perfect for:
- **DevOps Engineers** - Manage multiple credentials securely
- **Content Creators** - Work safely on stream without exposing secrets
- **Developers** - Quick access to API keys, tokens, and config values
- **System Administrators** - Centralized credential management

### Features

- 🔒 **Encrypted Storage** - All values encrypted with a master password
- 💾 **Local Only** - No cloud storage, complete privacy and security
- ⏱️ **TTL Support** - Set expiration times for temporary credentials
- 🏷️ **Tagged Keys** - Organize and search secrets by tags
- 📋 **Clipboard Ready** - Copy values to clipboard with one command
- 🔄 **Portable** - Works seamlessly across multiple terminal sessions
- 🌐 **Unicode Support** - Store any character set

## Quick Start

### Initialize KVStoK

```sh
kvstok init
```

### Store a Secret

```sh
kvstok add mykey mysecretvalue
# or with alias
kvstok a mykey mysecretvalue
```

### Retrieve a Secret

```sh
kvstok get mykey
# or with alias
kvstok g mykey
```

### Copy to Clipboard

```sh
kvstok copy mykey
# or with alias
kvstok cp mykey
```

### List All Secrets

```sh
kvstok list
# or with alias
kvstok l
```

### Delete a Secret

```sh
kvstok del mykey
# or with alias
kvstok d mykey
```

### Temporary Secret (TTL)

```sh
# Expire in 10 minutes
kvstok ttl mytempkey mytempvalue 10
```

### Search for Keys

```sh
kvstok search pattern
# or with alias
kvstok s pattern
```

## Usage Examples

### Direct Terminal Commands

```sh
# Store credentials
$ kvstok add dockerpwd MySecurePassword123

# Retrieve and use in commands
$ docker login -u myuser -p $(kvstok get dockerpwd)

# Copy to clipboard for manual paste
$ kvstok copy dockerpwd
✅ Key 'dockerpwd' copied to the clipboard!

# Use in scripts
#!/bin/bash
USER=$(kvstok get ghuser)
TOKEN=$(kvstok get ghtoken)
curl -u $USER:$TOKEN https://api.github.com/user

# Unicode support
$ kvstok add emoji "🔒"
$ kvstok get emoji
🔒
```

### Master Password Management

```sh
# Initialize with master password
$ kvstok init

# Check status
$ kvstok master status

# Change password
$ kvstok master change

# Provide password via CLI or environment
$ kvstok --master YOURPASSWORD get mykey
$ export KVSTOK_MASTER_PASSWORD=YOURPASSWORD
$ kvstok get mykey
```

### Export & Import

```sh
# Export all keys to JSON
$ kvstok export

# Import from JSON backup
$ kvstok import kvstok.json
```

## Installation

### macOS

```sh
# Download the latest release
# Extract and move to bin
tar xzvf kvstok_x.x.x_darwin_XXXX.tar.gz
mv kvstok ~/bin/

# Ensure ~/bin is in PATH
export PATH="$HOME/bin:$PATH"
```

### Linux

```sh
# Download the latest release
tar xzvf kvstok_x.x.x_linux_XXXX.tar.gz
mv kvstok ~/bin/

# Ensure ~/bin is in PATH
export PATH="$HOME/bin:$PATH"
```

### Windows

```powershell
# Download kvstok_x.x.x_windows_XXXX.zip
# Extract to your preferred location
# Add to your PATH
```

Or install via package managers (when available).

## All Commands

## All Commands

| Command | Aliases | Description |
|---------|---------|-------------|
| `add` | `addkv`, `a` | Add or update a secret value |
| `copy` | `cp` | Copy a secret value to clipboard |
| `del` | `delkv`, `d` | Delete a secret |
| `get` | `getkv`, `g` | Retrieve a secret value |
| `list` | `listkv`, `l` | List all secrets |
| `export` | `exportkv`, `e` | Export secrets to JSON file |
| `import` | `importkv`, `i` | Import secrets from JSON file |
| `ttl` | `ttladdkv`, `t` | Create a secret with expiration time |
| `search` | `s` | Search for secrets by pattern |
| `tag` | - | Manage secret tags |
| `env` | - | Export secrets as environment variables |
| `master` | - | Manage master password |
| `init` | - | Initialize KVStoK |

## Development

Built with:
- [Go](https://go.dev/) - Programming language
- [NutsDB](https://github.com/nutsdb/nutsdb) - Embedded database
- [Cobra](https://cobra.dev/) - CLI framework

## Contributing

## Contributing

We welcome contributions! Please refer to [CONTRIBUTING.md](./CONTRIBUTING.md) for guidelines on how to contribute to this project.

## Security

For security issues, please refer to [SECURITY.md](./SECURITY.md).

## License

KVStoK is licensed under the [Apache License 2.0](./LICENSE).

SPDX-License-Identifier: Apache-2.0

Copyright 2022-2026 Waldir Borba Junior (<wborbajr@gmail.com>)

## Star History

[![Star History Chart](https://api.star-history.com/svg?repos=waldirborbajr/kvstok&type=Date)](https://star-history.com/#waldirborbajr/kvstok&Date)
