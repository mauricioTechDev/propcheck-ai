package session

import (
	"os"
	"testing"

	"github.com/mauricioTechDev/propcheck-ai/internal/types"
)

func TestCreateAndLoad(t *testing.T) {
	dir := t.TempDir()

	s, err := Create(dir)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if s.Phase != types.PhaseProperty {
		t.Errorf("Phase = %q, want %q", s.Phase, types.PhaseProperty)
	}

	loaded, err := Load(dir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if loaded.Phase != types.PhaseProperty {
		t.Errorf("Loaded Phase = %q, want %q", loaded.Phase, types.PhaseProperty)
	}
	if loaded.NextID != 1 {
		t.Errorf("Loaded NextID = %d, want 1", loaded.NextID)
	}
}

func TestExists(t *testing.T) {
	dir := t.TempDir()

	if Exists(dir) {
		t.Error("Exists should be false before Create")
	}

	Create(dir)
	if !Exists(dir) {
		t.Error("Exists should be true after Create")
	}
}

func TestLoadOrFail(t *testing.T) {
	dir := t.TempDir()

	_, err := LoadOrFail(dir)
	if err == nil {
		t.Error("LoadOrFail should error when no session exists")
	}

	Create(dir)
	s, err := LoadOrFail(dir)
	if err != nil {
		t.Fatalf("LoadOrFail: %v", err)
	}
	if s.Phase != types.PhaseProperty {
		t.Errorf("Phase = %q, want %q", s.Phase, types.PhaseProperty)
	}
}

func TestSaveAndLoadWithData(t *testing.T) {
	dir := t.TempDir()

	s := types.NewSession()
	s.TestCmd = "go test ./..."
	s.AgentMode = true
	s.AddProperty("sort idempotent", "invariant")
	s.SetCurrentProperty(1)
	s.ShrinkAnalysis = "found counter-example"
	s.AddEvent("init")

	if err := Save(dir, s); err != nil {
		t.Fatalf("Save: %v", err)
	}

	loaded, err := Load(dir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if loaded.TestCmd != "go test ./..." {
		t.Errorf("TestCmd = %q", loaded.TestCmd)
	}
	if !loaded.AgentMode {
		t.Error("AgentMode should be true")
	}
	if len(loaded.Properties) != 1 {
		t.Fatalf("Properties = %d, want 1", len(loaded.Properties))
	}
	if loaded.Properties[0].Category != "invariant" {
		t.Errorf("Category = %q", loaded.Properties[0].Category)
	}
	if loaded.CurrentPropertyID == nil || *loaded.CurrentPropertyID != 1 {
		t.Error("CurrentPropertyID should be 1")
	}
	if loaded.ShrinkAnalysis != "found counter-example" {
		t.Errorf("ShrinkAnalysis = %q", loaded.ShrinkAnalysis)
	}
	if len(loaded.History) != 1 {
		t.Errorf("History = %d events, want 1", len(loaded.History))
	}
}

func TestFilePath(t *testing.T) {
	got := FilePath("/tmp/myproject")
	want := "/tmp/myproject/.propcheck-ai.json"
	if got != want {
		t.Errorf("FilePath = %q, want %q", got, want)
	}
}

func TestLoadInvalidJSON(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(FilePath(dir), []byte("not json"), 0o644)

	_, err := Load(dir)
	if err == nil {
		t.Error("Load should error on invalid JSON")
	}
}
