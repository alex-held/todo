package store

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alex-held/todo/pkg/utils"
)

func TestStore_Watch2(t *testing.T) {

	tests := []struct {
		name         string
		existingDirs []string
		createDirs   []string
		deleteDirs   []string
		modifyDirs   []string
		expected     []string
	}{
		{
			name:         "only existing",
			existingDirs: []string{"a", "b", "c"},
			createDirs:   nil,
			deleteDirs:   nil,
			modifyDirs:   nil,
			expected:     []string{"a", "b", "c"},
		},
		{
			name:         "existing and create",
			existingDirs: []string{"a", "b", "c"},
			createDirs:   []string{"d"},
			deleteDirs:   nil,
			modifyDirs:   nil,
			expected:     []string{"a", "b", "c", "d"},
		},
		{
			name:         "existing and create and delete",
			existingDirs: []string{"a", "b", "c"},
			createDirs:   []string{"d"},
			deleteDirs:   []string{"a", "d"},
			expected:     []string{"b", "c"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			utils.SetupGlobalLogger(os.Stdout)
			tmpDir, err := os.MkdirTemp("", "todo_store_watch")
			require.NoError(t, err)
			defer os.RemoveAll(tmpDir)

			for _, existing := range tt.existingDirs {
				err = os.MkdirAll(filepath.Join(tmpDir, existing), os.ModePerm)
				require.NoError(t, err)
			}

			sut := NewStore(tmpDir)

			for _, create := range tt.createDirs {
				err = os.MkdirAll(filepath.Join(tmpDir, create), os.ModePerm)
				require.NoError(t, err)
			}

			for _, del := range tt.deleteDirs {
				err = os.RemoveAll(filepath.Join(tmpDir, del))
				require.NoError(t, err)
			}

			time.Sleep(150 * time.Millisecond)
			sections, err := sut.GetSections()

			for i, section := range sections {
				log.Info().Int("#", i).Str("section", section).Msg(" ")
			}
			assert.NoError(t, err)
			assert.ElementsMatch(t, sections, tt.expected)
		})

	}

}

//
// func TestFSWatcher_Watch(t *testing.T) {
// 	utils.SetupGlobalLogger(zerolog.DebugLevel)
//
// 	//	sut := NewStore("/Users/dev/tmp/todo/store_test")
//
// 	fchan := make(chan notify.EventInfo, 1)
// 	sut := FSWatcher{
// 		FChan: fchan,
// 		Paths: []string{"/Users/dev/tmp/todo/store_test"},
// 	}
// 	sut.FSWatcherStart()
//
// 	var sections []string
// 	// Process events
// 	go func() {
// 		for {
// 			select {
// 			case ev := <-fchan:
// 				fmt.Printf("watched '%s' event at '%s'\n", ev.Event(), ev.Path())
// 				sections = append(sections, ev.Path())
// 			}
// 		}
// 	}()
//
// 	time.Sleep(5 * time.Second)
// 	assert.Len(t, sections, 3)
// }
