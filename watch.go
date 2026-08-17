package main

import (
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

func (a *App) startFolderWatch() {
	a.watchMu.Lock()
	if a.watchStop != nil {
		close(a.watchStop)
		a.watchStop = nil
	}
	stop := make(chan struct{})
	a.watchStop = stop
	a.watchMu.Unlock()
	go a.hashLoop(stop)
}

func (a *App) hashLoop(stop chan struct{}) {
	a.scanDiskHashes()
	tick := time.NewTicker(time.Second)
	defer tick.Stop()
	for {
		select {
		case <-stop:
			return
		case <-tick.C:
			a.scanDiskHashes()
		}
	}
}

func (a *App) scanDiskHashes() {
	s, err := a.ready()
	if err != nil {
		return
	}
	root, err := a.writingRootIfSet()
	if err != nil || root == "" {
		return
	}
	a.watchMu.Lock()
	if a.syncing {
		a.watchMu.Unlock()
		return
	}
	a.syncing = true
	a.watchMu.Unlock()
	changed, err := s.SyncDiskHashes(root)
	a.watchMu.Lock()
	a.syncing = false
	a.watchMu.Unlock()
	if err != nil || !changed {
		return
	}
	a.emitFoldersChanged("")
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
	if a.watchStop != nil {
		close(a.watchStop)
		a.watchStop = nil
	}
}
