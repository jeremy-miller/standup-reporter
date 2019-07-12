[![Build Status](https://travis-ci.org/jeremy-miller/standup-reporter.svg?branch=master)](https://travis-ci.org/jeremy-miller/standup-reporter)
[![Go Report Card](https://goreportcard.com/badge/github.com/jeremy-miller/standup-reporter)](https://goreportcard.com/report/github.com/jeremy-miller/standup-reporter)
[![Conventional Commits](https://img.shields.io/badge/Conventional%20Commits-1.0.0-yellow.svg)](https://conventionalcommits.org)
[![MIT Licensed](https://img.shields.io/badge/license-MIT-blue.svg)](https://github.com/jeremy-miller/standup-reporter/blob/master/LICENSE)

# Standup Reporter
Generate reports for standup meetings.

## Table of Contents
- [Git Pre-Commit Hooks](#git-pre-commit-hooks)
    - [Install `pre-commit`](#install-pre-commit)
    - [Update Hooks](#update-hooks)
- [`modd`](#modd)
    - [Run `modd`](#run-modd)
- [Run](#run)
    - [Create Asana Personal Access Token](#create-asana-personal-access-token)

## Git Pre-Commit Hooks
### Install `pre-commit`
To install [pre-commit](https://pre-commit.com/):
1. `pip3 install pre-commit`
2. Pre-commit hooks: `pre-commit install -c githooks/.pre-commit-config.yaml -t pre-commit`
3. Pre-push hooks: `pre-commit install -c githooks/.pre-commit-config.yaml -t pre-push`

After install, the pre-commit hooks configured in `.pre-commit-config.yaml` will be executed before every commit.

### Update Hooks
To update `pre-commit` hooks: `make update-hooks`

## `modd`
Install [`modd`](https://github.com/cortesi/modd) by downloading a binary from [here](https://github.com/cortesi/modd/releases/latest).

### Run `modd`
To run: `make modd`

When running, it will re-run commands from the `config/modd.conf` file when any `*.go` files change.

## Run
To run: `standup-reporter --asana=<token>`

### Create Asana Personal Access Token
To create an Asana personal access token:

1. Login to Asana
2. Click your profile photo from the top bar and select _My Profile Settings..._
3. Navigate to the _Apps_ tab
4. Select _Manage Developer Apps_
5. Select _+ Create New Personal Access Token_
6. Add a _Description_ and choose _Never include numeric IDs_ under _Webhook ID Behavior_
7. Click the _Create_ button
