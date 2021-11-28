package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/ghodss/yaml"
)

const (
	GH_CONFIG_DIR   = "GH_CONFIG_DIR"
	XDG_CONFIG_HOME = "XDG_CONFIG_HOME"
	XDG_STATE_HOME  = "XDG_STATE_HOME"
	XDG_DATA_HOME   = "XDG_DATA_HOME"
)

const dataDir = "todo"

func DataDir() string {
	switch runtime.GOOS {
	case "linux":
		return filepath.Join(os.Getenv(XDG_DATA_HOME), dataDir)
	case "darwin":
		homeDir, err := os.UserHomeDir()
		if err != nil {
			panic(err)
		}
		return filepath.Join(homeDir, "tmp", dataDir)
	default:
		panic("unsupported GOOS")
	}
}

type configError struct {
	err        error
	configFile string
}

func (c *configError) Error() string {
	msg := fmt.Sprintf(`Couldn't find a sections.yml configuration file.
Create one under: %s
Example of a sections.yml file:

  - title: Private
    tags:
      - home
      - calendar
      - buy
    filters: author:@me
  - title: Work
    tags:
      - work
  - title: Dev
    tags:
      - dev
      - go
      - gh

For more info, go to https://github.com/dlvhdr/gh-prs
press q to exit.
	
Original error: %v`, c.configFile, c.err)
	return msg
}

func ParseSectionConfig() ([]SectionConfig, error) {
	file := filepath.Join(DataDir(), "sections.yaml")
	data, err := os.ReadFile(file)

	if err != nil {
		return nil, &configError{configFile: file, err: err}
	}

	var sections []SectionConfig
	err = yaml.Unmarshal(data, sections)
	if err != nil {
		return sections, fmt.Errorf("failed parsing sections.yml: %w", err)
	}

	return sections, nil
}
