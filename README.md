# BUX
> Bitcoin UTXO & xPub Management Suite

[![Release](https://img.shields.io/github/release-pre/BuxOrg/bux.svg?logo=github&style=flat&v=1)](https://github.com/BuxOrg/bux/releases)
[![Build Status](https://img.shields.io/github/workflow/status/BuxOrg/bux/run-go-tests?logo=github&v=1)](https://github.com/BuxOrg/bux/actions)
[![Report](https://goreportcard.com/badge/github.com/BuxOrg/bux?style=flat&v=1)](https://goreportcard.com/report/github.com/BuxOrg/bux)
[![codecov](https://codecov.io/gh/BuxOrg/bux/branch/master/graph/badge.svg?v=1)](https://codecov.io/gh/BuxOrg/bux)
[![Mergify Status](https://img.shields.io/endpoint.svg?url=https://gh.mergify.io/badges/BuxOrg/bux&style=flat&v=1)](https://mergify.io)
[![Go](https://img.shields.io/github/go-mod/go-version/BuxOrg/bux?v=1)](https://golang.org/)
<br>
[![Gitpod Ready-to-Code](https://img.shields.io/badge/Gitpod-ready--to--code-blue?logo=gitpod)](https://gitpod.io/#https://github.com/BuxOrg/bux)
[![standard-readme compliant](https://img.shields.io/badge/readme%20style-standard-brightgreen.svg?style=flat)](https://github.com/RichardLitt/standard-readme)
[![Makefile Included](https://img.shields.io/badge/Makefile-Supported%20-brightgreen?=flat&logo=probot)](Makefile)
[![Sponsor](https://img.shields.io/badge/sponsor-mrz1836-181717.svg?logo=github&style=flat&v=1)](https://github.com/sponsors/mrz1836)
[![Donate](https://img.shields.io/badge/donate-bitcoin-ff9900.svg?logo=bitcoin&style=flat&v=1)](https://gobitcoinsv.com/#sponsor?utm_source=github&utm_medium=sponsor-link&utm_campaign=bux&utm_term=bux&utm_content=bux)

<br/>

## Table of Contents
- [About](#about)
- [Installation](#installation)
- [Documentation](#documentation)
- [Examples & Tests](#examples--tests)
- [Benchmarks](#benchmarks)
- [Code Standards](#code-standards)
- [Usage](#usage)
- [Contributing](#contributing)
- [License](#license)

<br/>

## About

> **TLDR;**
>
>Application developers should focus on their applications and should not be bogged down with managing UTXOs or XPubs. Developers should be able to use an open-source, easy to install solution to rapidly build full-featured Bitcoin applications.

<br/>

----
#### DISCLAIMER
> BUX is still considered _"ALPHA"_ and should not be used in production until v1.0.0
----
<br/>

#### BUX: Out-of-the-box Features:
- xPub & UTXO State Management (state, tip, balance, utxos, destinations)
- Bring your own Database (MySQL, PostgreSQL, SQLite, Mongo or interface your own)
- Caching (Ristretto, mCache, Redis or custom)
- Task Management (TaskQ, Machinery or custom)
- Transaction Processing (queue, broadcast, update state of xpubs)
- Plugins using [BRFC standards](http://bsvalias.org/01-brfc-specifications.html)

#### **Project Assumptions: MVP**
- _No private keys are used_, only the xPub (or access key) is given to bux
- (BYOX) `Bring your own xPub`
- Signing a transaction is outside this application (IE: buxServer or buxWallet)
- All transactions need to be submitted to the bux service to effectively track utxo states
- Database can be backed up, but not regenerated from chain
  - Certain data is not on chain, plus re-scanning xPub is expensive and not easily possible with WOC and limitations with blockchair


<br/>

## Installation

**go-template** requires a [supported release of Go](https://golang.org/doc/devel/release.html#policy).
```shell script
go get -u github.com/BuxOrg/bux
```

<br/>

## Documentation
View the generated [documentation](https://pkg.go.dev/github.com/BuxOrg/bux)

[![GoDoc](https://godoc.org/github.com/BuxOrg/bux?status.svg&style=flat&v=1)](https://pkg.go.dev/github.com/BuxOrg/bux)

<br/>

<details>
<summary><strong><code>Repository Features</code></strong></summary>
<br/>

This repository was created using [MrZ's `go-template`](https://github.com/mrz1836/go-template#about)

#### Built-in Features
- Continuous integration via [GitHub Actions](https://github.com/features/actions)
- Build automation via [Make](https://www.gnu.org/software/make)
- Dependency management using [Go Modules](https://github.com/golang/go/wiki/Modules)
- Code formatting using [gofumpt](https://github.com/mvdan/gofumpt) and linting with [golangci-lint](https://github.com/golangci/golangci-lint) and [yamllint](https://yamllint.readthedocs.io/en/stable/index.html)
- Unit testing with [testify](https://github.com/stretchr/testify), [race detector](https://blog.golang.org/race-detector), code coverage [HTML report](https://blog.golang.org/cover) and [Codecov report](https://codecov.io/)
- Releasing using [GoReleaser](https://github.com/goreleaser/goreleaser) on [new Tag](https://git-scm.com/book/en/v2/Git-Basics-Tagging)
- Dependency scanning and updating thanks to [Dependabot](https://dependabot.com) and [Nancy](https://github.com/sonatype-nexus-community/nancy)
- Security code analysis using [CodeQL Action](https://docs.github.com/en/github/finding-security-vulnerabilities-and-errors-in-your-code/about-code-scanning)
- Automatic syndication to [pkg.go.dev](https://pkg.go.dev/) on every release
- Generic templates for [Issues and Pull Requests](https://docs.github.com/en/communities/using-templates-to-encourage-useful-issues-and-pull-requests/configuring-issue-templates-for-your-repository) in Github
- All standard Github files such as `LICENSE`, `CONTRIBUTING.md`, `CODE_OF_CONDUCT.md`, and `SECURITY.md`
- Code [ownership configuration](.github/CODEOWNERS) for Github
- All your ignore files for [vs-code](.editorconfig), [docker](.dockerignore) and [git](.gitignore)
- Automatic sync for [labels](.github/labels.yml) into Github using a pre-defined [configuration](.github/labels.yml)
- Built-in powerful merging rules using [Mergify](https://mergify.io/)
- Welcome [new contributors](.github/mergify.yml) on their first Pull-Request
- Follows the [standard-readme](https://github.com/RichardLitt/standard-readme/blob/master/spec.md) specification
- [Visual Studio Code](https://code.visualstudio.com) configuration with [Go](https://code.visualstudio.com/docs/languages/go)
- (Optional) [Slack](https://slack.com), [Discord](https://discord.com) or [Twitter](https://twitter.com) announcements on new Github Releases
- (Optional) Easily add [contributors](https://allcontributors.org/docs/en/bot/installation) in any Issue or Pull-Request

</details>

<details>
<summary><strong><code>Package Dependencies</code></strong></summary>
<br/>

- [bitcoinschema/go-bitcoin](https://github.com/bitcoinschema/go-bitcoin)
- [bitcoinschema/go-map](https://github.com/bitcoinschema/go-map)
- [gorm.io/gorm](https://gorm.io/gorm)
- [libsv/go-bk](https://github.com/libsv/go-bk)
- [libsv/go-bt](https://github.com/libsv/go-bt)
- [mrz1836/go-cache](https://github.com/mrz1836/go-cache)
- [mrz1836/go-logger](https://github.com/mrz1836/go-logger)
- [mrz1836/go-mattercloud](https://github.com/mrz1836/go-mattercloud)
- [mrz1836/go-nownodes](https://github.com/mrz1836/go-nownodes)
- [mrz1836/go-whatsonchain](https://github.com/mrz1836/go-whatsonchain)
- [newrelic/go-agent](https://github.com/newrelic/go-agent)
- [stretchr/testify](https://github.com/stretchr/testify)
- [tonicpow/go-minercraft](https://github.com/tonicpow/go-minercraft)
- [tonicpow/go-paymail](https://github.com/tonicpow/go-paymail)
</details>

<details>
<summary><strong><code>Library Deployment</code></strong></summary>
<br/>

Releases are automatically created when you create a new [git tag](https://git-scm.com/book/en/v2/Git-Basics-Tagging)!

If you want to manually make releases, please install GoReleaser:

[goreleaser](https://github.com/goreleaser/goreleaser) for easy binary or library deployment to Github and can be installed:
- **using make:** `make install-releaser`
- **using brew:** `brew install goreleaser`

The [.goreleaser.yml](.goreleaser.yml) file is used to configure [goreleaser](https://github.com/goreleaser/goreleaser).

<br/>

### Automatic Releases on Tag Creation (recommended)
Automatic releases via [Github Actions](.github/workflows/release.yml) from creating a new tag:
```shell
make tag version=1.2.3
```

<br/>

### Manual Releases (optional)
Use `make release-snap` to create a snapshot version of the release, and finally `make release` to ship to production (manually).

<br/>

</details>

<details>
<summary><strong><code>Makefile Commands</code></strong></summary>
<br/>

View all `makefile` commands
```shell script
make help
```

List of all current commands:
```text
all                           Runs multiple commands
clean                         Remove previous builds and any cached data
clean-mods                    Remove all the Go mod cache
coverage                      Shows the test coverage
diff                          Show the git diff
generate                      Runs the go generate command in the base of the repo
godocs                        Sync the latest tag with GoDocs
help                          Show this help message
install                       Install the application
install-all-contributors      Installs all contributors locally
install-go                    Install the application (Using Native Go)
install-releaser              Install the GoReleaser application
lint                          Run the golangci-lint application (install if not found)
release                       Full production release (creates release in Github)
release                       Runs common.release then runs godocs
release-snap                  Test the full release (build binaries)
release-test                  Full production test release (everything except deploy)
replace-version               Replaces the version in HTML/JS (pre-deploy)
tag                           Generate a new tag and push (tag version=0.0.0)
tag-remove                    Remove a tag if found (tag-remove version=0.0.0)
tag-update                    Update an existing tag to current commit (tag-update version=0.0.0)
test                          Runs lint and ALL tests
test-all-db                   Runs all tests including embedded database tests
test-all-db-ci                Runs all tests including embedded database tests (CI)
test-ci                       Runs all tests via CI (exports coverage)
test-ci-no-race               Runs all tests via CI (no race) (exports coverage)
test-ci-short                 Runs unit tests via CI (exports coverage)
test-no-lint                  Runs just tests
test-short                    Runs vet, lint and tests (excludes integration tests)
test-unit                     Runs tests and outputs coverage
uninstall                     Uninstall the application (and remove files)
update-contributors           Regenerates the contributors html/list
update-linter                 Update the golangci-lint package (macOS only)
vet                           Run the Go vet application
```
</details>

<br/>

## Examples & Tests
All unit tests and [examples](examples) run via [Github Actions](https://github.com/BuxOrg/bux/actions) and
uses [Go version 1.17.x](https://golang.org/doc/go1.17). View the [configuration file](.github/workflows/run-tests.yml).

<br/>

Run all tests (including integration tests)
```shell script
make test
```

<br/>

Run tests (excluding integration tests)
```shell script
make test-short
```

<br/>

## Benchmarks
Run the Go benchmarks:
```shell script
make bench
```

<br/>

## Code Standards
Read more about this Go project's [code standards](.github/CODE_STANDARDS.md).

<br/>

## Usage
Checkout all the [examples](examples)!

<br/>

## Contributing
View the [contributing guidelines](.github/CONTRIBUTING.md) and follow the [code of conduct](.github/CODE_OF_CONDUCT.md).

<br/>

### How can I help?
All kinds of contributions are welcome :raised_hands:!
The most basic way to show your support is to star :star2: the project, or to raise issues :speech_balloon:.
You can also support this project by [becoming a sponsor on GitHub](https://github.com/sponsors/mrz1836) :clap:
or by making a [**bitcoin donation**](https://gobitcoinsv.com/#sponsor?utm_source=github&utm_medium=sponsor-link&utm_campaign=bux&utm_term=bux&utm_content=bux) to ensure this journey continues indefinitely! :rocket:

[![Stars](https://img.shields.io/github/stars/BuxOrg/bux?label=Please%20like%20us&style=social)](https://github.com/BuxOrg/bux/stargazers)

<br/>

### Contributors ✨
Thank you to these wonderful people ([emoji key](https://allcontributors.org/docs/en/emoji-key)):

<!-- ALL-CONTRIBUTORS-LIST:START - Do not remove or modify this section -->
<!-- prettier-ignore-start -->
<!-- markdownlint-disable -->
<table>
  <tr>
    <td align="center"><a href="https://mrz1818.com"><img src="https://avatars.githubusercontent.com/u/3743002?v=4?s=100" width="100px;" alt=""/><br /><sub><b>Mr. Z</b></sub></a><br /><a href="#infra-mrz1836" title="Infrastructure (Hosting, Build-Tools, etc)">🚇</a> <a href="https://github.com/mrz1836/go-template/commits?author=mrz1836" title="Code">💻</a> <a href="#maintenance-mrz1836" title="Maintenance">🚧</a> <a href="#security-mrz1836" title="Security">🛡️</a></td>
    <td align="center"><a href="https://github.com/icellan"><img src="https://avatars.githubusercontent.com/u/4411176?v=4?s=100" width="100px;" alt=""/><br /><sub><b>Siggi</b></sub></a><br /><a href="#infra-icellan" title="Infrastructure (Hosting, Build-Tools, etc)">🚇</a> <a href="https://github.com/mrz1836/go-template/commits?author=icellan" title="Code">💻</a> <a href="#security-icellan" title="Security">🛡️</a></td>
    <td align="center"><a href="https://github.com/galt-tr"><img src="https://avatars.githubusercontent.com/u/64976002?v=4?s=100" width="100px;" alt=""/><br /><sub><b>Dylan</b></sub></a><br /><a href="#infra-galt-tr" title="Infrastructure (Hosting, Build-Tools, etc)">🚇</a> <a href="https://github.com/mrz1836/go-template/commits?author=galt-tr" title="Code">💻</a></td>
  </tr>
</table>

<!-- markdownlint-restore -->
<!-- prettier-ignore-end -->

<!-- ALL-CONTRIBUTORS-LIST:END -->

> This project follows the [all-contributors](https://github.com/all-contributors/all-contributors) specification.

<br/>

## License

[![License](https://img.shields.io/github/license/BuxOrg/bux.svg?style=flat&v=1)](LICENSE)
