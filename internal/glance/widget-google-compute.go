package glance

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"path"
	"sort"
	"strings"
	"time"

	"google.golang.org/api/compute/v1"
	"google.golang.org/api/option"
)

var googleComputeWidgetTemplate = mustParseTemplate("google-compute.html", "widget-base.html")

type googleComputeWidget struct {
	widgetBase        `yaml:",inline"`
	ProjectID         string        `yaml:"project-id"`
	ServiceAccountKey string        `yaml:"service-account-key"`
	Zones             []string      `yaml:"zones"`
	Instances         []gceInstance `yaml:"-"`

	credentialsJSON []byte `yaml:"-"`
}

type gceInstance struct {
	Name        string
	Zone        string
	Status      string
	StatusClass string
	MachineType string
	InternalIP  string
	ExternalIP  string
	CanStart    bool
	CanStop     bool
	CanReset    bool
}

func (widget *googleComputeWidget) initialize() error {
	widget.withTitle("Google Compute Engine").withCacheDuration(1 * time.Minute)

	if widget.ProjectID == "" {
		return fmt.Errorf("project-id is required")
	}

	if widget.ServiceAccountKey == "" {
		return fmt.Errorf("service-account-key is required")
	}

	creds, err := widget.parseServiceAccountKey()
	if err != nil {
		return err
	}

	widget.credentialsJSON = creds

	return nil
}

func (widget *googleComputeWidget) parseServiceAccountKey() ([]byte, error) {
	key := strings.TrimSpace(widget.ServiceAccountKey)
	if key == "" {
		return nil, fmt.Errorf("service-account-key is required")
	}

	if fileInfo, err := os.Stat(key); err == nil && !fileInfo.IsDir() {
		data, readErr := os.ReadFile(key)
		if readErr != nil {
			return nil, fmt.Errorf("reading service account file: %w", readErr)
		}
		if json.Valid(data) {
			return data, nil
		}
		return nil, fmt.Errorf("service account file does not contain valid JSON")
	}

	if decoded, err := base64.StdEncoding.DecodeString(key); err == nil && json.Valid(decoded) {
		return decoded, nil
	}

	if json.Valid([]byte(key)) {
		return []byte(key), nil
	}

	return nil, fmt.Errorf("service-account-key must be a path, base64 string or JSON service account key")
}

func (widget *googleComputeWidget) update(ctx context.Context) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	instances, err := widget.fetchInstances(ctx)
	if !widget.canContinueUpdateAfterHandlingErr(err) {
		return
	}

	widget.Instances = instances
}

func (widget *googleComputeWidget) fetchInstances(ctx context.Context) ([]gceInstance, error) {
	service, err := widget.newComputeService(ctx)
	if err != nil {
		return nil, fmt.Errorf("creating compute service: %w", err)
	}

	call := service.Instances.AggregatedList(widget.ProjectID).Context(ctx)

	allowedZones := make(map[string]struct{})
	for _, z := range widget.Zones {
		allowedZones[strings.ToLower(z)] = struct{}{}
	}

	var instances []gceInstance
	err = call.Pages(ctx, func(resp *compute.InstanceAggregatedList) error {
		for zoneKey, scopedList := range resp.Items {
			zoneName := path.Base(zoneKey)
			if len(allowedZones) > 0 {
				if _, ok := allowedZones[strings.ToLower(zoneName)]; !ok {
					continue
				}
			}

			for _, instance := range scopedList.Instances {
				instances = append(instances, mapGCEInstance(instance, zoneName))
			}
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("listing instances: %w", err)
	}

	sort.Slice(instances, func(i, j int) bool {
		return strings.ToLower(instances[i].Name) < strings.ToLower(instances[j].Name)
	})

	return instances, nil
}

func mapGCEInstance(instance *compute.Instance, zone string) gceInstance {
	machineType := path.Base(instance.MachineType)

	var internalIP, externalIP string
	for _, iface := range instance.NetworkInterfaces {
		if internalIP == "" && iface.NetworkIP != "" {
			internalIP = iface.NetworkIP
		}

		for _, cfg := range iface.AccessConfigs {
			if externalIP == "" && cfg.NatIP != "" {
				externalIP = cfg.NatIP
			}
		}
	}

	status := strings.ToUpper(instance.Status)

	return gceInstance{
		Name:        instance.Name,
		Zone:        zone,
		Status:      status,
		StatusClass: statusToClass(status),
		MachineType: machineType,
		InternalIP:  internalIP,
		ExternalIP:  externalIP,
		CanStart:    status == "TERMINATED" || status == "STOPPED" || status == "SUSPENDED",
		CanStop:     status == "RUNNING",
		CanReset:    status == "RUNNING",
	}
}

func statusToClass(status string) string {
	switch status {
	case "RUNNING":
		return "running"
	case "STOPPED", "TERMINATED":
		return "stopped"
	case "PROVISIONING", "STAGING", "STARTING":
		return "starting"
	case "SUSPENDING", "SUSPENDED", "STOPPING":
		return "stopping"
	default:
		return "unknown"
	}
}

func (widget *googleComputeWidget) newComputeService(ctx context.Context) (*compute.Service, error) {
	if len(widget.credentialsJSON) == 0 {
		return nil, fmt.Errorf("service-account-key is not configured")
	}

	return compute.NewService(ctx, option.WithCredentialsJSON(widget.credentialsJSON))
}

func (widget *googleComputeWidget) performInstanceAction(ctx context.Context, action, zone, name string) error {
	if name == "" || zone == "" {
		return fmt.Errorf("instance name and zone are required")
	}

	allowedZones := make(map[string]struct{})
	for _, z := range widget.Zones {
		allowedZones[strings.ToLower(z)] = struct{}{}
	}

	if len(allowedZones) > 0 {
		if _, ok := allowedZones[strings.ToLower(zone)]; !ok {
			return fmt.Errorf("zone %s is not allowed for this widget", zone)
		}
	}

	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	service, err := widget.newComputeService(ctx)
	if err != nil {
		return err
	}

	switch action {
	case "start":
		_, err = service.Instances.Start(widget.ProjectID, zone, name).Context(ctx).Do()
	case "stop":
		_, err = service.Instances.Stop(widget.ProjectID, zone, name).Context(ctx).Do()
	case "restart":
		_, err = service.Instances.Reset(widget.ProjectID, zone, name).Context(ctx).Do()
	default:
		return fmt.Errorf("unsupported action: %s", action)
	}

	if err != nil {
		return fmt.Errorf("executing %s: %w", action, err)
	}

	return nil
}

func (widget *googleComputeWidget) Render() template.HTML {
	return widget.renderTemplate(widget, googleComputeWidgetTemplate)
}
