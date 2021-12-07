# dependabot-circleci
>⚠️Disclaimer⚠️ it's early days, prepare for breaking changes!

<br/>

![CircleCI](https://img.shields.io/circleci/build/github/BESTSELLER/dependabot-circleci/master)
![GitHub repo size](https://img.shields.io/github/repo-size/BESTSELLER/dependabot-circleci)
![GitHub All Releases](https://img.shields.io/github/downloads/BESTSELLER/dependabot-circleci/total)
![GitHub](https://img.shields.io/github/license/BESTSELLER/dependabot-circleci)

<br/>

`dependabot-circleci` is as its name suggests a small dependabot for circleci orbs and images.
We have created this as it is at the moment nearly impossible getting changes into the official [dependabot](https://github.com/dependabot/dependabot-core).

---
<br/>

## Getting Started
You enable `dependabot-circleci` by checking a dependabot-circleci.yml configuration file in to your repository's .github directory. `dependabot-circleci` then raises pull requests to keep the dependencies you configure up-to-date.

<br/>

#### Example *dependabot-circleci.yml* file

The example *dependabot-circleci.yml* file below configures version updates. If it finds outdated dependencies, it will raise pull requests against the target branch to update the dependencies.

```yaml
# example dependabot-circleci.yml file

assignees:
  - github_username # for a single user
  - org/team_name # for a whole team (nested teams is the same syntax org/team_name)
labels:
  - label1
  - label2
reviewers:
  - github_username # for a single user
  - org/team_name # for a whole team (nested teams is the same syntax org/team_name)
target-branch: main
directory: "/template" # Used if .github directory is nested inside another directory
schedule: "monthly" # Options are (daily, weekly, monthly)

```

---
<br/>

## Configuration options for dependency updates
The `dependabot-circleci` configuration file, dependabot-circleci.yml, uses YAML syntax. 
You must store this file in the .github directory of your repository.

| Option                            | Required | Description                            | Default                    |
| :-------------------------------- | :------: | :------------------------------------- | -------------------------- |
| [`assignees`](#assignees)         |          | Assignees to set on pull requests      | n/a                        |
| [`labels`](#labels)               |          | Labels to set on pull requests         | n/a                        |
| [`reviewers`](#reviewers)         |          | Reviewers to set on pull requests      | n/a                        |
| [`target-branch`](#target-branch) |          | Branch to create pull requests against | Default branch in the repo |
| [`directory`](#directory)         |          | Location of .github directory          | Root of repo               |
| [`schedule`](#schedule)           |          | When to look for updates               | daily                      |


---
<br/>

## Contributing
We are open for issues, pull requests etc.
