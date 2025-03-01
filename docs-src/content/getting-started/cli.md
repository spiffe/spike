+++
# //    \\ SPIKE: Secure your secrets with SPIFFE.
# //  \\\\\ Copyright 2024-present SPIKE contributors.
# // \\\\\\\ SPDX-License-Identifier: Apache-2.

title = "SPIKE CLI"
weight = 5
sort_by = "weight"
+++

## SPIKE Command Line Interface

> **SPIKE in Action**
>
> To watch **SPIKE** CLI in action, [you can check out **SPIKE** presentations
and demo recordings][demos].

**SPIKE** uses **SPIKE Pilot** (*the command line tool*) to interact with 
**SPIKE Nexus** (*the secrets store*).

Since **SPIKE** is in active development, its command line interface changes
a lot. The best way to learn about it will be to use its `--help` flag.

Here is how the interface looks like. Note that what you see might be 
different based on the version you use.

```bash
spike (main)$ spike
>> Secure your secrets with SPIFFE: https://spike.ist/ #

Usage:
  spike [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  operator    Manage admin operations
  policy      Manage policies
  secret      Manage secrets

Flags:
  -h, --help   help for spike

Use "spike [command] --help" for more information about a command.
```

And here is how we can get help about a certain using the `--help` flag:

```bash
spike (main)$ spike secret --help
Manage secrets

Usage:
  spike secret [command]

Available Commands:
  delete      Delete secrets at the specified path
  get         Get secrets from the specified path
  list        List all secret paths
  metadata    Manage secret metadata
  put         Put secrets at the specified path
  undelete    Undelete secrets at the specified path

Flags:
  -h, --help   help for secret

Use "spike secret [command] --help" for more information about a command.
```

Let's dig in further:

```bash
spike (main)$ spike secret put --help
Put secrets at the specified path

Usage:
  spike secret put <path> <key=value>... [flags]

Flags:
  -h, --help   help for put
```

Okay, that explains a lot. Let's try the command:

```bash
spike (main)$ spike secret put /tenants/acme/db/creds pass=SPIKERocks
OK
```

Now let's try to read this secret:

```bash
spike (main)$ spike secret get --help
Get secrets from the specified path

Usage:
  spike secret get <path> [flags]

Flags:
  -h, --help          help for get
  -v, --version int   Specific version to retrieve
```

Now that we know how to use the `spike secret get` command, let's try it.

```bash
spike (main)$ spike secret get /tenants/acme/db/creds
pass: SPIKERocks
```

That's about it. You can use other **SPIKE** commands similarly.

[demos]: @/community/presentations.md