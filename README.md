# Drone Pull Request Checker Plugin

Drone plugin to run various checks on the Pull request as part of review process

> **IMPORTANT:** This plugin is under development and the parameters are subject to change

## Usage

The following settings changes this plugin's behavior,

| Property            |  Type   | Description                                    |   Default    |
| :------------------ | :-----: | :--------------------------------------------- | :----------: |
| `prefixes`          |  list   | A list of accepted PR title prefixes           |      []      |
| `regexp`            | string  | A regular expression for a valid PR title      |      ""      |
| `skipOnLabels`      |  list   | A list of on which the checks will be disabled |      []      |
| `ignoreGithubError` | boolean | A boolean value to ignore github api errors    |    false     |
| `checklist`         | boolean | A boolean value to enable checklist checks     |    false     |
| `checklistTitle`    | string  | A string value from which to find PR checklist | ## Checklist |

## Credentials

- `github_token`: its required to access pull request data to check for labels and checklists on the PR content.

## Pipeline

```yaml
kind: pipeline
type: docker
name: default

steps:
  - name: check pull request
    image: thomasnyambati/drone-pr-checker
    settings:
      prefixes: []
      regexp: ""
      skipOnLabels: []
      ignoreGithubError: false
      checklist: false
      checklistTitle: ""
    environment:
      GITHUB_TOKEN:
        from_secret: github_token
```

Now load the image using the command,

```shell
docker load < ./dist/hello-world-0.0.1_$(uname -m).tar
```

## Building Plugin

The plugin build relies on:

- [Makefile](https://www.gnu.org/software/make/manual/make.html)

Build plugin

```shell
make build
```

## Testing

Build plugin packages,

```shell
make test
```

Build plugin container image,

```shell
make build
```

Create `.env`

```shell
cat<<EOF | tee .env
PLUGIN_PREFIXES="feat:,chore:,hotfix:"
PLUGIN_REGEXP="^(chore|hotfix|docs):.+"
PLUGIN_SKIP_ON_LABELS="skip"
PLUGIN_IGNORE_GITHUB_ERROR="false"
PLUGIN_CHECKLIST="true"
PLUGIN_CHECKLIST_TITLE="## Checklist"
DRONE_PULL_REQUEST_TITLE="feat: sample pull request"
EOF
```

Run the plugin

```shell
docker run --rm --env-file=.env  --volume "$PWD:/workspace" thomasnyambati/drone-pr-checker:main
```
