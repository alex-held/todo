package store

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/cortesi/moddwatch"
	"github.com/radovskyb/watcher"
	"github.com/rjeczalik/notify"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/syncmap"
)

type store struct {
	dir string
	*moddwatch.Watcher

	todos    syncmap.Map
	sections sync.Map
}

func NewStore(dir string) Store {
	dir, err := filepath.EvalSymlinks(dir)
	if err != nil {
		log.Error().Err(err).Str("path", dir).Msg("unable to eval symlink")
	}

	s := &store{
		dir: dir,
	}

	if err := filepath.Walk(dir, func(path string, fi fs.FileInfo, err error) error {
		if path == dir {
			return nil
		}
		if fi.IsDir() {
			name := filepath.Base(path)
			// s.todos.Store(name, path)
			s.sections.Store(name, path)
		}
		return err
	}); err != nil {
		log.Error().Err(err).Str("path", s.dir).Msgf("unable to collect initial sections")
		return nil
	}

	if err := s.watch(); err != nil {
		log.Error().Err(err).Str("path", s.dir).Msgf("unable to create store watcher")
		return s
	}

	return s
}

func (s *store) initWatcher(c chan []string) error {
	root := s.dir
	modC := make(chan *moddwatch.Mod, 1)

	go func() {
		for mod := range modC {
			fmt.Printf("modified files: %#v", mod.All())
			if !mod.Empty() {
				c <- mod.All()
			}
		}
	}()

	watch, err := moddwatch.Watch(
		root,
		[]string{root + "/...", "**"},
		[]string{},
		time.Millisecond*100,
		modC,
	)

	if err != nil {
		return err
	}
	s.Watcher = watch

	time.Sleep(5 * time.Second)

	return nil
}

func (s *store) watch() error {

	dir := s.dir + "/..."

	createC := make(chan notify.EventInfo, 2048)
	deleteC := make(chan notify.EventInfo, 2048)

	go func() {
		for {
			select {
			case event := <-createC:
				if event.Path() == s.dir {
					continue
				}
				name := filepath.Base(event.Path())
				log.Debug().Str("name", name).Str("op", event.Event().String()).Msg("create section")
				//	s.todos.Store(name, event.Path())
				s.sections.Store(name, event.Path())
			}
		}
	}()

	go func() {
		for {
			select {
			case event := <-deleteC:
				if event.Path() == s.dir {
					continue
				}
				name := filepath.Base(event.Path())
				log.Debug().Str("name", name).Str("op", event.Event().String()).Msg("delete section")
				// s.todos.Delete(name)
				s.sections.Delete(name)
			}
		}
	}()

	if err := notify.Watch(dir, createC, notify.Create, notify.Write); err != nil {
		return err
	}
	if err := notify.Watch(dir, deleteC, notify.Remove); err != nil {
		return err
	}

	return nil
}

func (s *store) GetSections() (sections []string, err error) {
	s.sections.Range(func(key, value interface{}) bool {
		if sectionName, ok := key.(string); ok {
			sections = append(sections, sectionName)
			return true
		}
		err = fmt.Errorf("key is not of type string! type=%T, value=%v\n, err=%v", key, key, err)
		return true
	})
	return sections, err
}

func (s *store) AddSection(section string) (err error) {
	return os.MkdirAll(filepath.Join(s.dir, section), os.ModePerm)
}

type Store interface {
	GetSections() (sections []string, err error)
	AddSection(section string) (err error)
}

//
// func (s *store) Watch() error {
// 	watcher, err := fsnotify.NewWatcher()
// 	if err != nil {
// 		return err
// 	}
// 	defer watcher.Close()
//
// 	doneC := make(chan bool)
//
// 	err = filepath.Walk(s.dir, func(path string, fi fs.FileInfo, err error) error {
// 		// since fsnotify can watch all the files in a directory, watchers only need
// 		// to be added to each nested directory
// 		if fi.Mode().IsDir() {
// 			return watcher.Add(path)
// 		}
// 		return nil
// 	})
// 	if err != nil {
// 		fmt.Println("failed while walking dir.\nERROR ", err)
// 		return err
// 	}
//
// 	go func() {
// 		select {
// 		case event := <-watcher.Events:
// 			fmt.Printf("EVENT! %#v\n", event)
// 		case err := <-watcher.Errors:
// 			fmt.Printf("ERROR! %#v\n", err)
// 		}
// 	}()
//
// 	<-doneC
// }

func (s *store) Watch() error {
	w := watcher.New()

	w.SetMaxEvents(1)
	w.FilterOps(watcher.Write, watcher.Create, watcher.Remove)

	go func() {
		for {
			select {
			case event := <-w.Event:
				log.Debug().
					Str("dir", event.Name()).
					Str("path", event.Path).
					Stringer("op", event.Op).
					Str("value", event.String()).
					Interface("event", event).
					Msg("fs entry changed")

				switch event.Op {
				case watcher.Remove:
					log.Debug().Stringer("event", event).Str("name", event.Name()).Msg("remove section")
					s.todos.Delete(event.Name())
				case watcher.Create, watcher.Write:
					log.Debug().Stringer("event", event).Str("name", event.Name()).Stringer("op", event.Op).Msg("create / update section")
					s.todos.Store(event.Name(), event)
				case watcher.Rename:
					oldName := filepath.Base(event.OldPath)
					log.Debug().Stringer("event", event).Str("name", event.Name()).Str("oldName", oldName).Stringer("op", event.Op).Msg("rename section")
					s.todos.Delete(oldName)
					s.todos.Store(event.Name(), event)
				default:
					log.Debug().Stringer("event", event).Str("name", event.Name()).Stringer("op", event.Op).Msg("unsupported op")
				}
			case err := <-w.Error:
				log.Error().Err(err).Msg("error while watching")
			case <-w.Closed:
				log.Debug().Msg("closing watcher")
				return
			}
		}
	}()

	if err := w.AddRecursive(s.dir); err != nil {
		return err
	}

	err := w.Start(time.Millisecond * 100)
	if err != nil {
		fmt.Println("error while Start() ", err)
		w.Close()
		return err
	}

	return nil
}

func (s *store) SectionChanges() (watch *watcher.Watcher, err error) {

	w := watcher.New()

	w.SetMaxEvents(1)
	w.FilterOps(watcher.Write, watcher.Create, watcher.Remove)

	go func() {
		select {
		case event := <-w.Event:
			if event.IsDir() {
				log.Debug().
					Str("dir", event.Name()).
					Str("path", event.Path).
					Stringer("op", event.Op).
					Str("value", event.String()).
					Interface("event", event).
					Msg("directory changed")
			} else {
				log.Debug().
					Str("file", event.Name()).
					Str("path", event.Path).
					Stringer("op", event.Op).
					Str("value", event.String()).
					Interface("event", event).
					Msg("file changed")
			}
			switch event.Op {
			case watcher.Remove:
				log.Debug().Stringer("event", event).Str("name", event.Name()).Msg("remove section")
				s.todos.Delete(event.Name())
			case watcher.Create, watcher.Write:
				log.Debug().Stringer("event", event).Str("name", event.Name()).Stringer("op", event.Op).Msg("create / update section")
				s.todos.Store(event.Name(), event)
			case watcher.Rename:
				oldName := filepath.Base(event.OldPath)
				log.Debug().Stringer("event", event).Str("name", event.Name()).Str("oldName", oldName).Stringer("op", event.Op).Msg("rename section")
				s.todos.Delete(oldName)
				s.todos.Store(event.Name(), event)
			default:
				log.Debug().Stringer("event", event).Str("name", event.Name()).Stringer("op", event.Op).Msg("unsupported op")
			}
		case err := <-w.Error:
			log.Error().Err(err).Msg("error while watching")
		case <-w.Closed:
			log.Debug().Msg("closing watcher")
			return
		}
	}()

	if err := w.AddRecursive(s.dir); err != nil {
		return w, err
	}

	// Print a list of all of the files and folders currently
	// being watched and their paths.
	for path, f := range w.WatchedFiles() {
		log.Debug().Str("path", path).Str("file", f.Name()).Bool("isDir", f.IsDir()).Msg("watching fs entry")
	}

	// Close the watcher after watcher started.
	go func() {
		w.Wait()
		//	w.Close()
	}()

	return w, nil
}
