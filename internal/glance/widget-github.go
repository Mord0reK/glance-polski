package glance

import (
	"context"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

var githubWidgetTemplate = mustParseTemplate("github.html", "widget-base.html")

type githubWidget struct {
	widgetBase    `yaml:",inline"`
	Token         string       `yaml:"token"`
	CollapseAfter int          `yaml:"collapse-after"`
	Sort          string       `yaml:"sort"`
	TitleLink     string       `yaml:"title-link"`
	Repositories  []githubRepo `yaml:"-"`
}

type githubRepo struct {
	Name        string
	FullName    string
	Description string
	Stars       int
	Language    string
	LastCommit  time.Time
	URL         string
	Visibility  string
}

func (widget *githubWidget) initialize() error {
	widget.withTitle("GitHub").withCacheDuration(15 * time.Minute)

	if widget.TitleLink != "" {
		widget.TitleURL = widget.TitleLink
	}

	if widget.CollapseAfter == 0 {
		widget.CollapseAfter = 5
	}

	if widget.Sort == "" {
		widget.Sort = "pushed"
	}

	return nil
}

func (widget *githubWidget) update(ctx context.Context) {
	repos, err := fetchUserRepositoriesFromGithub(
		widget.Token,
		widget.Sort,
	)

	if !widget.canContinueUpdateAfterHandlingErr(err) {
		return
	}

	widget.Repositories = repos
}

func (widget *githubWidget) Render() template.HTML {
	return widget.renderTemplate(widget, githubWidgetTemplate)
}

type githubUserRepoResponseJson struct {
	Name          string `json:"name"`
	FullName      string `json:"full_name"`
	Description   string `json:"description"`
	Stars         int    `json:"stargazers_count"`
	Language      string `json:"language"`
	UpdatedAt     string `json:"updated_at"`
	PushedAt      string `json:"pushed_at"`
	DefaultBranch string `json:"default_branch"`
	HTMLURL       string `json:"html_url"`
	Visibility    string `json:"visibility"`
}

func githubRepoLastActivity(repo githubUserRepoResponseJson) time.Time {
	// pushed_at is a fallback for repositories where the latest default-branch
	// commit cannot be fetched. updated_at only changes for repository metadata,
	// so it must not be preferred over pushed_at.
	if repo.PushedAt != "" {
		if pushedAt, err := time.Parse(time.RFC3339, repo.PushedAt); err == nil {
			return pushedAt
		}
	}

	return parseRFC3339Time(repo.UpdatedAt)
}

type githubCommitDateResponseJson struct {
	Commit struct {
		Committer struct {
			Date string `json:"date"`
		} `json:"committer"`
	} `json:"commit"`
}

func fetchGithubRepoLastCommit(token string, repo githubUserRepoResponseJson) (time.Time, error) {
	parts := strings.SplitN(repo.FullName, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" || repo.DefaultBranch == "" {
		return time.Time{}, fmt.Errorf("missing repository name or default branch")
	}

	requestURL := fmt.Sprintf(
		"https://api.github.com/repos/%s/%s/commits?sha=%s&per_page=1",
		url.PathEscape(parts[0]),
		url.PathEscape(parts[1]),
		url.QueryEscape(repo.DefaultBranch),
	)
	req, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		return time.Time{}, err
	}

	if token != "" {
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
	}
	req.Header.Add("Accept", "application/vnd.github.v3+json")

	commits, err := decodeJsonFromRequest[[]githubCommitDateResponseJson](defaultHTTPClient, req)
	if err != nil {
		return time.Time{}, err
	}
	if len(commits) == 0 || commits[0].Commit.Committer.Date == "" {
		return time.Time{}, fmt.Errorf("no commits found")
	}

	return parseRFC3339Time(commits[0].Commit.Committer.Date), nil
}

func fetchUserRepositoriesFromGithub(token string, sortBy string) ([]githubRepo, error) {
	perPage := 30

	url := fmt.Sprintf("https://api.github.com/user/repos?visibility=all&sort=%s&per_page=%d", sortBy, perPage)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: could not create request: %v", errNoContent, err)
	}

	if token != "" {
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
	}
	req.Header.Add("Accept", "application/vnd.github.v3+json")

	var repos []githubUserRepoResponseJson
	repos, err = decodeJsonFromRequest[[]githubUserRepoResponseJson](defaultHTTPClient, req)
	if err != nil {
		return nil, fmt.Errorf("%w: could not get repositories: %s", errNoContent, err)
	}

	result := make([]githubRepo, 0, len(repos))
	for _, r := range repos {
		result = append(result, githubRepo{
			Name:        r.Name,
			FullName:    r.FullName,
			Description: r.Description,
			Stars:       r.Stars,
			Language:    r.Language,
			LastCommit:  githubRepoLastActivity(r),
			URL:         r.HTMLURL,
			Visibility:  r.Visibility,
		})
	}

	job := newJob(func(repo githubUserRepoResponseJson) (time.Time, error) {
		return fetchGithubRepoLastCommit(token, repo)
	}, repos).withWorkers(10)
	commitTimes, commitErrors, jobErr := workerPoolDo(job)
	if jobErr == nil {
		for i := range result {
			if commitErrors[i] == nil && !commitTimes[i].IsZero() {
				result[i].LastCommit = commitTimes[i]
			}
		}
	}

	if sortBy == "pushed" {
		sortGithubReposByLastCommit(result)
	}

	return result, nil
}

func sortGithubReposByLastCommit(repos []githubRepo) {
	sort.SliceStable(repos, func(i, j int) bool {
		return repos[i].LastCommit.After(repos[j].LastCommit)
	})
}
