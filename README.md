![SPIKE](docs-src/themes/zola_easydocs_theme/static/assets/spike-banner-lg.png)

## Secure Production Identity for Key Encryption (SPIKE)

**SPIKE** is a lightweight secrets store that uses [SPIFFE][spiffe]
as its identity control plane.

**SPIKE** protects your secrets and helps your ops, SREs, and sysadmins
`#sleepmore`.

For more information, [see the documentation][docs].

[docs]: https://spike.ist/
[spiffe]: https://spiffe.io/

## The Elevator Pitch

[**SPIKE**][spike] is a streamlined, highly-reliable secrets store that leverages 
[SPIFFE][spiffe] framework for strong, production-grade identity control. 

Built with simplicity and high availability in mind, SPIKE empowers ops teams, 
SREs, and sysadmins to protect sensitive data and `#sleepmore` by securing 
secrets across distributed environments.

Key components include:

* **SPIKE Nexus**: The heart of SPIKE, handling secret encryption, decryption, 
  and root key management.
* **SPIKE Keeper**: A redundancy mechanism that safely holds root keys in memory, 
  enabling fast recovery if Nexus fails.
* **SPIKE Pilot**: A secure CLI interface, translating commands into **mTLS** 
  API calls, reducing system vulnerability by containing all admin access.

With its minimal footprint and robust security, **SPIKE** provides peace of mind 
for your team and critical data resilience when it counts.

## 🚨 Alpha Release Notice 🚨

* **Project Status**: **Alpha**

This project is currently in the Alpha stage. It's functional and available for
experimentation, but it's **NOT** yet ready for production use: You may encounter
bugs, incomplete features, or breaking changes as the project evolves.

Use this project at your own risk if you're experimenting or contributing to its
development. For production-level stability, please wait for a more stable
release.

Please note that the [**SPIKE** documentation][docs] is a work in progress too.
It might be incomplete or inaccurate at times, and what the document
states may not fully reflect how the code or the product behaves.

Please 🐻 with us for now, and send your feedback to [team@spike.ist](mailto:team@spike.ist).

We will let you know through various channels when the project reaches adequate
maturity for public adoption.

## Getting Your Hands Dirty

[Check out the quickstart guide][quickstart] to start playing with the project.

[You can also read the documentation][spike] to learn more about **SPIKE**'s
architecture and design philosophy.

## A Note on Security

We take **SPIKE**'s security seriously. If you believe you have
found a vulnerability, please responsibily disclose it to 
[security@spike.ist](mailto:security@spike.ist).

See [SECURITY.md](SECURITY.md) for additional details.

## Community

Open Source is better together.

If you are a security enthusiast, [join SPIFFE's Slack Workspace][spiffe-slack]
and let us change the world together 🤘.

## Links

* **Homepage and Docs**: <https://spike.ist>
* **Community**:
    * [Join **SPIFFE** Slack Workspace][spiffe-slack]

## Folder Structure

Here are the important folders and files in this repository:

* `./app`: Contains **SPIKE** components' source code:
  * `./app/keeper`: **SPIKE** Keeper
  * `./app/nexus`: **SPIKE** Nexus
  * `./app/spike`: **SPIKE** Pilot
* `./config`: Contains configuration files to run SPIRE in a development
  environment.
* `./docs`: Public documentation.
* `./hack`: Useful scripts to build and test the project.
* `./internal`: Internal modules shared among **SPIKE** components.

## Code Of Conduct

[Be a nice citizen](CODE_OF_CONDUCT.md).

## Contributing

To contribute to **SPIKE**, [follow the contributing 
guidelines](CONTRIBUTING.md) to get started.

Use GitHub issues to request features or file bugs.

## Communications

* [SPIFFE **Slack** is where the community hangs out][spiffe-slack].
* [Send comments and suggestions to
  **feedback@spike.ist**](mailto:feedback@spike.ist).

## License

[Mozilla Public License v2.0](LICENSE).

[spiffe-slack]: https://slack.spiffe.io/
[spiffe]: https://spiffe.io/
[spike]: https://spike.ist/
[quickstart]: https://spike.ist/#/quickstart
