package plugins

import (
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/alex-held/todo/pkg/config"
)

func DefaultStreams() *IOStreams {
	return &IOStreams{
		In:     os.Stdin,
		Out:    os.Stdout,
		ErrOut: os.Stderr,
	}
}

type IOStreams struct {
	In     io.ReadCloser
	Out    io.Writer
	ErrOut io.Writer
}

type manager struct {
	dataDirFn func() string
	newCmdFn  func(cmd string, args ...string) *exec.Cmd
	io        *IOStreams
}

type Kind int

const (
	BinaryKind Kind = iota
	GitKind
)

type Extension struct {
	name  string
	path  string
	local bool
	kind  Kind
}

func (e Extension) Name() string {
	return strings.TrimPrefix(e.name, pluginPrefix)
}
func (e Extension) Path() string {
	return e.path
}
func (e Extension) IsLocal() bool {
	return e.local
}
func (e Extension) Kind() Kind {
	return e.kind
}

func (k Kind) Stringer() string {
	switch k {
	case BinaryKind:
		return "binary"
	case GitKind:
		return "git"
	default:
		panic(fmt.Errorf("unknown kind %v", k))
	}
}
func NewManager(io *IOStreams) *manager {
	m := &manager{
		dataDirFn: config.DataDir,
		newCmdFn:  exec.Command,
		io:        io,
	}
	return m
}

const pluginPrefix = "devctl-"
const manifestName = "manifest.yaml"

func (m *manager) installDir() string {
	return filepath.Join(m.dataDirFn(), "plugins")
}

func (m *manager) Dispatch(args []string, stdin io.Reader, stdout, stderr io.Writer) (bool, error) {
	if len(args) == 0 {
		return false, fmt.Errorf("not enough arguments")
	}

	var exe string
	// var ext Extension
	extName := args[0]
	forwardArgs := args[1:]

	fmt.Printf("ExtName: %s\n", extName)
	extensions, _ := m.List()
	for _, e := range extensions {
		if e.Name() == extName {
			exe = e.Path()
			//	ext = e
			break
		}
	}
	if exe == "" {
		return false, nil
	}

	cmd := m.newCmdFn(exe, forwardArgs...)
	cmd.Stdin = stdin
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	return true, cmd.Run()
}

func (m *manager) List() ([]Extension, error) {
	installDir := m.installDir()
	entries, err := ioutil.ReadDir(installDir)
	println("install dir: " + installDir)
	if err != nil {
		return nil, err
	}

	var results []Extension
	for _, f := range entries {
		if !strings.HasPrefix(f.Name(), pluginPrefix) {
			continue
		}
		var ext Extension
		var err error
		if f.IsDir() {
			ext, err = m.parseExtensionDir(f)
			if err != nil {
				return nil, err
			}
			results = append(results, ext)
		}
	}

	return results, nil

}

func (m *manager) parseExtensionDir(fi fs.FileInfo) (Extension, error) {
	id := m.installDir()
	if _, err := os.Stat(filepath.Join(id, fi.Name(), manifestName)); err == nil {
		return m.parseBinaryExtensionDir(fi)
	}
	return Extension{}, fmt.Errorf("unable to parse extension at %s", fi.Name())
}

func (m *manager) parseBinaryExtensionDir(fi fs.FileInfo) (Extension, error) {
	id := m.installDir()
	exePath := filepath.Join(id, fi.Name(), fi.Name())
	ext := Extension{
		name:  fi.Name(),
		path:  exePath,
		local: true,
		kind:  BinaryKind,
	}
	return ext, nil
}
