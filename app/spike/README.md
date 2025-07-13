```text
  \\ 
 \\\\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
\\\\\\
```

## SPIKE Pilot

**SPIKE Pilot** is the command-line interface for the **SPIKE** system.

It's a binary named `spike`.

It's helpful to define an alias to `spike` for ease-of-use:

```bash
# ~/.bashrc

# Define an alias to where your `spike` binary is:
alias spike=$HOME/WORKSPACE/spike-git-repo/spike
```

## Getting Help

Simply typing `spike` will show a summary of available commands.

```text
[~/Desktop/WORKSPACE/spike]$ spike
Usage: spike <command> [args...]
Commands:
  init
  put <path> <key=value>...
  get <path> [-version=<n>]
  delete <path> [-versions=<n1,n2,...>]
  undelete <path> [-versions=<n1,n2,...>]
  list
```

Note that the CLI is a constant work in progress, so what you see above might be 
slightly different from the version that you are using.

For additional help, you can check the official documentation.

Also note that the official documentation is a work in progress too.

TODO:// add links to the official docs when it's ready.
