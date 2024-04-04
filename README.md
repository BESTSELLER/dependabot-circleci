# dependabot-circleci

<br/>

![CircleCI](https://img.shields.io/circleci/build/github/BESTSELLER/dependabot-circleci/master)
![GitHub repo size](https://img.shields.io/github/repo-size/BESTSELLER/dependabot-circleci)
![GitHub](https://img.shields.io/github/license/BESTSELLER/dependabot-circleci)

<br/>

`dependabot-circleci` is, as its name suggests, a small dependabot for CircleCI orbs and container images.
We have created this as at the time of creation it was nearly impossible to get changes into the official [dependabot](https://github.com/dependabot/dependabot-core).

---
<br/>

## Getting Started
1. Install the `dependabot-circleci` [GitHub App](https://github.com/apps/dependabot-circleci) in your organization.
2. You enable `dependabot-circleci` on specific repositories by creating a `dependabot-circleci.yml` configuration file in your repository's `.github` directory. `dependabot-circleci` then raise pull requests to keep the dependencies you configure up-to-date.

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
directory: "/.circleci" # Folder where the circleci config files are located
schedule: "monthly" # Options are (daily, weekly, monthly)

```

---
<br/>

## Configuration options for dependency updates
The `dependabot-circleci` configuration file, dependabot-circleci.yml, uses YAML syntax. 
You must store this file in the .github directory of your repository.

| Option                            | Required | Description                                        | Default                    |
|:----------------------------------|:--------:|:---------------------------------------------------|----------------------------|
| [`assignees`](#assignees)         |          | Assignees to set on pull requests                  | n/a                        |
| [`labels`](#labels)               |          | Labels to set on pull requests                     | n/a                        |
| [`reviewers`](#reviewers)         |          | Reviewers to set on pull requests                  | n/a                        |
| [`target-branch`](#target-branch) |          | Branch to create pull requests against             | Default branch in the repo |
| [`directory`](#directory)         |          | Folder where the circleci config files are located | `/.circleci`               |
| [`schedule`](#schedule)           |          | When to look for updates                           | daily                      |


---
<br/>

## Contributing
We are open for issues, pull requests etc.

## Running locally
1. Clone the repository
2. Make sure to have your secrets file in place  
    2.1 BESTSELLER folks can use Harpocrates to get them from Vault.
      ```bash
      harpocrates -f secrets-local.yaml --vault-token $(vault token create -format=json | jq -r '.auth.client_token')
      ```  
    2.2 Others will have to fill out this template in any other way.  
      ```json
      {
        "datadog": {
          "api_key": ""
        },
        "github": {
          "app": {
            "integration_id": "" ,
            "private_key": "",
            "webhook_secret": ""
          },
          "oauth": {
            "client_id": "",
            "client_secret": ""
          },
          "v3_api_url": "https://api.github.com/"
        },
        "http": {
          "token": ""
        },
        "server": {
          "port": 3000,
          "public_url": ""
        },
        "bestseller_specific": {
          "token": ""
        }
      }
      ```
3. Run `dependabot-circleci` by using Docker compose
    > `--build` will ensure that the latest version of the code is used
    ```bash
    docker-compose up --build
    ```