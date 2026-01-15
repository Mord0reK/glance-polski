package glance

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

var vikunjaWidgetTemplate = mustParseTemplate("vikunja.html", "widget-base.html")

type vikunjaWidget struct {
	widgetBase     `yaml:",inline"`
	URL            string `yaml:"url"`
	Token          string `yaml:"token"`
	Limit          int    `yaml:"limit"`
	ProjectID      int    `yaml:"project-id"` // Project ID for creating new tasks
	AffineURL      string `yaml:"affine-url"`
	AffineEmail    string `yaml:"affine-email"`
	AffinePassword string `yaml:"affine-password"`
	Tasks          []vikunjaTask
}

type vikunjaTask struct {
	ID              int
	Title           string
	DueDate         time.Time
	Done            bool
	PercentDone     int
	Labels          []vikunjaLabel
	TimeLeft        string
	DueDateStr      string
	Reminder        time.Time
	ReminderStr     string
	IsOverdue       bool
	TaskURL         string
	AffineNoteURL   string
	AffineNoteTitle string
	CustomLinkURL   string
	CustomLinkTitle string
}

type vikunjaLabel struct {
	ID    int
	Title string
	Color string
}

type vikunjaAPITask struct {
	ID            int                  `json:"id"`
	Title         string               `json:"title"`
	Done          bool                 `json:"done"`
	DueDate       string               `json:"due_date"`
	PercentDone   float64              `json:"percent_done"`
	Labels        []vikunjaAPILabel    `json:"labels"`
	Reminders     []vikunjaAPIReminder `json:"reminders"`
	Description   string               `json:"description"`
	AffineNoteURL string               `json:"affine_note_url,omitempty"`
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
	ID         int    `json:"id"`
	Title      string `json:"title"`
	IsArchived bool   `json:"is_archived"`
}

// Affine API structures
type affineSignInRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type affineSignInResponse struct {
	Token string `json:"token"`
}

type affineGraphQLRequest struct {
	Query         string                 `json:"query"`
	Variables     map[string]interface{} `json:"variables"`
	OperationName string                 `json:"operationName"`
}

type affineGraphQLResponse struct {
	Data struct {
		Workspace struct {
			Doc struct {
				ID          string `json:"id"`
				Mode        string `json:"mode"`
				DefaultRole string `json:"defaultRole"`
				Public      bool   `json:"public"`
				Title       string `json:"title"`
				Summary     string `json:"summary"`
			} `json:"doc"`
		} `json:"workspace"`
	} `json:"data"`
}

func (widget *vikunjaWidget) initialize() error {
	widget.withTitle("Vikunja").withCacheDuration(5 * time.Minute)

	if widget.URL == "" {
		return fmt.Errorf("URL is required")
	}
	widget.URL = strings.TrimSuffix(widget.URL, "/")

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

	// Authenticate with Affine once if configured, before processing tasks
	var affineToken string
	if widget.AffineURL != "" && widget.AffineEmail != "" && widget.AffinePassword != "" {
		affineToken, err = widget.affineSignIn()
		if err != nil {
			// Log error but continue - we can still show tasks without Affine note titles
			slog.Error("Failed to authenticate with Affine", "error", err)
			affineToken = ""
		}
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
			TaskURL:     fmt.Sprintf("%s/tasks/%d", strings.TrimRight(widget.URL, "/"), apiTask.ID),
		}

		if apiTask.DueDate != "" {
			dueDate, err := time.Parse(time.RFC3339, apiTask.DueDate)
			if err == nil {
				task.DueDate = dueDate
				task.DueDateStr = dueDate.Format("2006-01-02 15:04")
				task.TimeLeft = formatTimeLeft(now, dueDate)
				if dueDate.Before(now) {
					task.IsOverdue = true
				}
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

		// Extract Affine note URL, Custom link URL and Custom link title from description
		// Format: AFFINE_NOTE:https://affine-url/workspace/...
		//         CUSTOM_LINK:https://any-url/...
		//         CUSTOM_LINK_TITLE:My Link Title
		// Description can contain all, separated by newlines
		if apiTask.Description != "" {
			lines := strings.Split(apiTask.Description, "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if strings.HasPrefix(line, "AFFINE_NOTE:") {
					affineURL := strings.TrimPrefix(line, "AFFINE_NOTE:")
					task.AffineNoteURL = affineURL

					// Fetch note title if Affine token is available
					if affineToken != "" {
						noteTitle, err := widget.fetchAffineNoteTitleWithToken(affineURL, affineToken)
						if err == nil && noteTitle != "" {
							task.AffineNoteTitle = noteTitle
						}
					}
				} else if strings.HasPrefix(line, "CUSTOM_LINK:") {
					task.CustomLinkURL = strings.TrimPrefix(line, "CUSTOM_LINK:")
				} else if strings.HasPrefix(line, "CUSTOM_LINK_TITLE:") {
					task.CustomLinkTitle = strings.TrimPrefix(line, "CUSTOM_LINK_TITLE:")
				}
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

func (widget *vikunjaWidget) updateTaskBasic(taskID int, title string, dueDate string, affineNoteURL string, customLinkURL string, customLinkTitle string) error {
	url := fmt.Sprintf("%s/api/v1/tasks/%d", widget.URL, taskID)

	payload := map[string]interface{}{
		"title": title,
	}

	if dueDate != "" {
		payload["due_date"] = dueDate
	}

	// Store Affine note URL, Custom link URL and Custom link title in description with special prefixes
	var descriptionParts []string
	if affineNoteURL != "" {
		descriptionParts = append(descriptionParts, "AFFINE_NOTE:"+affineNoteURL)
	}
	if customLinkURL != "" {
		descriptionParts = append(descriptionParts, "CUSTOM_LINK:"+customLinkURL)
	}
	if customLinkTitle != "" {
		descriptionParts = append(descriptionParts, "CUSTOM_LINK_TITLE:"+customLinkTitle)
	}
	payload["description"] = strings.Join(descriptionParts, "\n")

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

	return nil
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

	// Filter archived projects
	var activeProjects []vikunjaProject
	for _, p := range projects {
		if !p.IsArchived {
			activeProjects = append(activeProjects, p)
		}
	}

	return activeProjects, nil
}

func (widget *vikunjaWidget) createTask(title string, dueDate string, labelIDs []int, projectID int, affineNoteURL string, customLinkURL string, customLinkTitle string) (*vikunjaAPITask, error) {
	// Use the configured project ID for creating tasks unless a specific project ID is provided
	targetProjectID := widget.ProjectID
	if projectID > 0 {
		targetProjectID = projectID
	}

	url := fmt.Sprintf("%s/api/v1/projects/%d/tasks", widget.URL, targetProjectID)

	// Build payload matching Vikunja API structure
	// Based on Vikunja API documentation and user-provided payload structure
	// Note: labels are added separately after task creation
	var descriptionParts []string
	if affineNoteURL != "" {
		descriptionParts = append(descriptionParts, "AFFINE_NOTE:"+affineNoteURL)
	}
	if customLinkURL != "" {
		descriptionParts = append(descriptionParts, "CUSTOM_LINK:"+customLinkURL)
	}
	if customLinkTitle != "" {
		descriptionParts = append(descriptionParts, "CUSTOM_LINK_TITLE:"+customLinkTitle)
	}
	description := strings.Join(descriptionParts, "\n")

	payload := map[string]interface{}{
		"title":       title,
		"description": description,
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

	if task.ID == 0 {
		return nil, fmt.Errorf("created task has invalid ID (0)")
	}

	// Add labels to the task separately
	// This must be done after task creation via a separate API call
	for _, labelID := range labelIDs {
		if err := widget.addLabelToTask(task.ID, labelID); err != nil {
			// Silently continue if label addition fails - task is already created
			continue
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
				return fmt.Errorf("failed to delete reminder (status %d): %s", resp.StatusCode, string(body))
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
			return fmt.Errorf("failed to update reminder (status %d): %s", resp.StatusCode, string(body))
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
			return fmt.Errorf("failed to create reminder (status %d): %s", resp.StatusCode, string(body))
		}
	}

	return nil
}

// parseAffineURL extracts workspaceID and pageID from Affine URL
// Expected format: https://affine-url/workspace/WORKSPACE_ID/PAGE_ID
func (widget *vikunjaWidget) parseAffineURL(affineURL string) (workspaceID, pageID string, err error) {
	if affineURL == "" {
		return "", "", fmt.Errorf("empty Affine URL")
	}

	u, err := url.Parse(affineURL)
	if err != nil {
		return "", "", fmt.Errorf("invalid URL: %w", err)
	}

	// Parse path segments
	// Expected: /workspace/WORKSPACE_ID/PAGE_ID
	parts := strings.Split(strings.Trim(u.Path, "/"), "/")

	// Find workspace and page IDs
	for i, part := range parts {
		if part == "workspace" && i+2 < len(parts) {
			workspaceID = parts[i+1]
			pageID = parts[i+2]
			return workspaceID, pageID, nil
		}
	}

	return "", "", fmt.Errorf("could not extract workspace ID and page ID from URL")
}

// affineSignIn authenticates with Affine and returns a token
func (widget *vikunjaWidget) affineSignIn() (string, error) {
	if widget.AffineURL == "" || widget.AffineEmail == "" || widget.AffinePassword == "" {
		return "", fmt.Errorf("Affine credentials not configured")
	}

	signInURL := strings.TrimSuffix(widget.AffineURL, "/") + "/api/auth/sign-in"

	payload := affineSignInRequest{
		Email:    widget.AffineEmail,
		Password: widget.AffinePassword,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal sign-in request: %w", err)
	}

	request, err := http.NewRequest("POST", signInURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create sign-in request: %w", err)
	}

	request.Header.Set("Content-Type", "application/json")

	response, err := defaultHTTPClient.Do(request)
	if err != nil {
		return "", fmt.Errorf("failed to execute sign-in request: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK && response.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(response.Body)
		return "", fmt.Errorf("sign-in failed with status %d: %s", response.StatusCode, string(body))
	}

	// Try to extract token from cookies first (common pattern)
	for _, cookie := range response.Cookies() {
		if cookie.Name == "affine_session" || cookie.Name == "token" {
			return cookie.Value, nil
		}
	}

	// Try to parse JSON response
	var signInResp affineSignInResponse
	if err := json.NewDecoder(response.Body).Decode(&signInResp); err != nil {
		return "", fmt.Errorf("failed to decode sign-in response: %w", err)
	}

	if signInResp.Token == "" {
		return "", fmt.Errorf("no token in sign-in response")
	}

	return signInResp.Token, nil
}

// fetchAffineNoteTitle fetches the title of an Affine note
// This method authenticates each time it's called
func (widget *vikunjaWidget) fetchAffineNoteTitle(affineNoteURL string) (string, error) {
	if affineNoteURL == "" {
		return "", nil
	}

	// Sign in to Affine
	token, err := widget.affineSignIn()
	if err != nil {
		return "", fmt.Errorf("failed to sign in to Affine: %w", err)
	}

	return widget.fetchAffineNoteTitleWithToken(affineNoteURL, token)
}

// fetchAffineNoteTitleWithToken fetches the title of an Affine note using a provided token
func (widget *vikunjaWidget) fetchAffineNoteTitleWithToken(affineNoteURL string, token string) (string, error) {
	if affineNoteURL == "" {
		return "", nil
	}

	if token == "" {
		return "", fmt.Errorf("empty authentication token provided")
	}

	// Parse the Affine URL to extract workspace and page IDs
	workspaceID, pageID, err := widget.parseAffineURL(affineNoteURL)
	if err != nil {
		return "", fmt.Errorf("failed to parse Affine URL: %w", err)
	}

	// Prepare GraphQL request
	graphQLURL := strings.TrimSuffix(widget.AffineURL, "/") + "/graphql"

	query := `query getWorkspacePageById($workspaceId: String!, $pageId: String!) {
  workspace(id: $workspaceId) {
    doc(docId: $pageId) {
      id
      mode
      defaultRole
      public
      title
      summary
    }
  }
}`

	variables := map[string]interface{}{
		"workspaceId": workspaceID,
		"pageId":      pageID,
	}

	graphQLRequest := affineGraphQLRequest{
		Query:         query,
		Variables:     variables,
		OperationName: "getWorkspacePageById",
	}

	jsonData, err := json.Marshal(graphQLRequest)
	if err != nil {
		return "", fmt.Errorf("failed to marshal GraphQL request: %w", err)
	}

	request, err := http.NewRequest("POST", graphQLURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create GraphQL request: %w", err)
	}

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Bearer "+token)

	response, err := defaultHTTPClient.Do(request)
	if err != nil {
		return "", fmt.Errorf("failed to execute GraphQL request: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(response.Body)
		return "", fmt.Errorf("GraphQL request failed with status %d: %s", response.StatusCode, string(body))
	}

	var graphQLResp affineGraphQLResponse
	if err := json.NewDecoder(response.Body).Decode(&graphQLResp); err != nil {
		return "", fmt.Errorf("failed to decode GraphQL response: %w", err)
	}

	return graphQLResp.Data.Workspace.Doc.Title, nil
}
