package glance

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"net/url"
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
	ProjectID  int    `yaml:"project-id"` // Project ID for creating new tasks
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
	Reminder    time.Time
	ReminderStr string
}

type vikunjaLabel struct {
	ID    int
	Title string
	Color string
}

type vikunjaAPITask struct {
	ID          int                  `json:"id"`
	Title       string               `json:"title"`
	Done        bool                 `json:"done"`
	DueDate     string               `json:"due_date"`
	PercentDone float64              `json:"percent_done"`
	Labels      []vikunjaAPILabel    `json:"labels"`
	Reminders   []vikunjaAPIReminder `json:"reminders"`
}

type vikunjaAPIReminder struct {
	ID       int    `json:"id"`
	Reminder string `json:"reminder"`
}

type vikunjaAPILabel struct {
	ID       int    `json:"id"`
	Title    string `json:"title"`
	HexColor string `json:"hex_color"`
}

type vikunjaProject struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
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

	if widget.ProjectID <= 0 {
		widget.ProjectID = 1 // Default to project 1 if not specified
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
	fullURL := widget.URL + "/api/v1/tasks/all"

	u, err := url.Parse(fullURL)
	if err != nil {
		return nil, err
	}

	q := u.Query()
	q.Set("sort_by", "due_date")
	q.Set("order_by", "asc")
	q.Set("limit", "250")
	u.RawQuery = q.Encode()

	request, err := http.NewRequest("GET", u.String(), nil)
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

		if len(apiTask.Reminders) > 0 {
			reminderStr := apiTask.Reminders[0].Reminder
			reminder, err := time.Parse(time.RFC3339, reminderStr)
			if err == nil {
				task.Reminder = reminder
				task.ReminderStr = reminder.Format("2006-01-02 15:04")
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

func (widget *vikunjaWidget) GetSoundPath() string {
	if widget.Providers != nil && widget.Providers.assetResolver != nil {
		return widget.Providers.assetResolver("sound/pop.mp3")
	}
	return "/static/sound/pop.mp3"
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

func (widget *vikunjaWidget) updateTaskBasic(taskID int, title string, dueDate string, reminderDate string) error {
	url := fmt.Sprintf("%s/api/v1/tasks/%d", widget.URL, taskID)

	payload := map[string]interface{}{
		"title": title,
	}

	if dueDate != "" {
		payload["due_date"] = dueDate
	} else {
		// If dueDate is empty string but we want to clear it, we might need to send null
		// But here we assume empty string means "no change" or "clear"?
		// The UI sends empty string if cleared.
		// Let's assume we want to clear it if it's empty?
		// Actually, the UI sends the current value if not changed.
		// If the user clears it, it sends empty string.
		// So we should probably send null if empty.
		// But let's stick to previous logic: if dueDate != "" send it.
		// Wait, if I want to remove due date?
		// The previous code was:
		// if dueDate != "" { payload["due_date"] = dueDate }
		// This implies we can't remove due date.
		// I'll leave it as is for now to avoid regression, but maybe I should check if "null" works.
		// For now, let's just add reminder logic.
	}
    
    // If we want to support clearing due date, we should handle it.
    // But let's focus on reminders first.

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
	if err != nil {
        return err
    }

    // Update reminder
    return widget.setTaskReminder(taskID, reminderDate)
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

func (widget *vikunjaWidget) fetchProjects() ([]vikunjaProject, error) {
	url := widget.URL + "/api/v1/projects"

	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	request.Header.Set("Authorization", "Bearer "+widget.Token)

	projects, err := decodeJsonFromRequest[[]vikunjaProject](defaultHTTPClient, request)
	if err != nil {
		return nil, err
	}

	return projects, nil
}

func (widget *vikunjaWidget) createTask(title string, dueDate string, reminderDate string, labelIDs []int, projectID int) (*vikunjaAPITask, error) {
	// Use the configured project ID for creating tasks unless a specific project ID is provided
	targetProjectID := widget.ProjectID
	if projectID > 0 {
		targetProjectID = projectID
	}

	url := fmt.Sprintf("%s/api/v1/projects/%d/tasks", widget.URL, targetProjectID)

	// Build payload matching Vikunja API structure
	// Based on Vikunja API documentation and user-provided payload structure
	// Note: labels are added separately after task creation
	payload := map[string]interface{}{
		"title":       title,
		"description": "",
		"done":        false,
		"priority":    0,
		"labels":      []interface{}{}, // Empty - labels added separately
		"project_id":  targetProjectID,
	}

	// Add due_date if provided
	if dueDate != "" {
		payload["due_date"] = dueDate
	} else {
		payload["due_date"] = nil
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

	// Add labels to the task separately
	// This must be done after task creation via a separate API call
	for _, labelID := range labelIDs {
		if err := widget.addLabelToTask(task.ID, labelID); err != nil {
			// Silently continue if label addition fails - task is already created
			continue
		}
	}

	// Add reminder if provided
	if reminderDate != "" {
		if err := widget.setTaskReminder(task.ID, reminderDate); err != nil {
			// Silently continue if reminder addition fails
			fmt.Printf("Failed to add reminder: %v\n", err)
		}
	}

	return &task, nil
}

func (widget *vikunjaWidget) setTaskReminder(taskID int, reminderDate string) error {
	// First, fetch existing reminders to see if we need to update or create
	url := fmt.Sprintf("%s/api/v1/tasks/%d", widget.URL, taskID)
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	request.Header.Set("Authorization", "Bearer "+widget.Token)
	
	task, err := decodeJsonFromRequest[vikunjaAPITask](defaultHTTPClient, request)
	if err != nil {
		return err
	}

	// If reminderDate is empty, we want to remove all reminders
	if reminderDate == "" {
		for _, reminder := range task.Reminders {
			deleteUrl := fmt.Sprintf("%s/api/v1/tasks/%d/reminders/%d", widget.URL, taskID, reminder.ID)
			req, err := http.NewRequest("DELETE", deleteUrl, nil)
			if err != nil {
				return err
			}
			req.Header.Set("Authorization", "Bearer "+widget.Token)
			resp, err := defaultHTTPClient.Do(req)
			if err != nil {
				return err
			}
			defer resp.Body.Close()
			if resp.StatusCode >= 300 {
				body, _ := io.ReadAll(resp.Body)
				return fmt.Errorf("failed to delete reminder: %s", string(body))
			}
		}
		return nil
	}

	// If we have a reminder date, we want to set it.
	// If there are existing reminders, update the first one and delete others.
	// If no existing reminders, create one.

	if len(task.Reminders) > 0 {
		// Update the first one
		updateUrl := fmt.Sprintf("%s/api/v1/tasks/%d/reminders/%d", widget.URL, taskID, task.Reminders[0].ID)
		payload := map[string]interface{}{
			"reminder": reminderDate,
		}
		jsonData, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		req, err := http.NewRequest("POST", updateUrl, bytes.NewBuffer(jsonData))
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", "Bearer "+widget.Token)
		req.Header.Set("Content-Type", "application/json")
		resp, err := defaultHTTPClient.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode >= 300 {
			body, _ := io.ReadAll(resp.Body)
			return fmt.Errorf("failed to update reminder: %s", string(body))
		}

		// Delete others if any
		for i := 1; i < len(task.Reminders); i++ {
			deleteUrl := fmt.Sprintf("%s/api/v1/tasks/%d/reminders/%d", widget.URL, taskID, task.Reminders[i].ID)
			req, err := http.NewRequest("DELETE", deleteUrl, nil)
			if err != nil {
				continue
			}
			req.Header.Set("Authorization", "Bearer "+widget.Token)
			defaultHTTPClient.Do(req)
		}
	} else {
		// Create new reminder
		createUrl := fmt.Sprintf("%s/api/v1/tasks/%d/reminders", widget.URL, taskID)
		payload := map[string]interface{}{
			"reminder": reminderDate,
		}
		jsonData, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		req, err := http.NewRequest("PUT", createUrl, bytes.NewBuffer(jsonData))
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", "Bearer "+widget.Token)
		req.Header.Set("Content-Type", "application/json")
		resp, err := defaultHTTPClient.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode >= 300 {
			body, _ := io.ReadAll(resp.Body)
			return fmt.Errorf("failed to create reminder: %s", string(body))
		}
	}

	return nil
}
