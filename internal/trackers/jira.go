package trackers

import (
	"fmt"
	"strings"

	"github.com/andygrunwald/go-jira"
)

type jiraFilter struct {
	issueType   string
	issueStatus string
}

type Jira struct {
	client               *jira.Client
	closedFeatureFilters []jiraFilter
	fixedBugFilters      []jiraFilter
	jqlKnownIssue        string
}

type JiraOpt func(*Jira)

func WithFixedBugFilter(issueType, issueStatus string) JiraOpt {
	return func(j *Jira) {
		j.fixedBugFilters = append(j.fixedBugFilters, jiraFilter{issueType: strings.ToUpper(issueType), issueStatus: strings.ToUpper(issueStatus)})
	}
}

func WithClosedFeatureFilter(issueType, issueStatus string) JiraOpt {
	return func(j *Jira) {
		j.closedFeatureFilters = append(j.closedFeatureFilters, jiraFilter{issueType: strings.ToUpper(issueType), issueStatus: strings.ToUpper(issueStatus)})
	}
}

func WithKnownIssueJql(jql string) JiraOpt {
	return func(j *Jira) {
		j.jqlKnownIssue = jql
	}
}

func NewJira(URL, username, password string, opts ...JiraOpt) (Jira, error) {
	tp := jira.BasicAuthTransport{
		Username: username,
		Password: password,
	}

	client, err := jira.NewClient(tp.Client(), URL)
	if err != nil {
		return Jira{}, err
	}

	j := Jira{client: client}
	for _, o := range opts {
		o(&j)
	}

	return j, nil
}

func (j Jira) GetIssues(keys []string) ([]IssueDetail, error) {
	if len(keys) == 0 {
		return []IssueDetail{}, nil
	}

	jql := fmt.Sprintf("issue in (%s)", strings.Join(keys, ","))
	opt := &jira.SearchOptions{
		MaxResults: 1000,
		StartAt:    0,
	}
	fmt.Printf("retrieving issues info using JQL \"%s\"\n", jql)
	jissues, _, err := j.client.Issue.Search(jql, opt)
	if err != nil {
		return nil, err
	}

	issues := make([]IssueDetail, 0, len(jissues))
	isPresent := map[string]bool{}

	subTaskParents := []string{}
	for _, issue := range jissues {
		if isPresent[issue.Key] {
			continue
		}
		isPresent[issue.Key] = true
		parent := issue.Fields.Parent
		if issue.Fields.Type.Subtask && parent != nil && parent.Key != "" && !isPresent[parent.Key] {
			fmt.Printf("issue %s is a subtask. add parent key %s instead\n", issue.Key, parent.Key)
			subTaskParents = append(subTaskParents, parent.Key)
			continue
		}
		issues = append(issues, IssueDetail{
			Category:     j.extractIssueCategory(issue),
			IssueKey:     issue.Key,
			IssueSummary: issue.Fields.Summary,
			IssueType:    issue.Fields.Type.Name,
			IssueStatus:  issue.Fields.Status.Name,
		})
	}
	if len(subTaskParents) == 0 {
		return issues, nil
	}

	jql = fmt.Sprintf("issue in (%s)", strings.Join(subTaskParents, ","))
	fmt.Printf("retrieving issue parents info using JQL \"%s\"\n", jql)
	pIssues, _, err := j.client.Issue.Search(jql, opt)
	if err != nil {
		return nil, err
	}

	for _, issue := range pIssues {
		if isPresent[issue.Key] {
			continue
		}
		isPresent[issue.Key] = true
		issues = append(issues, IssueDetail{
			Category:     j.extractIssueCategory(issue),
			IssueKey:     issue.Key,
			IssueSummary: issue.Fields.Summary,
			IssueType:    issue.Fields.Type.Name,
			IssueStatus:  issue.Fields.Status.Name,
		})
	}
	return issues, nil
}

func (j Jira) extractIssueCategory(ji jira.Issue) CategoryType {
	if ji.Fields.Type.Subtask {
		return SUB_TASK
	}
	issueType := strings.ToUpper(ji.Fields.Type.Name)
	issueStatus := strings.ToUpper(ji.Fields.Status.Name)
	for _, jf := range j.closedFeatureFilters {
		if issueType == jf.issueType && issueStatus == jf.issueStatus {
			return CLOSED_FEATURE
		}
	}
	for _, jf := range j.fixedBugFilters {
		if issueType == jf.issueType && issueStatus == jf.issueStatus {
			return FIXED_BUG
		}
	}

	return OTHER
}

func (j Jira) GetKnownIssues(project, component string) ([]IssueDetail, error) {
	opt := &jira.SearchOptions{
		MaxResults: 1000,
		StartAt:    0,
	}
	jqls := []string{}
	if j.jqlKnownIssue != "" {
		jqls = append(jqls, j.jqlKnownIssue)
	}

	if project != "" {
		jqls = append(jqls, fmt.Sprintf("project = \"%s\"", project))
	}

	if component != "" {
		jqls = append(jqls, fmt.Sprintf("component = \"%s\"", component))
	}

	if len(jqls) == 0 {
		return nil, nil
	}

	jql := strings.Join(jqls, " and ")
	fmt.Printf("\nretrieving known issues using Jira jql \"%s\"\n", jql)
	jIssues, _, err := j.client.Issue.Search(jql, opt)
	if err != nil {
		return nil, err
	}

	issues := make([]IssueDetail, 0, len(jIssues))
	for _, issue := range jIssues {
		issues = append(issues, IssueDetail{
			Category:     j.extractIssueCategory(issue),
			IssueKey:     issue.Key,
			IssueSummary: issue.Fields.Summary,
			IssueType:    issue.Fields.Type.Name,
			IssueStatus:  issue.Fields.Status.Name,
		})
	}
	return issues, nil
}
