package glance

import (
	"context"
	"fmt"
	"html/template"
	"net/http"
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
		widget.Sort = "updated"
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
	Name        string `json:"name"`
	FullName    string `json:"full_name"`
	Description string `json:"description"`
	Stars       int    `json:"stargazers_count"`
	Language    string `json:"language"`
	UpdatedAt   string `json:"updated_at"`
	HTMLURL     string `json:"html_url"`
}

func fetchUserRepositoriesFromGithub(token string, sort string) ([]githubRepo, error) {
	perPage := 30

	url := fmt.Sprintf("https://api.github.com/user/repos?visibility=all&sort=%s&per_page=%d", sort, perPage)
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
			LastCommit:  parseRFC3339Time(r.UpdatedAt),
			URL:         r.HTMLURL,
		})
	}

	return result, nil
}
