package trackers

import (
	"fmt"
	"strings"

	"github.com/andygrunwald/go-jira"
	"github.com/happyagosmith/jig/internal/git"
	"github.com/happyagosmith/jig/internal/model"
	"github.com/happyagosmith/jig/internal/parsers"
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
	itp                  parsers.IssueTrackerType
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

	j := Jira{client: client, itp: parsers.JIRA}
	for _, o := range opts {
		o(&j)
	}

	return j, nil
}

func (j Jira) Type() parsers.IssueTrackerType {
	return parsers.JIRA
}

func (j Jira) GetIssues(gds []git.CommitDetail) ([]model.ExtractedIssue, error) {
	if len(gds) == 0 {
		return []model.ExtractedIssue{}, nil
	}

	keys := []string{}
	breakingChange := map[string]bool{}
	for _, gc := range gds {
		if gc.IssueTracker != parsers.JIRA {
			continue
		}

		keys = append(keys, gc.IssueKey)
		breakingChange[gc.IssueKey] = gc.IsBreaking
	}

	if len(keys) == 0 {
		return []model.ExtractedIssue{}, nil
	}

	jql := fmt.Sprintf("issue in (%s)", strings.Join(keys, ","))
	opt := &jira.SearchOptions{
		MaxResults: 1000,
		StartAt:    0,
	}
	fmt.Printf("\nRetrieving issues using Jira jql \"%s\"\n", jql)
	jissues, _, err := j.client.Issue.Search(jql, opt)
	if err != nil {
		return nil, err
	}

	issues := make([]model.ExtractedIssue, 0, len(jissues))

	subTaskParents := []string{}
	for _, issue := range jissues {
		parent := issue.Fields.Parent
		if issue.Fields.Type.Subtask && parent != nil && parent.Key != "" {
			subTaskParents = append(subTaskParents, parent.Key)
			breakingChange[parent.Key] = breakingChange[issue.Key]
			continue
		}
		issues = append(issues, model.ExtractedIssue{
			Category:         j.extractIssueCategory(issue),
			IssueKey:         issue.Key,
			IssueSummary:     issue.Fields.Summary,
			IssueType:        issue.Fields.Type.Name,
			IssueStatus:      issue.Fields.Status.Name,
			IssueTrackerType: j.itp,
			IsBreakingChange: breakingChange[issue.Key]})
	}
	if len(subTaskParents) == 0 {
		return issues, nil
	}

	jql = fmt.Sprintf("issue in (%s)", strings.Join(subTaskParents, ","))
	fmt.Printf("\nRetrieving subtask parents using Jira jql \"%s\"\n", jql)
	pIssues, _, err := j.client.Issue.Search(jql, opt)
	if err != nil {
		return nil, err
	}

	for _, issue := range pIssues {
		issues = append(issues, model.ExtractedIssue{
			Category:         j.extractIssueCategory(issue),
			IssueKey:         issue.Key,
			IssueSummary:     issue.Fields.Summary,
			IssueType:        issue.Fields.Type.Name,
			IssueStatus:      issue.Fields.Status.Name,
			IssueTrackerType: j.itp,
			IsBreakingChange: breakingChange[issue.Key]})
	}
	return issues, nil
}

func (j Jira) extractIssueCategory(ji jira.Issue) model.CategoryType {
	if ji.Fields.Type.Subtask {
		return model.SUB_TASK
	}
	issueType := strings.ToUpper(ji.Fields.Type.Name)
	issueStatus := strings.ToUpper(ji.Fields.Status.Name)
	for _, jf := range j.closedFeatureFilters {
		if issueType == jf.issueType && issueStatus == jf.issueStatus {
			return model.CLOSED_FEATURE
		}
	}
	for _, jf := range j.fixedBugFilters {
		if issueType == jf.issueType && issueStatus == jf.issueStatus {
			return model.FIXED_BUG
		}
	}

	return model.OTHER
}

func (j Jira) GetKnownIssues(project, component string) ([]model.ExtractedIssue, error) {
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
	fmt.Printf("\nRetrieving known issues using Jira jql \"%s\"\n", jql)
	jIssues, _, err := j.client.Issue.Search(jql, opt)
	if err != nil {
		return nil, err
	}

	issues := make([]model.ExtractedIssue, 0, len(jIssues))
	for _, issue := range jIssues {
		issues = append(issues, model.ExtractedIssue{
			Category:         j.extractIssueCategory(issue),
			IssueKey:         issue.Key,
			IssueSummary:     issue.Fields.Summary,
			IssueType:        issue.Fields.Type.Name,
			IssueStatus:      issue.Fields.Status.Name,
			IssueTrackerType: j.itp})
	}
	return issues, nil
}
