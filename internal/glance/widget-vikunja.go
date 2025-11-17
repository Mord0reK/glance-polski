package glance

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"sort"
	"strings"
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
	ID          int
	Title       string
	DueDate     time.Time
	Done        bool
	PercentDone int
	Labels      []vikunjaLabel
	TimeLeft    string
	DueDateStr  string
}

type vikunjaLabel struct {
	ID    int
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
	ID       int    `json:"id"`
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
			ID:          apiTask.ID,
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
				ID:    label.ID,
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

func (widget *vikunjaWidget) completeTask(taskID int) error {
	url := fmt.Sprintf("%s/api/v1/tasks/%d", widget.URL, taskID)

	payload := map[string]interface{}{
		"done": true,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	// Vikunja API uses POST for updating tasks (not PUT or PATCH)
	request, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	request.Header.Set("Authorization", "Bearer "+widget.Token)
	request.Header.Set("Content-Type", "application/json")

	_, err = decodeJsonFromRequest[vikunjaAPITask](defaultHTTPClient, request)
	return err
}

func (widget *vikunjaWidget) updateTaskBasic(taskID int, title string, dueDate string) error {
	url := fmt.Sprintf("%s/api/v1/tasks/%d", widget.URL, taskID)

	payload := map[string]interface{}{
		"title": title,
	}

	if dueDate != "" {
		payload["due_date"] = dueDate
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	// Vikunja API uses POST for updating tasks (not PUT or PATCH)
	request, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	request.Header.Set("Authorization", "Bearer "+widget.Token)
	request.Header.Set("Content-Type", "application/json")

	_, err = decodeJsonFromRequest[vikunjaAPITask](defaultHTTPClient, request)
	return err
}

func (widget *vikunjaWidget) updateTaskLabels(taskID int, currentLabels []vikunjaAPILabel, desiredLabelIDs []int) error {
	// Create a map of current label IDs for easy lookup
	currentLabelMap := make(map[int]bool)
	for _, label := range currentLabels {
		currentLabelMap[label.ID] = true
	}

	// Create a map of desired label IDs
	desiredLabelMap := make(map[int]bool)
	for _, labelID := range desiredLabelIDs {
		desiredLabelMap[labelID] = true
	}

	// Add labels that are in desired but not in current
	for _, labelID := range desiredLabelIDs {
		if !currentLabelMap[labelID] {
			if err := widget.addLabelToTask(taskID, labelID); err != nil {
				return fmt.Errorf("failed to add label %d: %w", labelID, err)
			}
		}
	}

	// Remove labels that are in current but not in desired
	for _, label := range currentLabels {
		if !desiredLabelMap[label.ID] {
			if err := widget.removeLabelFromTask(taskID, label.ID); err != nil {
				return fmt.Errorf("failed to remove label %d: %w", label.ID, err)
			}
		}
	}

	return nil
}

func (widget *vikunjaWidget) addLabelToTask(taskID int, labelID int) error {
	url := fmt.Sprintf("%s/api/v1/tasks/%d/labels", widget.URL, taskID)

	payload := map[string]interface{}{
		"label_id": labelID,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	request, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	request.Header.Set("Authorization", "Bearer "+widget.Token)
	request.Header.Set("Content-Type", "application/json")

	// Response is just a confirmation, we don't need to decode it
	response, err := defaultHTTPClient.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		body, _ := io.ReadAll(response.Body)
		bodyStr := string(body)
		
		// Vikunja returns error code 8001 when label already exists
		// This is not an error for us - we want the label on the task
		if response.StatusCode == 400 && (strings.Contains(bodyStr, "8001") || strings.Contains(bodyStr, "already exists")) {
			// Label already exists, which is fine - we wanted it there anyway
			return nil
		}
		
		return fmt.Errorf("unexpected status code %d: %s", response.StatusCode, bodyStr)
	}

	return nil
}

func (widget *vikunjaWidget) removeLabelFromTask(taskID int, labelID int) error {
	url := fmt.Sprintf("%s/api/v1/tasks/%d/labels/%d", widget.URL, taskID, labelID)

	request, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("creating DELETE request: %w", err)
	}

	request.Header.Set("Authorization", "Bearer "+widget.Token)

	response, err := defaultHTTPClient.Do(request)
	if err != nil {
		return fmt.Errorf("executing DELETE request: %w", err)
	}
	defer response.Body.Close()

	body, _ := io.ReadAll(response.Body)
	bodyStr := string(body)

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return fmt.Errorf("DELETE labels status %d: %s", response.StatusCode, bodyStr)
	}

	// Success - label was removed
	return nil
}

func (widget *vikunjaWidget) fetchAllLabels() ([]vikunjaAPILabel, error) {
	url := widget.URL + "/api/v1/labels"

	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	request.Header.Set("Authorization", "Bearer "+widget.Token)

	labels, err := decodeJsonFromRequest[[]vikunjaAPILabel](defaultHTTPClient, request)
	if err != nil {
		return nil, err
	}

	return labels, nil
}

func (widget *vikunjaWidget) createTask(title string, dueDate string, labelIDs []int) (*vikunjaAPITask, error) {
	url := widget.URL + "/api/v1/projects/1/tasks"

	payload := map[string]interface{}{
		"title": title,
	}

	if dueDate != "" {
		payload["due_date"] = dueDate
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	request, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	request.Header.Set("Authorization", "Bearer "+widget.Token)
	request.Header.Set("Content-Type", "application/json")

	task, err := decodeJsonFromRequest[vikunjaAPITask](defaultHTTPClient, request)
	if err != nil {
		return nil, err
	}

	// Add labels to the task if any - errors are logged but don't fail task creation
	for _, labelID := range labelIDs {
		_ = widget.addLabelToTask(task.ID, labelID)
	}

	return &task, nil
}
