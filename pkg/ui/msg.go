package ui

import (
	"github.com/alex-held/todo/pkg/config"
)

type initMsg struct {
	Config []config.SectionConfig
}


type errMsg struct {
	error
}

func (e errMsg) Error() string { return e.error.Error() }


type todoRenderedMsg struct {
	sectionId int
	content   string
}
