[![Build Status](https://travis-ci.org/jeremy-miller/standup-reporter.svg?branch=master)](https://travis-ci.org/jeremy-miller/standup-reporter)
[![Coverage Status](https://coveralls.io/repos/github/jeremy-miller/standup-reporter/badge.svg?branch=master)](https://coveralls.io/github/jeremy-miller/standup-reporter?branch=master)
[![GoDoc](https://godoc.org/github.com/jeremy-miller/standup-reporter?status.svg)](https://godoc.org/github.com/jeremy-miller/standup-reporter)
[![Go Report Card](https://goreportcard.com/badge/github.com/jeremy-miller/standup-reporter)](https://goreportcard.com/report/github.com/jeremy-miller/standup-reporter)
[![GitHub release](https://img.shields.io/github/release/jeremy-miller/standup-reporter.svg)](https://github.com/jeremy-miller/standup-reporter/releases)
[![Conventional Commits](https://img.shields.io/badge/Conventional%20Commits-1.0.0-yellow.svg)](https://conventionalcommits.org)
[![semantic-release](https://img.shields.io/badge/%20%20%F0%9F%93%A6%F0%9F%9A%80-semantic--release-e10079.svg)](https://github.com/semantic-release/semantic-release)
[![Powered By: GoReleaser](https://img.shields.io/badge/Powered%20By-Goreleaser-brightgreen.svg)](https://github.com/goreleaser)
[![MIT Licensed](https://img.shields.io/badge/license-MIT-blue.svg)](https://github.com/jeremy-miller/standup-reporter/blob/master/LICENSE)

# Standup Reporter
Generate reports for standup meetings.

Currently only [Asana](https://asana.com/) is supported.  The `standup-reporter` will print both completed tasks from a
configurable number of days in the past, as well as all incomplete tasks.  All projects in your Asana workspace will be
used.

## Install
To install `standup-reporter`, download the
[latest release](https://github.com/jeremy-miller/standup-reporter/releases/latest).

## Usage
The usage for `standup-reporter` is also available by using the `--help` or `-h` switch.

### Create Asana Personal Access Token
An Asana personal access token is required to use `standup-reporter`.  To create an Asana personal access token:

1. Login to [Asana](https://asana.com/)
2. Click your profile photo from the top bar and select _My Profile Settings..._
3. Navigate to the _Apps_ tab
4. Select _Manage Developer Apps_
5. Select _+ Create New Personal Access Token_
6. Add a _Description_ and choose _Never include numeric IDs_ under _Webhook ID Behavior_
7. Click the _Create_ button

## Development
Below are instructions for developing `standup-reporter` locally.

### Prerequisites
- [Go](https://golang.org/dl/)

### Makefile
To view all available `make` targets: `make help`

### Environment Setup
To setup the local development environment: `make setup`

#### Pre-Commit
One tool which is installed during `make setup` is [pre-commit](https://pre-commit.com/).  To update `pre-commit`
hooks: `make update-hooks`

### Build
To build `standup-reporter` for the supported operating systems and architectures: `make build`

### Compile Without Outputting Binaries
To build all packages without producing binaries (i.e. check for errors): `make check`

### `modd`
To use [`modd`](https://github.com/cortesi/modd) to trigger builds based on filesystem changes, download a binary from
[here](https://github.com/cortesi/modd/releases/latest).

#### Run `modd`
To run: `make modd`

When running, it will re-run commands from the `config/modd.conf` file when any `*.go` files change.

### Lint
To lint all files: `make lint`

### Test
To run all tests (including data race checking): `make test`

### Test Coverage
To run all tests (including data race checking) and generate code coverage: `make coverage`

### Update Dependencies
To update all dependency versions: `make update-deps`

### Remove Unused Dependencies
To remove unused dependencies: `make tidy`

### Clean
To remove generated/compiled files: `make clean`

## License
[MIT](https://github.com/jeremy-miller/standup-reporter/blob/master/LICENSE)
