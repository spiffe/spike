![SPIKE](assets/spike-banner-lg.png)

## Welcome

Thank you for your interest in contributing to **SPIKE** 🤘.

We appreciate any help, be it in the form of code, documentation, design,
or even bug reports and feature requests.

When contributing to this repository, please first discuss the change you wish
to make via an issue, email, or any other method before making a change.
This way, we can avoid misunderstandings and wasted effort.

One great way to initiate such discussions is asking a question 
[SPIFFE Slack Community][slack].

[slack]: https://slack.spiffe.io/ "Join SPIFFE on Slack"

Please note that [we have a code of conduct](CODE_OF_CONDUCT.md). We expect all
contributors to adhere to it in all interactions with the project.

Also, make sure you read, understand and accept
[The Developer Certificate of Origin Contribution Guide](CONTRIBUTING_DCO.md)
as it is a requirement to contribute to this project and contains more details
about the contribution process.

## Audit and Test Your Code Before You Submit

Before submitting a pull request, run the following commands and make sure
that there are no issues:

## Before Submitting a Pull Request

* `make build`: Ensure that the code builds first.
* `make test`: Run the tests.
* `make audit`: Run security audits and linters.
* `make start`: Start SPIKE components and run smoke tests. Ensure you see
  the following output to confirm all components pass:
  ```txt
  > Everything is set up.
  > You can now experiment with SPIKE.
  ```
  Press `Ctrl+C` to stop the components after verification.

If all of the above pass, you're ready to submit a pull request.

## Optional: AI-Assisted Development Tooling

SPIKE ships configuration for a few AI-assisted development tools. They
are **entirely optional**. You do not need any of them to build, test,
or contribute to SPIKE, and the standard `make` workflow above is the
only supported path. Use them only if they fit your workflow.

### ctx (persistent context)

`ctx` gives AI agents persistent, project-scoped memory under
`.context/`. If you work with Claude Code or a similar agent and want it
to carry context across sessions, install `ctx` and run
`ctx system bootstrap` from the project root. See the getting-started
guide at <https://ctx.ist/home/getting-started/>.

When `ctx` is not installed, nothing changes: `.context/` is left to
your own tooling and the `make` targets in `makefiles/Ctx.mk` simply go
unused.

### GitNexus (code intelligence)

GitNexus indexes the codebase for impact analysis, symbol navigation,
and safe refactoring, exposed to agents via MCP tools. To build the
index, run `npx gitnexus analyze` from the project root. Usage details
live in [GITNEXUS.md](GITNEXUS.md).

If GitNexus is not installed, ignore [GITNEXUS.md](GITNEXUS.md); the
tools and generated skills it references will not be present.
