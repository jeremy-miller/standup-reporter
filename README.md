# Standup Reporter
Generate reports for standup meetings.

## Table of Contents
- [Git Pre-Commit Hooks](#git-pre-commit-hooks)
    - [Install `pre-commit`](#install-pre-commit)
    - [Update Hooks](#update-hooks)
- [Run](#run)
    - [Create Asana Personal Access Token](#create-asana-personal-access-token)

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

### Create Asana Personal Access Token
To create an Asana personal access token:

1. Login to Asana
2. Click your profile photo from the top bar and select _My Profile Settings_
3. Navigate to the _Apps_ tab
4. Select _Personal Access Token_
5. Add a _Description_
5. Click the _Create_ button
