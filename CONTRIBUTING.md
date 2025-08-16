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

## Audit Your Code Before You Submit

Before submitting a pull request run `make audit` and ensure that there are no
issues. Running `make audit` will find issues with the code that would trigger
CI failures and prevent it from being merged.

## How To Run Tests

`make audit` already runs the tests for you; however, running the tests
separately is still useful.

Before merging your changes, make sure all tests pass.

To run the unit tests locally, run `go test ./...` on the project root.
