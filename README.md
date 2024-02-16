<!-- Improved compatibility of back to top link: See: https://github.com/othneildrew/Best-README-Template/pull/73 -->
<a name="readme-top"></a>

<!-- PROJECT LOGO -->
<br />
<div align="center">
    <img src="images/logo.png" alt="Logo" width="80" height="80">
</div>

<!-- TABLE OF CONTENTS -->
<details>
  <summary>Table of Contents</summary>
  <ol>
    <li>
      <a href="#about-the-project">About The Project</a>
    </li>
    <li>
      <a href="#getting-started">Getting Started</a>
      <ul>
        <li><a href="#docker-usage">Docker Usage</a></li>
        <li><a href="#building-from-source">Building from Source</a></li>
      </ul>
    </li>
    <li><a href="#roadmap">Roadmap</a></li>
    <li><a href="#contributing">Contributing</a></li>
    <li><a href="#license">License</a></li>
  </ol>
</details>

<!-- ABOUT THE PROJECT -->
## About The Project
The project, named "Jig", is a tool designed to automate the creation of release notes for software products composed of one or more components, each with its own Git repository. Jig leverages the information found in commit messages across these repositories and enriches it with details from the issue tracker (e.g., Jira).

The primary goal of Jig is to streamline the process of generating release notes, which is often a time-consuming task. By automatically pulling data from commit messages and issue trackers, Jig ensures that all relevant changes are documented, reducing the risk of missing important information.

Jig is particularly useful for large projects with multiple components and frequent updates. It provides a standardized format for release notes, making it easier for users to understand what has changed in each version.

Here's why:
* Automating the generation of release notes saves valuable time and reduces the risk of human error.
* By using data from commit messages and issue trackers, Jig ensures that all relevant changes are included in the release notes.
* Standardizing the format of release notes improves readability and helps users understand the changes in each release.

Jig assumes that each commit contains an issue key, which can be in the conventional commit format (where the scope is the issue key) or any other format.

This project is built with:

* [![Cobra][Cobra]][Cobra-url]
* [![go-template][go-template]][go-template-url]

and we're always open to contributions and suggestions for improvement.

## How Jig Works

Jig is a tool based on Go text templates. It extracts information from commit messages and issue trackers to generate release notes through the following steps:

1. **Accessing Git Repositories**: Jig uses the `model.yaml` file as configuration details to connect to the different Git repositories of the software product. Here is an example of what this configuration might look like:
```yaml
services:
- gitRepoID: 1234
  label: service1
  previousVersion: 0.0.1
  version: 0.0.2
- gitRepoID: 5678
  jiraComponent: jComponent # used to retrieve the known issues from jira
  jiraProject: jProject # used to retrieve the known issues from jira
  label: service2
  previousVersion: 1.2.0
  version: 1.2.1
```

Each service in the services list represents a different component of the software product, with its own Git repository. The gitRepoID is the identifier of the Git repository, and the version and previousVersion fields indicate the current and previous versions of the component.

Jig uses a separate configuration file named `config.yaml` to connect to Git and to the issue tracker. The configuration details include the URLs, usernames, and tokens/passwords for the Git repositories and the issue tracker. Here is an example of what this configuration might look like:

```yaml
jiraURL: "https://jiraURL/"
jiraUsername: "userEmail"
jiraPassword: "userJiraToken"
gitURL: "https://gitlab.com/"
gitToken: "userGitToken"
```

2. **Parsing Commit Messages**: Once it has access to the repositories, Jig parses the commit messages. If the commit messages follow a certain format (like the Conventional Commits format), Jig extracts useful information such as the type of the commit (fix, feat, etc.) and a short summary of the changes.

3. **Linking to Issue Trackers**: If the commit messages contain issue keys, Jig uses these keys to link the commits to issues in the issue tracker (like Jira). This allows Jig to pull in additional information about the changes, such as the issue description, the issue status (open, closed, in progress, etc.), and any comments or discussions about the issue.

4. **Enriching the Model File**: The extracted information is used to enrich the `model.yaml` file. This file serves as the data source for the Go text template. In addition to the configuration details mentioned in the point 1, Jig also generates and adds the following information to the `model.yaml` file:

```yaml
generatedValues:
  features: 
    service1:
    - issueKey: AAA-000
      issueSummary: 'Fix Comment from the issue tracker'
      issueType: TECH TASK
      issueStatus: Completata
      issueTrackerType: JIRA
      category: CLOSED_FEATURE
      isbreakingchange: true
  bugs:
    service1:
    - issueKey: AAA-111
      issueSummary: 'Fix Comment from the issue tracker'
      issueType: Bug
      issueStatus: RELEASED
      issueTrackerType: JIRA
      category: FIXED_BUG
      isBreakingChange: false
    service2:
    - issueKey: AAA-222
      issueSummary: 'Fix Comment from the issue tracker'
      issueType: Bug
      issueStatus: FIXED
      issueTrackerType: JIRA
      category: FIXED_BUG
      isBreakingChange: false
  knownIssues:
    service2:
    - issueKey: AAA-333
      issueType: TECH DEBT
      issueSummary: To be implemented
      issueStatus: Da completare
      issueTrackerType: JIRA
      category: OTHER
      isBreakingChange: false
  breakingChange: 
    service1:
    - issueKey: AAA-000
      issueSummary: 'Fix Comment from the issue tracker'
      issueType: TECH TASK
      issueStatus: Completata
      issueTrackerType: JIRA
      category: CLOSED_FEATURE
      isbreakingchange: true
  gitRepos:
  - gitRepoID: 1234
    label: service1
    previousVersion: 0.0.1
    version: 0.0.2
    extractedKeys:
    - category: BUG_FIX
      issueKey: AAA-000
      summary: 'fix comment from git'
      message: 'fix(j_AAA-000)!: fix comment from git'
      issueTrackerType: JIRA
      isbreakingchange: true
    - issueKey: AAA-111
      summary: 'fix comment from git'
      message: '[AAA-111] fix comment from git'
      issueTrackerType: JIRA
      isbreakingchange: false
  - gitRepoID: 5678
    label: service2
    jiraComponent: jComponent # used to retrieve the known issues
    jiraProject: jProject # used to retrieve the known issues
    previousVersion: 1.2.0
    version: 1.2.1
    extractedKeys:
    - category: BUG_FIX
      issueKey: AAA-222
      summary: 'fix comment from git'
      message: 'fix(j_AAA-222): fix comment from git'
      issueTrackerType: JIRA
      isbreakingchange: false
    - issueKey: AAA-333
      summary: 'comment from git'
      message: '[AAA-333] comment from git'
      issueTrackerType: JIRA
      isbreakingchange: false
```
This information includes the bugs fixed, known issues extracted from the issue tracker and the git repso with their respective versions and extracted keys from the git repository.

5. **Rendering the Template**: The `model.yaml` file is then used to render a completely customizable template. This template forms the basis of the generated release notes.

6. **Generating Release Notes**: Once Jig has all this information and the template is rendered, it generates the release notes. The release notes include a list of all the changes in the new version, grouped by the type of change (features, bug fixes, etc.) and the component they affect. Each change includes the commit message, the issue key, and a link to the issue in the issue tracker for more details.

This process can be automated, so that Jig generates the release notes every time a new version of the software product is released. This saves time and ensures that the release notes are accurate and comprehensive.

For more details on how to use Jig, including how to use the `extractedKeys` feature, please refer to the command line help. You can access this help by running the command `jig --help` in your terminal.

<p align="right">(<a href="#readme-top">back to top</a>)</p>

## Getting Started

## Docker Usage

### Building the Docker Image
You can create custom Docker images using the provided Dockerfile:
```shell
docker build -t jig:latest .
```

### Running Jig with Docker
The most straightforward way to run Jig (assuming you're in the root directory of the Git repository and the configuration file is available) is:

```shell
docker run -v ./examples:/config -v ./examples:/models --rm jig:latest --config config/.jig.yaml generate examples/rn.tpl -m models/model.yaml --withEnrich 
```

You might find it useful to redirect the standard output to a file:
```shell
docker run -v ./examples:/config -v ./examples:/models --rm jig:latest --config config/.jig.yaml generate examples/rn.tpl -m models/model.yaml --withEnrich 2>&1 | tee rn.md
```

Then, you can use the cat command to extract the lines related to the Release Note:

```shell
cat rn.md | sed -n '/\#/,$p'
```

<!-- DOCKER -->

## Building from Source

First, ensure that you have a compatible version of [Go](https://go.dev/) installed and set up on your system. The minimum required version of Go can be found in the [go.mod](go.mod) file.

To install Jig globally, execute the following command:

```shell
go install github.com/happyagosmith/jig@latest
```

<p align="right">(<a href="#readme-top">back to top</a>)</p>

<!-- SOURCE -->
### Running Jig from Command Line
The simplest way to run Jig (assuming the configuration file and the model file are present) is to execute:
```shell
jig --config config/.jig.yaml generate examples/rn.tpl -m models/model.yaml --withEnrich 
```

You might find it useful to redirect the standard output to a file:
```shell
jig --config config/.jig.yaml generate examples/rn.tpl -m models/model.yaml --withEnrich 2>&1 | tee rn.md
```

Then, you can use the cat command to extract the lines related to the Release Note:

```shell
cat rn.md | sed -n '/\#/,$p'
```

Or, you can directly output only the Release Note content:

```shell
jig --config config/.jig.yaml generate examples/rn.tpl -m models/model.yaml --withEnrich 2>&1 | sed -n '/\#/,$p'
```

You also have the option to enrich the model.yaml file first without generating the Release Note, and then generate the Release Note at a later time:

```shell
jig --config config/.jig.yaml enrich models/model.yaml
jig --config config/.jig.yaml generate examples/rn.tpl -m models/model.yaml 
```

<p align="right">(<a href="#readme-top">back to top</a>)</p>

<!-- ROADMAP -->
## Roadmap
See the [open issues](https://github.com/happyagosmith/jig/issues) for a full list of proposed features (and known issues).

<p align="right">(<a href="#readme-top">back to top</a>)</p>

<!-- CONTRIBUTING -->
## Contributing

Contributions are what make the open source community such an amazing place to learn, inspire, and create. Any contributions you make are **greatly appreciated**.

If you have a suggestion that would make this better, please fork the repo and create a pull request. You can also simply open an issue with the tag "enhancement".
Don't forget to give the project a star! Thanks again!

1. Fork the Project
2. Create your Feature Branch (`git checkout -b feature/AmazingFeature`)
3. Commit your Changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the Branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

<p align="right">(<a href="#readme-top">back to top</a>)</p>

<!-- LICENSE -->
## License

Distributed under the Apache Version 2.0 License. See `LICENSE.txt` for more information.

<p align="right">(<a href="#readme-top">back to top</a>)</p>


<!-- MARKDOWN LINKS & IMAGES -->
<!-- https://www.markdownguide.org/basic-syntax/#reference-style-links -->
[license-url]: https://github.com/happyagosmith/geco/blob/main/LICENSE
[go-template]: https://img.shields.io/static/v1?label=text-template&message=latest&color=blue
[go-template-url]: https://pkg.go.dev/text/template
[cobra]: https://img.shields.io/static/v1?label=cobra&message=v1.7.0&color=blue
[cobra-url]: https://github.com/spf13/cobra