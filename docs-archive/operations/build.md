## Cross-Building SPIKE Binaries

We cross-build SPIKE on an ARM64 Mac Machine.

Here is what's needed for a cross-compile:

## Prerequisites

Installed required tools via [Homebrew](https://brew.sh).

```bash
brew install FiloSottile/musl-cross/musl-cross
brew install gcc
brew install x86_64-linux-gnu-binutils
brew install aarch64-elf-gcc
```

## Build

To cross-compile the binaries, run the following:

```bash
./hack/build-spike-cross-platform.sh
```

After the script runs to completion, you should get the following artifacts:

```txt
-rwxr-xr-x   1 volkan  staff  16262498 Nov 10 10:41 keeper-darwin-arm64
-rwxr-xr-x   1 volkan  staff  16644567 Nov 10 10:42 keeper-linux-amd64
-rwxr-xr-x   1 volkan  staff  16122001 Nov 10 10:41 keeper-linux-arm64
-rwxr-xr-x   1 volkan  staff  20632146 Nov 10 10:41 nexus-darwin-arm64
-rwxr-xr-x   1 volkan  staff  22916584 Nov 10 10:42 nexus-linux-amd64
-rwxr-xr-x   1 volkan  staff  21563848 Nov 10 10:42 nexus-linux-arm64
-rwxr-xr-x   1 volkan  staff  16982594 Nov 10 10:41 spike-darwin-arm64
-rwxr-xr-x   1 volkan  staff  17379008 Nov 10 10:42 spike-linux-amd64
-rwxr-xr-x   1 volkan  staff  16783196 Nov 10 10:42 spike-linux-arm64
```