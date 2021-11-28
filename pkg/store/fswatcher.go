package store

import (
	"path/filepath"

	"github.com/rjeczalik/notify"
	"github.com/rs/zerolog/log"
)

// Watch
type FSWatcher struct {
	FChan chan notify.EventInfo
	Paths []string
}

// watcherStart
func (w *FSWatcher) FSWatcherStart() {
	//
	for _, path := range w.Paths {
		log.Info().Msgf("add '%s' to watcher", path)
		go watcherInit(w.FChan, path)
	}
}

// watcherStop
func (w *FSWatcher) FSWatcherStop() {
	notify.Stop(w.FChan)
}

// watcherRestart
func (w *FSWatcher) FSWatcherRestart() {
	w.FSWatcherStop()
	w.FSWatcherStart()
}

// watcherInit
func watcherInit(ec chan notify.EventInfo, path string) {
	path = filepath.Join(path, "/...")
	if err := notify.Watch(path, ec, notify.Create, notify.Write, notify.Remove, notify.Rename); err != nil {
		log.Error().Err(err).Msgf("watch path %s error: %s\n", path, err)
	}
}
