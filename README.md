## KVStok

<p>
<img alt="KVStoK Logo" src="https://github.com/waldirborbajr/kvstok/blob/main/assets/logo.png" width="120", heigth="120"/>
</p>

KVStoK is an open source software built-in with the main aim of being a personal [KEY][VALUE] store, to keep system variables as parameters or passwords or anything else stored in a single place.

### How to use

```sh

# Store a value
$ kvstok addkv containerpwd 123SecretPWD

# List all stored values if informed json will generate a json file
$ kvstok listkv [json]

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

### Installation

Make sure that `$HOME/.local/bin` it is in your library path.


```sh

$ cd ~/.local/bin

$ curl -LO https://github.com/WaldirBorbaJR/kvstok/-/blob/main/bin/kvstok

```

## Contributing

If you're interested in contributing to KVStoK see [CONTRIBUTING.md](./CONTRIBUTING.md).

OR

Just follow steps below if you are an expert.

1. Fork it
2. Clone v1: `git clone -b v1 https://github.com/WaldirBorbaJR/kvstok`
3. Create your feature branch: `git checkout -b v1-feature`
4. Make changes and add them: `git add .`
5. Commit: `git commit -m "add some feature"`
6. Push: `git push origin v1-feature`
7. Pull request

## License

[Apache](https://github.com/WaldirBorbaJR/kvstok/-/blob/main/LICENSE)

## Legal

Copyright 2022 Waldir Borba Junior (<mailto:wborbajr@gmail.com>)
SPDX-License-Identifier: Apache-2.0

## TODO

**Note: This file is no longer being updated.**

The todo file does not represent ALL of the missing features. This file just shows the features which I noticed were missing and I have to implement.

[TODO](./TODO.md)

## Technology

| <img src="assets/logo.png" alt="logo" width="45" hight="45"/> | <img src="assets/gopher.png" alt="gopher" width="45" hight="45"/> | <img src="assets/nutsdb.png" alt="nutsdb" width="45" hight="45"/> | <img src="assets/cobra.png" alt="cobra" width="45" hight="45"/> |


[KVStoK]|[GO](https://go.dev/)|[NutsDB](https://github.com/nutsdb/nutsdb)|[Cobra](https://cobra.dev/)|
