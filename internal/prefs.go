package internal

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

type Preferences struct {
	PrinterIP   string `json:"printer_ip"`
	PrinterPort string `json:"printer_port"`
}

type PrefsStore struct {
	appName  string
	dir      string
	Path     string
	prefs    Preferences
	loaded   bool
	modified bool
}

func NewPrefsStore(appName string) (*PrefsStore, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return nil, fmt.Errorf("get user config dir: %w", err)
	}
	dir := filepath.Join(base, appName)
	Path := filepath.Join(dir, "config.json")
	return &PrefsStore{appName: appName, dir: dir, Path: Path}, nil
}

func (s *PrefsStore) defaultPrefs() Preferences {
	return Preferences{
		PrinterIP:   "",
		PrinterPort: "9100",
	}
}

func (s *PrefsStore) Load() error {
	if err := os.MkdirAll(s.dir, 0o755); err != nil {
		return fmt.Errorf("ensure config dir: %w", err)
	}

	f, err := os.Open(s.Path)
	if errors.Is(err, os.ErrNotExist) {
		s.prefs = s.defaultPrefs()
		s.loaded = true
		return s.Save()
	} else if err != nil {
		return fmt.Errorf("open prefs: %w", err)
	}
	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		return fmt.Errorf("read prefs: %w", err)
	}
	if len(data) == 0 {
		s.prefs = s.defaultPrefs()
		s.loaded = true
		return s.Save()
	}

	if err := json.Unmarshal(data, &s.prefs); err != nil {
		s.prefs = s.defaultPrefs()
	}
	s.loaded = true
	return nil
}

func (s *PrefsStore) Save() error {
	if !s.loaded {
		return errors.New("prefs not loaded")
	}
	tmp := s.Path + fmt.Sprintf(".tmp-%d", time.Now().UnixNano())

	b, err := json.MarshalIndent(s.prefs, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal prefs: %w", err)
	}

	if err := os.WriteFile(tmp, b, 0o600); err != nil {
		return fmt.Errorf("write temp prefs: %w", err)
	}

	if err := os.Rename(tmp, s.Path); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("rename prefs: %w", err)
	}
	s.modified = false
	return nil
}

func (s *PrefsStore) GetPreferences() (Preferences, error) {
	if !s.loaded {
		if err := s.Load(); err != nil {
			return Preferences{}, err
		}
	}
	return s.prefs, nil
}

func (s *PrefsStore) UpdatePreferences(p Preferences) error {
	if !s.loaded {
		if err := s.Load(); err != nil {
			return err
		}
	}

	s.prefs = p
	s.modified = true
	return s.Save()
}
