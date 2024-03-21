<a name="readme-top"></a>
<details>
  <summary>Table of Contents</summary>
  <ol>
    <li>
      <a href="#key-concepts-and-best-practices-for-writing-release-notes"> Key Concepts and Best Practices for Writing Release Notes</a>
    </li>
    <li>
      <a href="#how-jig-works">How Jig Works</a>
    </li>
    <li>
      <a href="#parsing-mechanisms">Parsing Mechanisms</a>
    </li>
    <li>
      <a href="#install">Install</a>
    </li>
    <li>
      <a href="#getting-started">Getting Started</a>
      <ul>
        <li><a href="#running-jig-from-command-line">Running Jig from Command Line</a></li>
        <li><a href="#configuration">Configuration</a></li>
      </ul>
    </li>
    <li><a href="#roadmap">Roadmap</a></li>
    <li><a href="#contributing">Contributing</a></li>
    <li><a href="#license">License</a></li>
  </ol>
</details>

<p align="center" style="text-align: center">
  <h1 align="center" style="text-align: center">Jig</h1>
  <p align="center" style="text-align: center">A release note generator tool</p>
</p>

[![Run ci](https://github.com/happyagosmith/jig/actions/workflows/ci.yml/badge.svg)](https://github.com/happyagosmith/jig/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/happyagosmith/jig)](https://goreportcard.com/report/github.com/happyagosmith/jig)

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

# Key Concepts and Best Practices for Writing Release Notes

## Key Concepts

- **Issue**: An issue is a note that describes a task, feature, or bug related to the project. It represents the "what" of the work that needs to be done. It's a way of tracking and discussing the requirements and goals of a particular piece of work. Issues are typically used to discuss the work before it's done, and can be assigned to specific team members, categorized, and scheduled.

- **Merge Request**: A merge request (also known as a pull request in some platforms) is a proposal to merge a branch into another. It represents the "how" of implementing the work described in an issue. Unlike an issue, a merge request is about discussing a specific set of proposed changes. It includes the changes themselves (in the form of commits), as well as comments and reviews about those changes. Merge requests are typically used after the work has been done, to review it and then integrate it into the project.

- **Commit**: A commit is a change to a file (or set of files) in your project. It's a way of saving your work. Each commit has a unique ID that can be used to refer to that specific set of changes.

- **Issue Tracker**: An issue tracker is a tool that keeps track of issues. It's a way of organizing and managing work on the project.

- **Repo**: A repo (short for repository) is a location where all the files for a particular project are stored. It includes all the versions of all the files in the project, as well as the history of all changes made to them.

## Writing Release Notes

Creating effective release notes is crucial to provide clear and concise information about the changes in the new version of the software. Here's a guideline on what they should contain:

- **Introduction**: Start with a brief summary of the product and the purpose of the release. This sets the context for the rest of the release notes.

- **New Features**: Detail the new features in the release. For each feature, reference the issue it resolves and the merge request where it was implemented. This provides a clear link between the release notes and the detailed discussions and changes in the issue tracker and repo.

- **Improvements**: Describe any enhancements to existing features. Again, reference the relevant issues, merge requests, or commits. This helps users understand the evolution of existing features.

- **Bug Fixes**: List the bugs that have been fixed. Include references to the issues they resolve and the merge requests or commits where the fixes were implemented. This reassures users that problems have been addressed.

- **Known Issues**: Mention any known issues that are still present. Reference the issues where they are tracked. This keeps users informed about ongoing problems.

- **Deprecations**: Notify about any features that are being deprecated. Reference the issues, merge requests, or commits where the decision was made. This prepares users for future changes.

- **Thanks**: Finally, acknowledge the people who contributed to the release. This fosters a sense of community and gives credit where it's due.

Keep in mind, the purpose of release notes is to update users about the modifications in the new software version, enabling them to grasp what's been added, altered, and its impact on them. When outlining a change, refer to the issue that it addresses - the "what" - and the merge request where the change was executed - the "how". This establishes a direct connection between the release notes and the comprehensive discussions and alterations in the issue tracker and repository.

<p align="right">(<a href="#readme-top">back to top</a>)</p>

# How Jig Works

Jig is a tool based on Go text templates. It extracts information from commit messages and issue trackers to generate release notes through the following steps:

1. **Accessing Git Repositories**: Jig utilizes the `model.yaml` file to store configuration details necessary for connecting to various Git repositories of the software product. This file also contains `version` and `previousVersion` fields. These fields represent the current and preceding versions of the component, respectively, and are used to parse the commit for the release note. Jig uses a separate configuration file named `config.yaml` to connect to Git and to the issue tracker. The configuration details include the URLs, usernames, and tokens/passwords for the Git repositories and the issue tracker.

2. **Parsing Commit Messages**: Once it has access to the repositories, Jig parses the commit messages. If the commit messages follow a certain format (like the Conventional Commits, Closing Pattern or custom format), Jig extracts useful information such as the type of the commit (fix, feat, etc.) and a short summary of the changes.

3. **Linking to Issue Trackers**: If the commit messages contain issue keys, Jig uses these keys to link the commits to issues in the issue tracker (like Jira). This allows Jig to pull in additional information about the changes, such as the issue description, the issue status (open, closed, in progress, etc.).

4. **Enriching the Model File**: The extracted information is used to enrich the `model.yaml` file. This file serves as the data source for the Go text template. 

5. **Rendering the Template**: The `model.yaml` file is then used to render a completely customizable template. This template forms the basis of the generated release notes.

6. **Generating Release Notes**: Once Jig has all this information and the template is rendered, it generates the release notes. The release notes include a list of all the changes in the new version, grouped by the type of change (features, bug fixes, etc.) and the component they affect. Each change includes the commit message, the issue key, and a link to the issue in the issue tracker for more details.

This process can be automated, so that Jig generates the release notes every time a new version of the software product is released. This saves time and ensures that the release notes are accurate and comprehensive.

<p align="right">(<a href="#readme-top">back to top</a>)</p>

# Parsing Mechanisms

Jig utilizes multiple parsing strategies to extract data from merge requests and commits. The `gitMRBranch` parameter, if specified, denotes the branch that the merge request is being analyzed and processed for. Without this parameter, merge requests will not undergo processing.

The following parsing mechanisms are available:

## Conventional Commit Parsing

This parsing mechanism is used on the title of both the merge request and the commit. It is designed to extract the scope, type, and subject from the commit.

## Closing Pattern Parsing

This parser operates on the descriptions of both the commit and the merge request. If you incorporate certain keywords followed by issue numbers (for example, "Closes #4, #6, Related to #5") in a merge request description or commit, the parser will recognize and extract these issues.

The acceptable keywords include: Close, Closes, Closed, Closing, close, closes, closed, closing, Fix, Fixes, Fixed, Fixing, fix, fixes, fixed, fixing, Resolve, Resolves, Resolved, Resolving, resolve, resolves, resolved, resolving, Implement, Implements, Implemented, Implementing, implement, implements, implemented, implementing.

When the keywords 'close', 'implement' or their variations are used, the issue is categorized as a feature. Conversely, if the keywords 'fix', 'resolve' or their variations are used, the issue is categorized as a bug.

## Custom Pattern Parsing

This parsing mechanism is used on the title based on a custom regular expression pattern that extracts the scope and type, similar to the conventional commit parsing.

This can be configured with the `customCommitPattern` parameter. The pattern should include the named groups scope and subject.

## Issue Tracker Parsing

For each key extracted by the previous parsers, a configurable matching logic is applied to associate each key with its corresponding issue tracker. For instance, the key #1 is identified as a GIT issue, while ABC-123 is identified as a Jira issue. This is configurable.

This can be configured with the `issuePatterns` parameter. Here's an example:
```yaml
issuePatterns:
  - issuetracker: jira
    pattern: j_(.+)
  - issuetracker: git
    pattern: '#(\\d+)
```

After the issues have been parsed, the corresponding trackers are queried to categorize the issues as either features or bugs. 

For GIT, if the issue has the label 'bug', it is classified as a fixed bug. If it has the label 'feature', it is classified as a closed feature. If the issue does not have a label, it is considered a feature by default.

For Jira, the classification can be configured using the following parameters:

- `--jiraClosedFeatureFilter`: This is a list of filters of type:status that identify the closed features. The default value is "Story:GOLIVE,TECH TASK:Completata".
- `--jiraFixedBugFilter`: This is a list of filters of type:status that identify the fixed bugs. The default value is "BUG:FIXED,BUG:RELEASED".
- `--jiraKnownIssuesJQL`: This is a Jira JQL to retrieve the known issues. The default value is "status not in (Done, RELEASED, Fixed, GOLIVE, Cancelled) AND issuetype in (Bug, \"TECH DEBT\")".


<p align="right">(<a href="#readme-top">back to top</a>)</p>

# Install
## go Install

First, ensure that you have a compatible version of [Go](https://go.dev/) installed and set up on your system. The minimum required version of Go can be found in the [go.mod](go.mod) file.

To install Jig globally, execute the following command:

```shell
go install github.com/happyagosmith/jig@latest
```

<p align="right">(<a href="#readme-top">back to top</a>)</p>

# Getting Started
## Running Jig from Command Line
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
<p align="right">(<a href="#readme-top">back to top</a>)</p>

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
    jig-test:
      - issueTracker: JIRA
        issueKey: JIRA-123
        issueSummary: this is a jira story
        issueCategory: CLOSED_FEATURE
        issueDetail:
          extractedCategory: CLOSED_FEATURE
          issueKey: JIRA-123
          issueSummary: this is a jira story
          issueType: Story
          issueStatus: GOLIVE
          webURL: //browse/JIRA-123
        repoDetail:
          id: commit1
          shortId: short_commit1
          title: '[JIRA-123] custom pattern'
          message: |
            [JIRA-123] custom pattern
          createdAt: 2024-03-11T14:09:13.026Z
          webURL: http://gitlab.example.com/my/repo/commit/1
          origin: commit
          parsedSummary: custom pattern
          parsedKey: JIRA-123
          parsedIssueTracker: JIRA
          parser: customParser
      - issueTracker: GIT
        issueKey: "1"
        issueSummary: this is the gitlab issue title
        issueCategory: CLOSED_FEATURE
        issueDetail:
          extractedCategory: CLOSED_FEATURE
          issueKey: "1"
          issueSummary: this is the gitlab issue title
          issueType: issue
          issueStatus: closed
          webURL: https://gitlab.com/happyagosmith/jig-demo/-/issues/1
        repoDetail:
          id: commit4
          shortId: short_commit4
          title: 'feat: closing pattern'
          message: |-
            feat: closing pattern
             closes #1 gitlab issue
          createdAt: 2024-03-11T14:09:13.026Z
          origin: commit
          parsedCategory: FEATURE
          parsedKey: "1"
          parsedIssueTracker: GIT
          parser: closingPattern
          parsedType: close
      - issueTracker: SILK
        issueKey: SILK-222
        issueSummary: conventional commit
        issueCategory: CLOSED_FEATURE
        repoDetail:
          id: commit3
          shortId: short_commit3
          title: 'feat(SILK-222)!: conventional commit'
          message: "feat(SILK-222)!: conventional commit \n issue tracker not implemented"
          createdAt: 2024-03-11T14:09:13.026Z
          webURL: http://gitlab.example.com/my/repo/commit/3
          origin: commit
          parsedSummary: conventional commit
          parsedCategory: FEATURE
          parsedKey: SILK-222
          parsedIssueTracker: SILK
          parser: conventionalParser
          parsedType: feat
          isBreakingChange: true
  bugs:
    jig-test:
      - issueTracker: JIRA
        issueKey: JIRA-456
        issueSummary: this is a jira bug
        issueCategory: FIXED_BUG
        issueDetail:
          extractedCategory: FIXED_BUG
          issueKey: JIRA-456
          issueSummary: this is a jira bug
          issueType: BUG
          issueStatus: FIXED
          webURL: //browse/JIRA-456
        repoDetail:
          id: commit2
          shortId: short_commit2
          title: 'fix(j_JIRA-456): conventional commit bug fixed'
          message: "fix(j_JIRA-456): conventional commit bug fixed \n"
          createdAt: 2024-03-11T14:09:13.026Z
          webURL: http://gitlab.example.com/my/repo/commit/2
          origin: commit
          parsedSummary: conventional commit bug fixed
          parsedCategory: BUG_FIX
          parsedKey: JIRA-456
          parsedIssueTracker: JIRA
          parser: conventionalParser
          parsedType: fix
      - issueTracker: GIT
        issueKey: "2"
        issueSummary: this is the gitlab issue 2 title
        issueCategory: FIXED_BUG
        issueDetail:
          extractedCategory: FIXED_BUG
          issueKey: "2"
          issueSummary: this is the gitlab issue 2 title
          issueType: issue
          issueStatus: closed
          webURL: https://gitlab.com/happyagosmith/jig-demo/-/issues/2
        repoDetail:
          id: "10"
          shortId: "1"
          title: this is merge request 1
          message: 'this is a merge request description that fixes #2'
          webURL: http://gitlab.example.com/my/repo/merge_request/1
          origin: merge_request
          parsedCategory: BUG_FIX
          parsedKey: "2"
          parsedIssueTracker: GIT
          parser: closingPattern
          parsedType: fix
  knownIssues: {}
  breakingChange:
    jig-test:
      - issueTracker: SILK
        issueKey: SILK-222
        issueSummary: conventional commit
        issueCategory: CLOSED_FEATURE
        repoDetail:
          id: commit3
          shortId: short_commit3
          title: 'feat(SILK-222)!: conventional commit'
          message: "feat(SILK-222)!: conventional commit \n issue tracker not implemented"
          createdAt: 2024-03-11T14:09:13.026Z
          webURL: http://gitlab.example.com/my/repo/commit/3
          origin: commit
          parsedSummary: conventional commit
          parsedCategory: FEATURE
          parsedKey: SILK-222
          parsedIssueTracker: SILK
          parser: conventionalParser
          parsedType: feat
          isBreakingChange: true
  gitRepos:
    - label: jig-test
      gitRepoID: "123"
      previousVersion: 0.0.1
      version: 0.0.2
      checkVersion: '@testdata/version.yaml:$.versions.a'
      extractedKeys:
        - id: commit1
          shortId: short_commit1
          title: '[JIRA-123] custom pattern'
          message: |
            [JIRA-123] custom pattern
          createdAt: 2024-03-11T14:09:13.026Z
          webURL: http://gitlab.example.com/my/repo/commit/1
          origin: commit
          parsedSummary: custom pattern
          parsedKey: JIRA-123
          parsedIssueTracker: JIRA
          parser: customParser
        - id: commit2
          shortId: short_commit2
          title: 'fix(j_JIRA-456): conventional commit bug fixed'
          message: "fix(j_JIRA-456): conventional commit bug fixed \n"
          createdAt: 2024-03-11T14:09:13.026Z
          webURL: http://gitlab.example.com/my/repo/commit/2
          origin: commit
          parsedSummary: conventional commit bug fixed
          parsedCategory: BUG_FIX
          parsedKey: JIRA-456
          parsedIssueTracker: JIRA
          parser: conventionalParser
          parsedType: fix
        - id: commit3
          shortId: short_commit3
          title: 'feat(SILK-222)!: conventional commit'
          message: "feat(SILK-222)!: conventional commit \n issue tracker not implemented"
          createdAt: 2024-03-11T14:09:13.026Z
          webURL: http://gitlab.example.com/my/repo/commit/3
          origin: commit
          parsedSummary: conventional commit
          parsedCategory: FEATURE
          parsedKey: SILK-222
          parsedIssueTracker: SILK
          parser: conventionalParser
          parsedType: feat
          isBreakingChange: true
        - id: commit4
          shortId: short_commit4
          title: 'feat(%222)!: conventional commit'
          message: "feat(%222)!: conventional commit \n unknown issue tracker"
          createdAt: 2024-03-11T14:09:13.026Z
          webURL: http://gitlab.example.com/my/repo/commit/4
          origin: commit
          parsedSummary: conventional commit
          parsedCategory: FEATURE
          parsedKey: '%222'
          parsedIssueTracker: NONE
          parser: conventionalParser
          parsedType: feat
          isBreakingChange: true
        - id: commit4
          shortId: short_commit4
          title: 'feat: closing pattern'
          message: |-
            feat: closing pattern
             closes #1 gitlab issue
          createdAt: 2024-03-11T14:09:13.026Z
          origin: commit
          parsedCategory: FEATURE
          parsedKey: "1"
          parsedIssueTracker: GIT
          parser: closingPattern
          parsedType: close
        - id: "10"
          shortId: "1"
          title: this is merge request 1
          message: 'this is a merge request description that fixes #2'
          webURL: http://gitlab.example.com/my/repo/merge_request/1
          origin: merge_request
          parsedCategory: BUG_FIX
          parsedKey: "2"
          parsedIssueTracker: GIT
          parser: closingPattern
          parsedType: fix
      hasNewFeature: true
      hasBugFixed: true
```

### release note template

Jig employs the [go-template text](https://pkg.go.dev/text/template]) syntax to structure the release note. You can find an [example](examples/rn.tpl) that illustrates how to integrate the generated fields into the go-template.

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

##### Using Sprig Functions in Go Text Templates
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

<p align="right">(<a href="#readme-top">back to top</a>)</p>

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
