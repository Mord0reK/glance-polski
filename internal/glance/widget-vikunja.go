package glance

import (
	"context"
	"fmt"
	"html/template"
	"net/http"
	"sort"
	"time"
)

var vikunjaWidgetTemplate = mustParseTemplate("vikunja.html", "widget-base.html")

type vikunjaWidget struct {
	widgetBase `yaml:",inline"`
	URL        string `yaml:"url"`
	Token      string `yaml:"token"`
	Limit      int    `yaml:"limit"`
	Tasks      []vikunjaTask
}

type vikunjaTask struct {
	Title       string
	DueDate     time.Time
	Done        bool
	PercentDone int
	Labels      []vikunjaLabel
	TimeLeft    string
	DueDateStr  string
}

type vikunjaLabel struct {
	Title string
	Color string
}

type vikunjaAPITask struct {
	ID          int               `json:"id"`
	Title       string            `json:"title"`
	Done        bool              `json:"done"`
	DueDate     string            `json:"due_date"`
	PercentDone float64           `json:"percent_done"`
	Labels      []vikunjaAPILabel `json:"labels"`
}

type vikunjaAPILabel struct {
	Title    string `json:"title"`
	HexColor string `json:"hex_color"`
}

func (widget *vikunjaWidget) initialize() error {
	widget.withTitle("Vikunja").withCacheDuration(5 * time.Minute)

	if widget.URL == "" {
		return fmt.Errorf("URL is required")
	}

	if widget.Token == "" {
		return fmt.Errorf("token is required")
	}

	if widget.Limit <= 0 {
		widget.Limit = 10
	}

	return nil
}

func (widget *vikunjaWidget) update(ctx context.Context) {
	tasks, err := widget.fetchTasks()

	if !widget.canContinueUpdateAfterHandlingErr(err) {
		return
	}

	widget.Tasks = tasks
}

func (widget *vikunjaWidget) fetchTasks() ([]vikunjaTask, error) {
	url := widget.URL + "/api/v1/tasks/all"

	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	request.Header.Set("Authorization", "Bearer "+widget.Token)

	apiTasks, err := decodeJsonFromRequest[[]vikunjaAPITask](defaultHTTPClient, request)
	if err != nil {
		return nil, err
	}

	tasks := make([]vikunjaTask, 0)
	now := time.Now()

	for _, apiTask := range apiTasks {
		if apiTask.Done {
			continue
		}

		task := vikunjaTask{
			Title:       apiTask.Title,
			Done:        apiTask.Done,
			PercentDone: int(apiTask.PercentDone),
		}

		if apiTask.DueDate != "" {
			dueDate, err := time.Parse(time.RFC3339, apiTask.DueDate)
			if err == nil {
				task.DueDate = dueDate
				task.DueDateStr = dueDate.Format("2006-01-02 15:04")
				task.TimeLeft = formatTimeLeft(now, dueDate)
			}
		}

		task.Labels = make([]vikunjaLabel, len(apiTask.Labels))
		for i, label := range apiTask.Labels {
			color := label.HexColor
			// Ensure the color has # prefix
			if color != "" && color[0] != '#' {
				color = "#" + color
			}
			task.Labels[i] = vikunjaLabel{
				Title: label.Title,
				Color: color,
			}
		}

		tasks = append(tasks, task)
	}

	// Sortowanie zadań po dacie - zadania bez daty na końcu
	sort.Slice(tasks, func(i, j int) bool {
		// Jeśli oba zadania nie mają daty, zachowaj kolejność
		if tasks[i].DueDate.IsZero() && tasks[j].DueDate.IsZero() {
			return false
		}
		// Zadania bez daty idą na koniec
		if tasks[i].DueDate.IsZero() {
			return false
		}
		if tasks[j].DueDate.IsZero() {
			return true
		}
		// Sortuj po dacie rosnąco (wcześniejsze daty pierwsze)
		return tasks[i].DueDate.Before(tasks[j].DueDate)
	})

	// Obetnij do limitu po posortowaniu
	if len(tasks) > widget.Limit {
		tasks = tasks[:widget.Limit]
	}

	return tasks, nil
}

func formatTimeLeft(now, dueDate time.Time) string {
	if dueDate.IsZero() {
		return "-"
	}

	duration := dueDate.Sub(now)

	if duration < 0 {
		duration = -duration
		days := int(duration.Hours() / 24)
		hours := int(duration.Hours()) % 24

		if days > 0 {
			return fmt.Sprintf("-%d dni %d godz.", days, hours)
		}
		if hours > 0 {
			return fmt.Sprintf("-%d godz.", hours)
		}
		return fmt.Sprintf("-%d min.", int(duration.Minutes()))
	}

	days := int(duration.Hours() / 24)
	hours := int(duration.Hours()) % 24

	if days > 0 {
		return fmt.Sprintf("%d dni %d godz.", days, hours)
	}
	if hours > 0 {
		return fmt.Sprintf("%d godz.", hours)
	}
	return fmt.Sprintf("%d min.", int(duration.Minutes()))
}

func (widget *vikunjaWidget) Render() template.HTML {
	return widget.renderTemplate(widget, vikunjaWidgetTemplate)
}
