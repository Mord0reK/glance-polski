package glance

import (
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestGoogleComputeParseServiceAccountKeyRawJSON(t *testing.T) {
	raw := `{"type":"service_account","project_id":"demo","private_key_id":"abc"}`
	widget := &googleComputeWidget{ServiceAccountKey: raw}

	data, err := widget.parseServiceAccountKey()
	if err != nil {
		t.Fatalf("expected raw JSON to be accepted, got error: %v", err)
	}

	if !json.Valid(data) {
		t.Fatalf("expected valid JSON, got: %s", string(data))
	}
}

func TestGoogleComputeParseServiceAccountKeyBase64(t *testing.T) {
	raw := `{"type":"service_account","project_id":"demo","private_key_id":"abc"}`
	encoded := base64.StdEncoding.EncodeToString([]byte(raw))
	widget := &googleComputeWidget{ServiceAccountKey: encoded}

	data, err := widget.parseServiceAccountKey()
	if err != nil {
		t.Fatalf("expected base64 JSON to be accepted, got error: %v", err)
	}

	if string(data) != raw {
		t.Fatalf("expected decoded data to match original JSON")
	}
}

func TestGoogleComputeParseServiceAccountKeyFile(t *testing.T) {
	raw := `{"type":"service_account","project_id":"demo","private_key_id":"abc"}`

	tempDir := t.TempDir()
	path := filepath.Join(tempDir, "sa.json")

	if err := os.WriteFile(path, []byte(raw), 0o600); err != nil {
		t.Fatalf("failed to write temp key file: %v", err)
	}

	widget := &googleComputeWidget{ServiceAccountKey: path}
	data, err := widget.parseServiceAccountKey()
	if err != nil {
		t.Fatalf("expected file path to be accepted, got error: %v", err)
	}

	if string(data) != raw {
		t.Fatalf("expected file contents to be returned")
	}
}
