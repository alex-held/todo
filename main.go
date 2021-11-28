package main

import (
	"context"
	"fmt"
	stdlog "log"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/google/go-github/v40/github"
	"github.com/olekukonko/tablewriter"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/alex-held/todo/cmd"
	"github.com/alex-held/todo/pkg/plugins"
	"github.com/alex-held/todo/pkg/ui"
)

var app = "todo"

type model struct {
	msg string
}

func (m model) Init() tea.Cmd {
	return func() tea.Msg {
		return m.msg
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, func() tea.Msg {
		return "updated"
	}
}

func (m model) View() string {
	mainContentPadding := 25
	paddedContentStyle := lipgloss.NewStyle().
		PaddingLeft(mainContentPadding).
		PaddingRight(mainContentPadding)

	return paddedContentStyle.Render("TEST")

}

func main() {

	out := zerolog.NewConsoleWriter(func(w *zerolog.ConsoleWriter) {
		w.Out, _ = os.Open("/Users/dev/tmp/todo.log")
		w.TimeFormat = ""
		w.PartsExclude = []string{zerolog.CallerFieldName, zerolog.TimestampFieldName}
		w.FormatLevel = func(i interface{}) string { return strings.ToUpper(fmt.Sprintf("[%-5s]", i)) }
		w.PartsOrder = []string{zerolog.LevelFieldName, zerolog.MessageFieldName}
	})
	logger := zerolog.New(out)
	stdlog.SetOutput(logger)
	stdlog.SetFlags(0)


	model := ui.NewModel(logger)
	tea.NewProgram(model, tea.WithAltScreen())

	return

	cobra.CheckErr(cmd.Root().Execute())

	return

	fmt.Printf("devctl todo\n")

	io := plugins.DefaultStreams()
	mgr := plugins.NewManager(io)
	if os.Args[1] == "run" {
		if _, err := mgr.Dispatch(os.Args[2:], io.In, io.Out, io.ErrOut); err != nil {
			panic(err)
		}
		return
	}

	extensions, err := mgr.List()
	if err != nil {
		panic(err)
	}

	var data = [][]string{}
	for i, extension := range extensions {
		data = append(data, []string{fmt.Sprintf("%d", i), extension.Name(), extension.Path(), fmt.Sprintf("%v", extension.IsLocal()), extension.Kind().Stringer()})
	}
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"ID", "Name", "Path", "Local", "Kind"})

	for _, v := range data {
		table.Append(v)
	}
	table.Render() // Send output}
}

type Todo struct {
	Title string
	Body  string
}

func getTodos() []Todo {

	client := github.NewClient(nil)
	issues, _, err := client.Issues.ListByRepo(context.Background(), "alex-held", "todo", &github.IssueListByRepoOptions{
		Milestone:   "",
		State:       "",
		Assignee:    "",
		Creator:     "",
		Mentioned:   "",
		Labels:      nil,
		Sort:        "",
		Direction:   "",
		Since:       time.Time{},
		ListOptions: github.ListOptions{},
	})
	if err != nil {
		log.Printf("ERROR! %v\n", err)
		panic(err)
	}
	for i, issue := range issues {
		fmt.Printf("[%d] %v", i, *issue.Title)
	}

	return nil
}
