package session

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/mauricioTechDev/propcheck-ai/internal/types"
)

// DefaultFileName is the session file name.
const DefaultFileName = ".propcheck-ai.json"

// FilePath returns the full path to the session file in the given directory.
func FilePath(dir string) string {
	return filepath.Join(dir, DefaultFileName)
}

// Exists checks whether a session file exists in the given directory.
func Exists(dir string) bool {
	_, err := os.Stat(FilePath(dir))
	return err == nil
}

// Create initializes a new session and saves it to disk.
func Create(dir string) (*types.Session, error) {
	s := types.NewSession()
	if err := Save(dir, s); err != nil {
		return nil, fmt.Errorf("creating session: %w", err)
	}
	return s, nil
}

// Load reads the session from disk.
func Load(dir string) (*types.Session, error) {
	data, err := os.ReadFile(FilePath(dir))
	if err != nil {
		return nil, fmt.Errorf("reading session: %w", err)
	}
	var s types.Session
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("parsing session: %w", err)
	}
	return &s, nil
}

// LoadOrFail loads the session or returns a user-friendly error.
func LoadOrFail(dir string) (*types.Session, error) {
	if !Exists(dir) {
		return nil, fmt.Errorf("no propcheck-ai session found. Run 'propcheck-ai init' first")
	}
	return Load(dir)
}

// Save writes the session to disk as indented JSON.
func Save(dir string, s *types.Session) error {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("encoding session: %w", err)
	}
	data = append(data, '\n')
	if err := os.WriteFile(FilePath(dir), data, 0o644); err != nil {
		return fmt.Errorf("writing session: %w", err)
	}
	return nil
}
