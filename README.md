## KVStok

[![📬 Typo](https://github.com/waldirborbajr/kvstok/actions/workflows/1.typo-check.yaml/badge.svg)](https://github.com/waldirborbajr/kvstok/actions/workflows/1.typo-check.yaml)
[![🙏🏻 Li](https://github.com/waldirborbajr/kvstok/actions/workflows/2.lint.yaml/badge.svg)](https://github.com/waldirborbajr/kvstok/actions/workflows/2.lint.yaml)
[![🚨 CodeQL](https://github.com/waldirborbajr/kvstok/actions/workflows/3.codeql.yaml/badge.svg)](https://github.com/waldirborbajr/kvstok/actions/workflows/3.codeql.yaml)
[![🐕 Review DOG](https://github.com/waldirborbajr/kvstok/actions/workflows/4.reviewdog.yaml/badge.svg)](https://github.com/waldirborbajr/kvstok/actions/workflows/4.reviewdog.yaml)
[![🔨 Build and Test](https://github.com/waldirborbajr/kvstok/actions/workflows/5.build-test.yaml/badge.svg)](https://github.com/waldirborbajr/kvstok/actions/workflows/5.build-test.yaml)
[![🎉 Release](https://github.com/waldirborbajr/kvstok/actions/workflows/goreleaser.yaml/badge.svg)](https://github.com/waldirborbajr/kvstok/actions/workflows/goreleaser.yaml)

<p align="center"> <img alt="KVStoK Logo" src="./assets/logo.png" width="120", height="120"/> </p>

KVStoK is an open source software built-in with the main aim of being a personal [KEY][VALUE] store, to keep system variables as parameters or passwords or anything else stored in a single place.

### How to use

#### Typing `full` command name

```sh

# Store a value
$ kvstok addkv containerpwd 123SecretPWD

# List all stored values if informed json will generate a json file
$ kvstok lstkv

# Grab a value stored into a key
$ kvstok getkv containerpwd
123SecretPWD

# Remove a stored key/value from database storage
$ kvstok delkv containerpwd

# Unicode params are allowed too
$ kvstok addkv someParam 喵
$ kvstok getkv someParam
喵
```

### Integrated to shell script

```sh
#!/bin/bash

dosomething = $(kvstok getkv someParam)
echo ${dosomething}
..
.
```

#### Typing `alias` of command name, first letter

```sh

# Store a value
$ kvstok a containerpwd 123SecretPWD

# List all stored values if informed json will generate a json file
$ kvstok l

# Grab a value stored into a key
$ kvstok g containerpwd
123SecretPWD

# Remove a stored key/value from database storage
$ kvstok d containerpwd

# Unicode params are allowed too
$ kvstok a someParam 喵
$ kvstok g someParam
喵
```

### Integrated to shell script

```sh
#!/bin/bash

dosomething = $(kvstok g someParam)
echo ${dosomething}
..
.
```

### Install

### Download binary according to you OS version at

#### macOS

1. Download **kvstok_x.x.x_darwin_XXXX.tar.gz**
2. Extract: `tar xzvf kvstok_x.x.x_darwin_XXXX.tar.gz`
3. Move to `mv kvstok ~/.local/bin`
4. Make sure that `$HOME/.local/bin` it is in your library path.
5. Run `kvstok`

#### Linux

1. Download **kvstok_x.x.x_linux_XXXX.tar.gz**
2. Extract: `tar xzvf kvstok_x.x.x_linux_XXXX.tar.gz`
3. Move to `mv kvstok ~/.local/bin`
4. Make sure that `$HOME/.local/bin` it is in your library path.
5. Run `kvstok`

## How can I contribute?

Kindly refer to [CONTRIBUTING.md](./CONTRIBUTING.md) file to learn how to contribute!

And that's it!
Follow these steps to make your very first pull request.

## License

[Apache](https://github.com/WaldirBorbaJR/kvstok/-/blob/main/LICENSE)

## Legal

Copyright 2022 Waldir Borba Junior (<mailto:wborbajr@gmail.com>)
SPDX-License-Identifier: Apache-2.0

## TODO

**Note: This file is no longer being updated.**

The todo file does not represent ALL of the missing features. This file just shows the features which I noticed were missing and I have to implement.

For a list of all closed TODO: `is:issue is:closed TODO`

For a list of all open TODO: `is:issue is:open TODO`

## Technology

| <img src="assets/logo.png" alt="logo" width="45" height="45"/> | <img src="assets/gopher.png" alt="gopher" width="45" height="45"/> | <img src="assets/nutsdb.png" alt="nutsdb" width="45" height="45"/> | <img src="assets/cobra.png" alt="cobra" width="45" height="45"/> |


[KVStoK]|[GO](https://go.dev/)|[NutsDB](https://github.com/nutsdb/nutsdb)|[Cobra](https://cobra.dev/)|
