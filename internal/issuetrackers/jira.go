package issuetrackers

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	v2 "github.com/ctreminiom/go-atlassian/jira/v3"
	"github.com/ctreminiom/go-atlassian/pkg/infra/models"
	"github.com/happyagosmith/jig/internal/entities"
)

type jiraFilter struct {
	issueType   string
	issueStatus string
}

type Jira struct {
	client               *v2.Client
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
	client, err := v2.New(nil, URL)
	if err != nil {
		return Jira{}, err
	}

	client.Auth.SetBasicAuth(username, password)

	j := Jira{client: client}
	for _, o := range opts {
		o(&j)
	}

	return j, nil
}

// searchIssuesRaw calls the new /rest/api/3/search/jql endpoint using raw API call
func (j Jira) searchIssuesRaw(ctx context.Context, jql string, startAt, maxResults int) (*models.IssueSearchScheme, *models.ResponseScheme, error) {
	// Build the query parameters
	params := url.Values{}
	params.Add("jql", jql)
	params.Add("startAt", fmt.Sprintf("%d", startAt))
	params.Add("maxResults", fmt.Sprintf("%d", maxResults))

	apiEndpoint := fmt.Sprintf("rest/api/3/search/jql?%s", params.Encode())

	request, err := j.client.NewRequest(ctx, http.MethodGet, apiEndpoint, "", nil)
	if err != nil {
		return nil, nil, err
	}

	searchResult := new(models.IssueSearchScheme)
	response, err := j.client.Call(request, searchResult)
	if err != nil {
		return nil, response, err
	}

	return searchResult, response, nil
}

func (j Jira) GetIssues(ctx context.Context, repo *entities.Repo, keys []string) ([]entities.Issue, error) {
	if len(keys) == 0 {
		return []entities.Issue{}, nil
	}

	jql := fmt.Sprintf("issue in (%s)", strings.Join(keys, ","))
	fmt.Printf("retrieving issues info using JQL \"%s\"\n", jql)

	searchResult, response, err := j.searchIssuesRaw(ctx, jql, 0, 1000)
	if err != nil {
		if response != nil {
			fmt.Printf("Error response from Jira: endpoint=%s, status=%d\n", response.Endpoint, response.Code)
		}
		return nil, fmt.Errorf("failed to search issues: %w", err)
	}

	issues := make([]entities.Issue, 0, len(searchResult.Issues))
	isPresent := map[string]bool{}

	subTaskParents := []string{}
	for _, issue := range searchResult.Issues {
		if isPresent[issue.Key] {
			continue
		}
		isPresent[issue.Key] = true
		parent := issue.Fields.Parent
		if issue.Fields.IssueType.Subtask && parent != nil && parent.Key != "" && !isPresent[parent.Key] {
			fmt.Printf("issue %s is a subtask. add parent key %s instead\n", issue.Key, parent.Key)
			subTaskParents = append(subTaskParents, parent.Key)
			continue
		}

		issues = append(issues, entities.Issue{
			Category:     j.extractIssueCategory(issue),
			IssueKey:     issue.Key,
			IssueSummary: issue.Fields.Summary,
			IssueType:    issue.Fields.IssueType.Name,
			IssueStatus:  issue.Fields.Status.Name,
			WebURL:       fmt.Sprintf("/browse/%s", issue.Key),
		})
	}
	if len(subTaskParents) == 0 {
		return issues, nil
	}

	jql = fmt.Sprintf("issue in (%s)", strings.Join(subTaskParents, ","))
	fmt.Printf("retrieving issue parents info using JQL \"%s\"\n", jql)

	parentSearchResult, response, err := j.searchIssuesRaw(ctx, jql, 0, 1000)
	if err != nil {
		if response != nil {
			fmt.Printf("Error response from Jira: endpoint=%s, status=%d\n", response.Endpoint, response.Code)
		}
		return nil, fmt.Errorf("failed to search parent issues: %w", err)
	}

	for _, issue := range parentSearchResult.Issues {
		if isPresent[issue.Key] {
			continue
		}
		isPresent[issue.Key] = true
		issues = append(issues, entities.Issue{
			Category:     j.extractIssueCategory(issue),
			IssueKey:     issue.Key,
			IssueSummary: issue.Fields.Summary,
			IssueType:    issue.Fields.IssueType.Name,
			IssueStatus:  issue.Fields.Status.Name,
			WebURL:       fmt.Sprintf("/browse/%s", issue.Key),
		})
	}
	return issues, nil
}

func (j Jira) extractIssueCategory(issue *models.IssueScheme) entities.IssueCategory {
	if issue.Fields.IssueType.Subtask {
		return entities.SUB_TASK
	}
	issueType := strings.ToUpper(issue.Fields.IssueType.Name)
	issueStatus := strings.ToUpper(issue.Fields.Status.Name)
	for _, jf := range j.closedFeatureFilters {
		if issueType == jf.issueType && issueStatus == jf.issueStatus {
			return entities.CLOSED_FEATURE
		}
	}
	for _, jf := range j.fixedBugFilters {
		if issueType == jf.issueType && issueStatus == jf.issueStatus {
			return entities.FIXED_BUG
		}
	}

	return entities.OTHER
}

func (j Jira) GetKnownIssues(ctx context.Context, repo *entities.Repo) ([]entities.Issue, error) {
	if repo.Project == "" {
		return nil, nil
	}
	component := repo.Component
	project := repo.Project

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

	searchResult, response, err := j.searchIssuesRaw(ctx, jql, 0, 1000)
	if err != nil {
		if response != nil {
			fmt.Printf("Error response from Jira: endpoint=%s, status=%d\n", response.Endpoint, response.Code)
		}
		return nil, fmt.Errorf("failed to search known issues: %w", err)
	}

	issues := make([]entities.Issue, 0, len(searchResult.Issues))
	for _, issue := range searchResult.Issues {
		issues = append(issues, entities.Issue{
			Category:     j.extractIssueCategory(issue),
			IssueKey:     issue.Key,
			IssueSummary: issue.Fields.Summary,
			IssueType:    issue.Fields.IssueType.Name,
			IssueStatus:  issue.Fields.Status.Name,
			WebURL:       fmt.Sprintf("/browse/%s", issue.Key),
		})
	}
	return issues, nil
}
