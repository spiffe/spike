![SPIKE](assets/spike-banner-lg.png)

## Secure Production Identity for Key Encryption (SPIKE)

**SPIKE** is a lightweight secrets store that uses [SPIFFE][spiffe]
as its identity control plane.

**SPIKE** protects your secrets and helps your ops, SREs, and sysadmins
`#sleepmore`.

For more information, [see the documentation][docs].

[docs]: https://spike.ist/
[spiffe]: https://spiffe.io/

## The Elevator Pitch

[**SPIKE**][spike] is a streamlined, highly reliable secrets store that leverages 
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
* **SPIKE Bootstrap**: An initialization app to securely bootstrap the entire
  system and deliver root key shards to **SPIKE Nexus**.

With its minimal footprint and robust security, **SPIKE** provides peace of mind 
for your team and critical data resilience when it counts.

## Project Maturity: Development  ![Development Phase](https://github.com/spiffe/spiffe/blob/main/.img/maturity/dev.svg)

**SPIKE** is a SPIFFE-affiliated project that has reached **Development** 
maturity as defined in the [SPIFFE Project Lifecycle][lifecycle]. This means:

* **SPIKE** is functionally stable and suitable for broader experimentation and
  community involvement.
* **SPIKE** is not yet production-ready, and certain features or interfaces may
  continue to evolve.
* Stability and polish are improving, but users should expect occasional bugs or
  breaking changes.

We invite developers and early adopters to explore, test, and contribute. Your
input is invaluable in helping us shape a robust and reliable product.

Use in critical systems is not advised at this time.
We'll announce when the project is ready for production adoption.

🦔 Thanks for your patience and support. We welcome your thoughts at
📬 [team@spike.ist](mailto:team@spike.ist). 


[lifecycle]: https://github.com/spiffe/spiffe/blob/main/NEW_PROJECTS.md

## Getting Your Hands Dirty

[Check out the quickstart guide][quickstart] to start playing with the project.

[You can also read the documentation][spike] to learn more about **SPIKE**'s
architecture and design philosophy.

## A Note on Security

We take **SPIKE**'s security seriously. If you believe you have
found a vulnerability, please responsibly disclose it to
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
  * `./app/nexus`: **SPIKE** Nexus (secrets store)
  * `./app/keeper`: **SPIKE** Keeper (root key redundancy)
  * `./app/spike`: **SPIKE** Pilot (CLI)
  * `./app/bootstrap`: **SPIKE** Bootstrap (initialization)
  * `./app/demo`: Demo workloads for testing
* `./internal`: Internal modules shared among **SPIKE** components.
* `./config`: Configuration files to run SPIRE in development.
* `./docs-src`: Documentation source files.
  * `./docs`: Generated documentation.
* `./hack`: Scripts for building and testing.
* `./examples`: Usage examples.
* `./makefiles`: Makefiles for building and testing.
* `./ci`: CI/CD configuration.
* `./dockerfiles`: Container build files.
* `./assets`: Images and other static assets.

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

[Apache v2.0](LICENSE).

[spiffe-slack]: https://slack.spiffe.io/
[spiffe]: https://spiffe.io/
[spike]: https://spike.ist/
[quickstart]: https://spike.ist/#/quickstart
