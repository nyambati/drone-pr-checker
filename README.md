# Drone Pull Request Checker Plugin

Drone plugin to run various checks on the Pull request as part of review process

> **IMPORTANT:** This plugin is under development and the parameters are subject to change

## Usage

The following settings changes this plugin's behavior,

| Property            |  Type   |      Default |
| ------------------- | :-----: | -----------: |
| `prefixes`          | string  |           "" |
| `regexp`            | string  |           "" |
| `skipOnLabels`      | string  |           "" |
| `ignoreGithubError` | boolean |        false |
| `checklist`         | boolean |        false |
| `checklistTitle`    | string  | ## Checklist |

- `prefixes`: The accepted pull request title prefixes.
- `regexp`: The regular expression for a valid pull request title.
- `skipOnLabels`: The labelat which the checks will be disabled.
- `ignoreGithubError`: Ignores github error when fetching Pull request labels and checklist.
- `checklist`: Ensures that all checklist items are checked.
- `checklistTitle`: The title used to find the checklist.

## Credentials

- `github_token`: This will be used to fetch pull request labels and content from github.

## Pipeline

```yaml
kind: pipeline
type: docker
name: default

steps:
  - name: check pull request
    image: thomasnyambati/drone-pr-checker
    settings:
      prefixes:
      regexp:
      skipOnLabels:
      ignoreGithubError:
      checklist:
      checklistTitle:
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
docker run --rm \
  --env-file=.env \
  --volume "$PWD:/workspace" \
  thomasnyambati/drone-pr-checker:latest
```
