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
