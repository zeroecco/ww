// Package shell contains the `ww shell` command implementation.
package shell

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
	"text/template"
	"time"

	"github.com/chzyer/readline"
	"github.com/urfave/cli/v2"
	"go.uber.org/fx"

	ww "github.com/wetware/ww/pkg"
	"github.com/wetware/ww/pkg/lang"
	"github.com/wetware/ww/pkg/lang/core"
)

const (
	wwpath         = "WW_PATH"
	defaultPath    = "~/.ww"
	bannerTemplate = `Wetware v{{.App.Version}}
Copyright {{.App.Copyright}}
Compiled with {{.GoVersion}} for {{.GOOS}}
`
)

var (
	flags = []cli.Flag{
		&cli.StringSliceFlag{
			Name:    "join",
			Aliases: []string{"j"},
			Usage:   "connect to cluster through specified peers",
			EnvVars: []string{"WW_JOIN"},
		},
		&cli.StringFlag{
			Name:    "discover",
			Aliases: []string{"d"},
			Usage:   "automatic peer discovery settings",
			Value:   "/mdns",
			EnvVars: []string{"WW_DISCOVER"},
		},
		&cli.BoolFlag{
			Name:    "dial",
			Usage:   "dial into a cluster using -join and -discover",
			EnvVars: []string{"WW_AUTODIAL"},
		},
		&cli.DurationFlag{
			Name:  "timeout",
			Usage: "timeout for -dial",
			Value: time.Second * 10,
		},
		&cli.StringFlag{
			Name:    "namespace",
			Aliases: []string{"ns"},
			Usage:   "cluster namespace (must match dial host)",
			Value:   "ww",
			EnvVars: []string{"WW_NAMESPACE"},
		},
		&cli.BoolFlag{
			Name:    "quiet",
			Aliases: []string{"q"},
			Usage:   "suppress banner message on interactive startup",
			EnvVars: []string{"WW_QUIET"},
		},
		&cli.StringSliceFlag{
			Name:        "path",
			Usage:       "location of ww source files",
			DefaultText: defaultPath,
			EnvVars:     []string{wwpath},
		},
		&cli.PathFlag{
			Name:    "appdata",
			Usage:   "local data directory",
			Value:   defaultPath,
			EnvVars: []string{"WW_APPDATA"},
		},
		// debug flags (hidden)
		&cli.BoolFlag{
			Name:   "log-fx",
			Usage:  "output fx dependency injection logs",
			Hidden: true,
		},
	}
)

// Command constructor
func Command() *cli.Command {
	return &cli.Command{
		Name:   "shell",
		Usage:  "start an interactive REPL session",
		Flags:  flags,
		Action: run(),
	}
}

func run() cli.ActionFunc {
	return func(c *cli.Context) error {
		app := fx.New(
			fxLogger(c),
			fx.StartTimeout(c.Duration("timeout")),
			fx.Supply(c,
				fx.Annotated{Name: "prompt", Target: " »"},
				fx.Annotated{Name: "multiline", Target: " ›"}),
			fx.Provide(
				newVM,
				newREPL,
				newInput,
				newPrinter,
				newPathData,
				newAnchorProvider),
			fx.Invoke(
				banner,
				loop))

		return app.Err()
	}
}

func banner(c *cli.Context) error {
	if c.Bool("quiet") {
		return nil
	}

	t, err := template.New("banner").Parse(bannerTemplate)
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, struct {
		*cli.Context
		GoVersion, GOOS string
	}{
		Context:   c,
		GoVersion: runtime.Version(),
		GOOS:      runtime.GOOS,
	}); err != nil {
		return err
	}

	_, err = fmt.Fprintln(c.App.Writer, buf.String())
	return err
}

type promptConfig struct {
	fx.In

	Prompt    string `name:"prompt"`
	Multiline string `name:"multiline"`
}

func loop(repl *REPL, p promptConfig) error {
	repl.CurrentNS = "ww" // TODO

	for repl.Init(); repl.More(); repl.Next() {
		ns := repl.CurrentNS
		line := p.Prompt
		if repl.LineNo > 1 {
			ns = strings.Repeat(" ", len(ns)+1)
			line = p.Multiline
		}

		repl.SetPrompt(fmt.Sprintf("%s%s ", ns, line))
	}

	return repl.Err()
}

func newVM(root AnchorProvider, src lang.PathProvider) lang.VM {
	return lang.VM{
		Env: lang.NewEnv(),
		Analyzer: &lang.Analyzer{
			Root: root.Anchor(),
			Src:  src,
		},
	}
}

type staticPathResolver struct {
	usr   *user.User
	paths []string
}

func (s staticPathResolver) Paths() lang.PathSet {
	ps := make(lang.PathSet, len(s.paths))
	for i, p := range s.paths {
		if p[0] == '~' {
			p = expandHomeDir(s.usr, p)
		}

		ps[i] = p
	}

	return ps
}

func expandHomeDir(usr *user.User, path string) string {
	return filepath.Clean(strings.Replace(path, "~", usr.HomeDir, 1))
}

type dynamicPathResolver struct{ usr *user.User }

func (d dynamicPathResolver) Paths() lang.PathSet {
	if paths := filepath.SplitList(os.Getenv(wwpath)); len(paths) != 0 {
		return staticPathResolver{usr: d.usr, paths: paths}.Paths()
	}

	return staticPathResolver{usr: d.usr, paths: []string{defaultPath}}.Paths()
}

type pathData struct {
	fx.Out

	User         *user.User
	PathProvider lang.PathProvider
}

func newPathData(c *cli.Context) (p pathData, err error) {
	if p.User, err = user.Current(); err != nil {
		return
	}

	p.PathProvider = dynamicPathResolver{p.User}
	if paths := c.StringSlice("path"); paths != nil {
		p.PathProvider = staticPathResolver{
			usr:   p.User,
			paths: c.StringSlice("path"),
		}
	}

	return
}

func newInput(c *cli.Context, usr *user.User, lx fx.Lifecycle) (*readline.Instance, error) {
	appdir := c.Path("appdata")
	if appdir[0] == '~' {
		appdir = expandHomeDir(usr, appdir)
	}

	r, err := readline.NewEx(&readline.Config{
		HistoryFile: filepath.Join(appdir, "__history__.ww"),
		Stdin:       c.App.Reader.(io.ReadCloser),
		Stdout:      c.App.Writer,
		Stderr:      c.App.ErrWriter,

		InterruptPrompt: "⏎",
		EOFPrompt:       "(exit)",

		/*
			TODO(enhancemenbt):  pass in the lang.Ww and configure autocomplete.
								 The lang.Ww instance will need to supply completions.
		*/
		// AutoComplete: completer(ww),
	})

	if err == nil {
		lx.Append(closehook(r))
	}

	return r, err
}

type printer struct{ Stdout, Stderr io.Writer }

func newPrinter(c *cli.Context) Printer {
	return printer{
		Stdout: c.App.Writer,
		Stderr: c.App.ErrWriter,
	}
}

func (p printer) Print(any ww.Any) error {
	s, err := core.Render(any)
	if err == nil {
		_, err = fmt.Fprintln(p.Stdout, s)
	}
	return err
}

func (p printer) PrintErr(err error) error {
	_, err = fmt.Fprintf(p.Stderr, "ERROR: %v\n", err)
	return err
}

func fxLogger(c *cli.Context) fx.Option {
	if c.Bool("log-fx") {
		return fx.Options()
	}

	return fx.NopLogger
}

func closehook(c io.Closer) fx.Hook {
	return fx.Hook{
		OnStop: func(context.Context) error {
			return c.Close()
		},
	}
}
