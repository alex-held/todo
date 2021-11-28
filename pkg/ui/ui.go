package ui

import (
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/paginator"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/rs/zerolog"

	"github.com/alex-held/todo/pkg/config"
	"github.com/alex-held/todo/pkg/utils"
)

type Todo struct {
	Title string
	Done  bool
}

type section struct {
	ID        int
	Config    config.SectionConfig
	Prs       []Todo
	Spinner   sectionSpinner
	Paginator paginator.Model
}

func (s *section) fetchSectionTodos() []tea.Cmd {
	var cmds []tea.Cmd
	panic("implement me!")
	/*
		for _, tag := range s.Config.Tags {

		tag := tag
			cmds = append(cmds, func() tea.Msg {

				fetched, err := fetchTodoFromTag(tag)

				if err != nil {
					return repoPullRequestsFetchedMsg{
						SectionId: section.Id,
						RepoName:  repo,
						Prs:       []PullRequest{},
					}
				}

				return repoPullRequestsFetchedMsg{
					SectionId: section.Id,
					RepoName:  repo,
					Prs:       fetched,
				}
			})
		}
	*/
	return cmds
}

type tickMsg struct {
	SectionId       int
	InternalTickMsg tea.Msg
}

func (section *section) Tick(spinnerTickCmd tea.Cmd) func() tea.Msg {
	return func() tea.Msg {
		return tickMsg{
			SectionId:       section.ID,
			InternalTickMsg: spinnerTickCmd(),
		}
	}
}

func fetchTodoFromTag(tag string) (Todo, error) {
	panic("implement me!")
}

type sectionSpinner struct {
	Model           spinner.Model
	NumReposFetched int
}

type Model struct {
	keys     utils.KeyMap
	err      error
	configs  []config.SectionConfig
	data     *[]section
	viewport viewport.Model
	cursor   cursor
	help     help.Model
	ready    bool
	logger   zerolog.Logger
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(initScreen, tea.EnterAltScreen)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case initMsg:
		m.configs = msg.Config
		var data []section
		for i, sectionConfig := range m.configs {
			s := spinner.Model{Spinner: spinner.Dot}
			data = append(data, section{
				ID:      i,
				Config:  sectionConfig,
				Spinner: sectionSpinner{Model: s, NumReposFetched: 0},
			})
		}
		m.data = &data
		return m, m.startFetchingSectionsData()
	case tea.WindowSizeMsg:
		m.help.Width = msg.Width
		verticalMargins := headerHeight + footerHeight

		if !m.ready {
			m.viewport = viewport.Model{
				Width:  msg.Width - 2*mainContentPadding,
				Height: msg.Height - verticalMargins,
			}
			m.ready = true

			// Render the viewport one line below the header.
			m.viewport.YPosition = headerHeight + 1
		} else {
			m.viewport.Width = msg.Width - 2*mainContentPadding
			m.viewport.Height = msg.Height - verticalMargins
		}

	case errMsg:
		m.err = msg
		return m, nil
	}

	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	if m.err != nil {
		return m.err.Error()
	}

	if m.configs == nil {
		return "Reading config...\n"
	}

	/*
	   paddedContentStyle := lipgloss.NewStyle().
	   		PaddingLeft(mainContentPadding).
	   		PaddingRight(mainContentPadding)
	*/

	s := strings.Builder{}
	panic("implement me!")

	/*
		s.WriteString(m.renderTabs())
		s.WriteString("\n")
		s.WriteString(paddedContentStyle.Render(m.renderTableHeader()))
		s.WriteString("\n")
		s.WriteString(paddedContentStyle.Render(m.viewport.View()))
		s.WriteString("\n")
		s.WriteString(lipgloss.PlaceVertical(2, lipgloss.Bottom, m.help.View(m.keys)))
	*/
	return s.String()
}

func (m Model) startFetchingSectionsData() tea.Cmd {
	var cmds []tea.Cmd
	for _, section := range *m.data {
		section := section
		cmds = append(cmds, section.fetchSectionTodos()...)
		cmds = append(cmds, section.Tick(spinner.Tick))
	}
	return tea.Batch(cmds...)
}

func initScreen() tea.Msg {
	sections, err := config.ParseSectionConfig()
	if err != nil {
		return errMsg{err}
	}

	return initMsg{Config: sections}
}

type cursor struct {
	currSectionId int
	currTodoId    int
}

func NewModel(logger zerolog.Logger) Model {
	m := Model{
		keys: utils.Keys,
		err:  nil,
		cursor: cursor{
			currSectionId: 0,
			currTodoId:    0,
		},
		help:   help.NewModel(),
		logger: logger,
	}

	return m
}
