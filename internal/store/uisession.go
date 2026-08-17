package store

import (
	"encoding/json"
	"strings"
)

const (
	uiSessionKey     = "ui_session"
	lastPenKey       = "last_pen"
	legacySessionKey = "last_open_session"
	legacyRestoreKey = "restore_open_items"
)

type UIFileSelection struct {
	Type string `json:"type"`
	Dir  string `json:"dir"`
	Name string `json:"name"`
}

type UISession struct {
	Selection   UIFileSelection `json:"selection"`
	OpenSeries  []string        `json:"open_series"`
	OpenStories []string        `json:"open_stories"`
	OpenFolders []string        `json:"open_folders"`
	RestoreOpen bool            `json:"restore_open"`
	LastPen     string          `json:"last_pen"`
}

func emptyUISession() UISession {
	return UISession{
		Selection:   UIFileSelection{Type: "empty"},
		OpenSeries:  []string{},
		OpenStories: []string{},
		OpenFolders: []string{},
		RestoreOpen: true,
	}
}

func (s *Store) GetUISession() (UISession, error) {
	out := emptyUISession()
	raw, err := s.settingValue(uiSessionKey)
	if err == nil && strings.TrimSpace(raw) != "" {
		if err := json.Unmarshal([]byte(raw), &out); err != nil {
			out = emptyUISession()
		}
	} else {
		out = s.migrateLegacySession()
	}
	if out.OpenSeries == nil {
		out.OpenSeries = []string{}
	}
	if out.OpenStories == nil {
		out.OpenStories = []string{}
	}
	if out.OpenFolders == nil {
		out.OpenFolders = []string{}
	}
	if out.Selection.Type == "" {
		out.Selection.Type = "empty"
	}
	if pen, err := s.settingValue(lastPenKey); err == nil {
		out.LastPen = pen
	}
	return out, nil
}

func (s *Store) SetUISelection(dir, name string) (UISession, error) {
	sess, err := s.GetUISession()
	if err != nil {
		return UISession{}, err
	}
	dir, name = strings.TrimSpace(dir), strings.TrimSpace(name)
	if dir == "" || name == "" {
		sess.Selection = UIFileSelection{Type: "empty"}
	} else {
		sess.Selection = UIFileSelection{Type: "file", Dir: dir, Name: name}
	}
	return s.saveUISession(sess)
}

func (s *Store) ToggleOpenSeries(path string) (UISession, error) {
	return s.toggleOpen(path, true)
}

func (s *Store) ToggleOpenStory(path string) (UISession, error) {
	return s.toggleOpen(path, false)
}

func (s *Store) EnsureOpenSeries(path string) (UISession, error) {
	return s.ensureOpen(path, true)
}

func (s *Store) EnsureOpenStory(path string) (UISession, error) {
	return s.ensureOpen(path, false)
}

func (s *Store) ToggleOpenFolder(path string) (UISession, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return s.GetUISession()
	}
	sess, err := s.GetUISession()
	if err != nil {
		return UISession{}, err
	}
	sess.OpenFolders = toggleID(sess.OpenFolders, path)
	return s.saveUISession(sess)
}

func (s *Store) EnsureOpenFolder(path string) (UISession, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return s.GetUISession()
	}
	sess, err := s.GetUISession()
	if err != nil {
		return UISession{}, err
	}
	sess.OpenFolders = addID(sess.OpenFolders, path)
	return s.saveUISession(sess)
}

func (s *Store) SetRestoreOpen(on bool) (UISession, error) {
	sess, err := s.GetUISession()
	if err != nil {
		return UISession{}, err
	}
	sess.RestoreOpen = on
	return s.saveUISession(sess)
}

func (s *Store) SetLastPen(name string) (UISession, error) {
	name = strings.TrimSpace(name)
	if _, err := s.PutSetting(lastPenKey, name); err != nil {
		return UISession{}, err
	}
	sess, err := s.GetUISession()
	if err != nil {
		return UISession{}, err
	}
	sess.LastPen = name
	return sess, nil
}

func (s *Store) toggleOpen(path string, series bool) (UISession, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return s.GetUISession()
	}
	sess, err := s.GetUISession()
	if err != nil {
		return UISession{}, err
	}
	if series {
		sess.OpenSeries = toggleID(sess.OpenSeries, path)
	} else {
		sess.OpenStories = toggleID(sess.OpenStories, path)
	}
	return s.saveUISession(sess)
}

func (s *Store) ensureOpen(path string, series bool) (UISession, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return s.GetUISession()
	}
	sess, err := s.GetUISession()
	if err != nil {
		return UISession{}, err
	}
	if series {
		sess.OpenSeries = addID(sess.OpenSeries, path)
	} else {
		sess.OpenStories = addID(sess.OpenStories, path)
	}
	return s.saveUISession(sess)
}

func (s *Store) saveUISession(sess UISession) (UISession, error) {
	if sess.OpenSeries == nil {
		sess.OpenSeries = []string{}
	}
	if sess.OpenStories == nil {
		sess.OpenStories = []string{}
	}
	if sess.OpenFolders == nil {
		sess.OpenFolders = []string{}
	}
	b, err := json.Marshal(sess)
	if err != nil {
		return UISession{}, err
	}
	if _, err := s.PutSetting(uiSessionKey, string(b)); err != nil {
		return UISession{}, err
	}
	return sess, nil
}

func (s *Store) settingValue(key string) (string, error) {
	got, err := s.GetSetting(key)
	if err != nil {
		return "", err
	}
	return got.Value, nil
}

func (s *Store) migrateLegacySession() UISession {
	out := emptyUISession()
	if v, err := s.settingValue(legacyRestoreKey); err == nil && v == "false" {
		out.RestoreOpen = false
	}
	raw, err := s.settingValue(legacySessionKey)
	if err != nil || strings.TrimSpace(raw) == "" {
		return out
	}
	var old struct {
		Selection   UIFileSelection `json:"selection"`
		OpenSeries  []string        `json:"openSeries"`
		OpenStories []string        `json:"openStories"`
	}
	if json.Unmarshal([]byte(raw), &old) != nil {
		return out
	}
	out.Selection = old.Selection
	out.OpenSeries = old.OpenSeries
	out.OpenStories = old.OpenStories
	return out
}

func toggleID(ids []string, id string) []string {
	out := make([]string, 0, len(ids))
	found := false
	for _, x := range ids {
		if x == id {
			found = true
			continue
		}
		out = append(out, x)
	}
	if !found {
		out = append(out, id)
	}
	return out
}

func addID(ids []string, id string) []string {
	for _, x := range ids {
		if x == id {
			return ids
		}
	}
	return append(ids, id)
}
