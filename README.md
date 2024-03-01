<a name="readme-top"></a>
<details>
  <summary>Table of Contents</summary>
  <ol>
    <li>
      <a href="#about-the-project">About The Project</a>
    </li>
    <li>
      <a href="#how-jig-works">How Jig Works</a>
    </li>
    <li>
      <a href="#install">Install</a>
    </li>
    <li>
      <a href="#getting-started">Getting Started</a>
      <ul>
        <li><a href="#configuration">Configuration</a></li>
        <li><a href="#running-jig-with-docker">Running Jig with Docker</a></li>
        <li><a href="#running-jig-from-command-line">Running Jig from Command Line</a></li>
      </ul>
    </li>
    <li><a href="#roadmap">Roadmap</a></li>
    <li><a href="#contributing">Contributing</a></li>
    <li><a href="#license">License</a></li>
  </ol>
</details>

# About The Project
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
* [![go-sprig][go-sprig]][go-sprig-url]
* [![yaml-jsonpath][yaml-jsonpath]][yaml-jsonpath]

and we're always open to contributions and suggestions for improvement.

<p align="right">(<a href="#readme-top">back to top</a>)</p>

# How Jig Works

Jig is a tool based on Go text templates. It extracts information from commit messages and issue trackers to generate release notes through the following steps:

1. **Accessing Git Repositories**: Jig utilizes the `model.yaml` file to store configuration details necessary for connecting to various Git repositories of the software product. This file also contains `version` and `previousVersion` fields. These fields represent the current and preceding versions of the component, respectively, and are used to parse the commit for the release note. Jig uses a separate configuration file named `config.yaml` to connect to Git and to the issue tracker. The configuration details include the URLs, usernames, and tokens/passwords for the Git repositories and the issue tracker.

2. **Parsing Commit Messages**: Once it has access to the repositories, Jig parses the commit messages. If the commit messages follow a certain format (like the Conventional Commits format), Jig extracts useful information such as the type of the commit (fix, feat, etc.) and a short summary of the changes.

3. **Linking to Issue Trackers**: If the commit messages contain issue keys, Jig uses these keys to link the commits to issues in the issue tracker (like Jira). This allows Jig to pull in additional information about the changes, such as the issue description, the issue status (open, closed, in progress, etc.).

4. **Enriching the Model File**: The extracted information is used to enrich the `model.yaml` file. This file serves as the data source for the Go text template. 

5. **Rendering the Template**: The `model.yaml` file is then used to render a completely customizable template. This template forms the basis of the generated release notes.

6. **Generating Release Notes**: Once Jig has all this information and the template is rendered, it generates the release notes. The release notes include a list of all the changes in the new version, grouped by the type of change (features, bug fixes, etc.) and the component they affect. Each change includes the commit message, the issue key, and a link to the issue in the issue tracker for more details.

This process can be automated, so that Jig generates the release notes every time a new version of the software product is released. This saves time and ensures that the release notes are accurate and comprehensive.

<p align="right">(<a href="#readme-top">back to top</a>)</p>

# Install
## Docker 
You can create custom Docker images using the provided Dockerfile:
```shell
docker build -t jig:latest .
```

## go Install

First, ensure that you have a compatible version of [Go](https://go.dev/) installed and set up on your system. The minimum required version of Go can be found in the [go.mod](go.mod) file.

To install Jig globally, execute the following command:

```shell
go install github.com/happyagosmith/jig@latest
```

<p align="right">(<a href="#readme-top">back to top</a>)</p>

# Getting Started
## Configuration
### config.yaml

Jig employs a distinct configuration file to establish connections to Git and the issue tracker. This file contains essential details such as URLs, usernames, and tokens/passwords for both the Git repositories and the issue tracker. Below is a sample configuration:

```yaml
jiraURL: "https://jiraURL/"
jiraUsername: "userEmail"
jiraPassword: "userJiraToken"
gitURL: "https://gitlab.com/"
gitToken: "userGitToken"
```

For a comprehensive list of properties that can be included in the file, refer to the help documentation by executing the following command in your terminal.

```shell
jig --help
```

### model.yaml
Jig uses the `model.yaml` file as configuration details to connect to the different Git repositories of the software product. Here is an example of what this configuration might look like:
```yaml
services:
- gitRepoID: 1234
  label: service1
  previousVersion: 0.0.1
  version: 0.0.2
  checkVersion: '@filepath:$.versions.a'
- gitRepoID: 5678
  jiraComponent: jComponent # used to retrieve the known issues from jira
  jiraProject: jProject # used to retrieve the known issues from jira
  label: service2
  previousVersion: 1.2.0
  version: 1.2.1
  checkVersion: '@filepath:$.a[?(@.b == ''label'')].c'
```

Each service in the services list represents a different component of the software product, with its own Git repository. The `gitRepoID` is the identifier of the Git repository, and the `version` and `previousVersion` fields indicate the current and previous versions of the component.

The `checkVersion` field allows you to specify the file and the YAML path from which the version information should be read. An example configuration might look like '@filepath:$.a.b'.

By executing the command 
```shell
jig setVersions model.yaml
```

the `version` field will be updated with the value pointed to by `checkVersion`. If a new value is provided, the `previousVersion` field will be updated with the former value of `version`.

This feature is particularly useful when the version is defined in a separate file within the repository, eliminating the need for redundant information. 

Besides, the model.yaml file will include the "gitRepoURL" and the "gitReleaseURL"  for each repo 
as well. Following an example:	

```yaml
services:
	- gitRepoID: 1234
	  label: service1
	  previousVersion: 0.0.1
	  version: 0.0.2
	  checkVersion: '@filepath:$.versions.a'
	  gitRepoURL: https://repo-service1-url
    gitReleaseURL: https://repo-service1-url/-/releases/0.0.2
	- gitRepoID: 5678
	  jiraComponent: jComponent # used to retrieve the known issues from jira
	  jiraProject: jProject # used to retrieve the known issues from jira
	  label: service2
	  previousVersion: 1.2.0
	  version: 1.2.1
	  checkVersion: '@filepath:$.a[?(@.b == ''label'')].c'
	  gitRepoURL: https://repo-service2-url/-/releases/1.2.1
    gitReleaseURL: https://repo-service2-url/-/releases/1.2.1
```

Upon running the command 

```shell
jig enrich model.yaml
```

the `model.yaml` file will be enriched with additional information. This includes details about fixed bugs, known issues pulled from the issue tracker, and Git repositories with their corresponding versions and extracted keys.

This enhanced `model.yaml` file then acts as the input for the Go text template.

Here is an example of the fields that Jig generates:

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
	  checkVersion: '@filepath:$.versions.a'
	  gitRepoURL: https://repo-service1-url
    gitReleaseURL: https://repo-service1-url/-/releases/0.0.2
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
    checkVersion: '@filepath:$.a[?(@.b == ''label'')].c'
	  gitURL: https://repo-service2-url/-/releases/1.2.1
    gitReleaseURL: https://repo-service2-url/-/releases/1.2.1
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

### release note template

Jig employs the  (go-template text)[https://pkg.go.dev/text/template] syntax to structure the release note. You can find an [example](examples/rn.tpl) that illustrates how to integrate the generated fields into the go-template.

Besides the standard functions offered by Go's text/template package, our library also encompasses a collection of custom functions and all functions provided by the Sprig library.

#### Using Custom Jig Functions in Go Text Templates
##### issuesFlatList
The issuesFlatList function is a custom function that takes an ExtractedIssue map as an argument and returns a slice of unique ExtractedIssue items.

###### Function Signature
```go
func(issuesMap ExtractedIssue) []ExtractedIssue
```
###### Usage

In your Go text template, you can use the issuesFlatList function like this:

```go
{{issuesFlatList .generatedValues.features}}
```

###### How It Works
The issuesFlatList function works by iterating over the ExtractedIssue map and creating a unique list of issues. It extracts the issue key from each issue and checks if it's already in the list of unique issues. If it's not, it adds the issue to the list. If it is, it updates the existing issue in the list. Additionally, for each unique issue, **it adds a key impactedServices** that contains a list of services that have had an impact.

#### Using Sprig Functions in Go Text Templates
Our library also includes all functions provided by the Sprig library. Sprig is a library that provides more than 100 commonly used template functions. It's inspired by the "Spring" Java library and the "Twig" PHP template engine.

You can use Sprig functions in your Go text templates just like you would use any other function. For a full list of available Sprig functions and their descriptions, please refer to the [Sprig documentation](https://masterminds.github.io/sprig/).

In your Go text template, you can utilize the issuesFlatList function in conjunction with the join function from Sprig. Here's an example of how you can do this:

```go
{{range (issuesFlatList .generatedValues.features)}} {{ .issueKey}}: {{ .impactedService | join ","}}{{end}}
```

In this example, the issuesFlatList function is used to iterate over the unique issues from .generatedValues.features. For each issue, it prints the issue key and the impacted services, with the services joined by a comma.

#### Retrieving the template from GitLab
You can retrieve the raw file of the template and of the model from GitLab using the following URL:

```shell
https://gitlab.com/api/v4/projects/12345/repository/files/xxxx/raw?ref=xxxx
```

Please note that the `PRIVATE-TOKEN` in the request header for authentication is automatically taken from the `gitToken` flag.

## Running Jig with Docker
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

### Running Jig from Command Line
The simplest way to run Jig (assuming the configuration file and the model file are present) is to execute:
```shell
jig --config config/.jig.yaml generate examples/rn.tpl -m models/model.yaml --withEnrich 
```

You might find it useful to save the release note into a file:
```shell
jig --config config/.jig.yaml generate examples/rn.tpl -m models/model.yaml --withEnrich -o rn.md
```

You also have the option to enrich the model.yaml file first without generating the Release Note, and then generate the Release Note at a later time:

```shell
jig --config config/.jig.yaml enrich models/model.yaml
jig --config config/.jig.yaml generate examples/rn.tpl -m models/model.yaml 
```

# Roadmap
See the [open issues](https://github.com/happyagosmith/jig/issues) for a full list of proposed features (and known issues).

<p align="right">(<a href="#readme-top">back to top</a>)</p>

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
[go-sprig]: https://img.shields.io/static/v1?label=sprig&message=latest&color=blue
[go-sprig-url]: https://masterminds.github.io/sprig/
[yaml-jsonpath]: https://img.shields.io/static/v1?label=yaml-jsonpath&message=v0.3.2&color=blue
[yaml-jsonpath-url]: https://github.com/vmware-labs/yaml-jsonpath