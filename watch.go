package main

import (
	"os"
	"time"

	"loremetry/internal/store"

	"github.com/fsnotify/fsnotify"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

func (a *App) startFolderWatch() {
	a.watchMu.Lock()
	defer a.watchMu.Unlock()
	if a.watcher != nil {
		_ = a.watcher.Close()
		a.watcher = nil
	}
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return
	}
	a.watcher = w
	for _, dir := range a.watchDirs() {
		_ = w.Add(dir)
	}
	go a.watchLoop(w)
}

func (a *App) watchDirs() []string {
	var out []string
	root, err := a.writingRootIfSet()
	if err == nil && root != "" {
		out = append(out, store.FolderWatchDirs(root)...)
	}
	if a.store != nil {
		paths, err := a.store.AllHeaderOverridePaths()
		if err == nil {
			for _, p := range paths {
				out = append(out, store.FolderWatchDirs(p)...)
			}
		}
	}
	return out
}

func (a *App) watchLoop(w *fsnotify.Watcher) {
	var delay *time.Timer
	for {
		select {
		case ev, ok := <-w.Events:
			if !ok {
				return
			}
			if ev.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Remove|fsnotify.Rename) == 0 {
				continue
			}
			name := ev.Name
			if delay != nil {
				delay.Reset(400 * time.Millisecond)
				continue
			}
			delay = time.AfterFunc(400*time.Millisecond, func() {
				a.refreshWatchDirs()
				a.emitFoldersChanged(name)
			})
		case _, ok := <-w.Errors:
			if !ok {
				return
			}
		}
	}
}

func (a *App) refreshWatchDirs() {
	if a.watcher == nil {
		return
	}
	for _, dir := range a.watchDirs() {
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			_ = a.watcher.Add(dir)
		}
	}
}

func (a *App) emitFoldersChanged(path string) {
	if a.ctx == nil {
		return
	}
	runtime.EventsEmit(a.ctx, "folders-changed", path)
}

func (a *App) stopFolderWatch() {
	a.watchMu.Lock()
	defer a.watchMu.Unlock()
	if a.watcher != nil {
		_ = a.watcher.Close()
		a.watcher = nil
	}
}
