# Standup Reporter
Generate reports for standup meetings.

## Table of Contents
- [Git Pre-Commit Hooks](#git-pre-commit-hooks)
    - [Install `pre-commit`](#install-pre-commit)
    - [Update Hooks](#update-hooks)
- [Run](#run)

## Git Pre-Commit Hooks
### Install `pre-commit`
To install [pre-commit](https://pre-commit.com/):
1. `pip3 install pre-commit`
2. `pre-commit install -c githooks/.pre-commit-config.yaml`

After install, the pre-commit hooks configured in `.pre-commit-config.yaml` will be executed before every commit.

### Update Hooks
To update `pre-commit` hooks: `make update-hooks`

## Run
To run: `standup-reporter --asana=<token>`
