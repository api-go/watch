package main

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"apigo.cc/go/shell"
	"apigo.cc/go/watch"
)

const debounceTime = 300 * time.Millisecond

type preset struct {
	Types        []string
	ExcludePaths []string
}

var presets = map[string]preset{
	"go": {
		Types:        []string{"go"},
		ExcludePaths: []string{"vendor/**", "**/node_modules/**"},
	},
	"web": {
		Types:        []string{"html", "css", "js", "ts", "vue", "jsx", "tsx"},
		ExcludePaths: []string{"node_modules/**", "dist/**"},
	},
	"js": {
		Types:        []string{"js", "ts", "json"},
		ExcludePaths: []string{"node_modules/**"},
	},
	"py": {
		Types:        []string{"py", "toml", "yaml", "yml"},
		ExcludePaths: []string{"__pycache__/**", ".venv/**"},
	},
}

func main() {
	args := os.Args[1:]

	var paths []string
	var types []string
	var excludeTypes []string
	var excludePaths []string
	var command []string

	i := 0
	for i < len(args) {
		arg := args[i]

		switch arg {
		case "-h", "--help":
			printHelp()
			return
		case "-path":
			i++
			if i < len(args) {
				paths = append(paths, strings.Split(args[i], ",")...)
			}
		case "-type":
			i++
			if i < len(args) {
				types = append(types, strings.Split(args[i], ",")...)
			}
		case "-exclude-type":
			i++
			if i < len(args) {
				excludeTypes = append(excludeTypes, strings.Split(args[i], ",")...)
			}
		case "-exclude-path":
			i++
			if i < len(args) {
				excludePaths = append(excludePaths, args[i])
			}
		case "-go", "-web", "-js", "-py":
			name := strings.TrimPrefix(arg, "-")
			if p, ok := presets[name]; ok {
				types = append(types, p.Types...)
				excludePaths = append(excludePaths, p.ExcludePaths...)
			}
		default:
			// 第一个非内置参数，它及其后所有内容作为命令
			command = args[i:]
			i = len(args)
			continue
		}
		i++
	}

	if len(paths) == 0 {
		paths = []string{"."}
	}

	config := watch.Config{
		Paths:        paths,
		Types:        types,
		ExcludeTypes: excludeTypes,
		Excludes:     excludePaths,
		Debounce:     debounceTime,
	}

	if len(command) == 0 {
		runMonitor(config)
	} else {
		runCommand(config, command)
	}
}

func runMonitor(config watch.Config) {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	w, err := watch.Start(config, func(e *watch.Event) {
		fmt.Printf("%s: %s\n", e.Type, e.Path)
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, "watch error:", err)
		os.Exit(1)
	}
	defer w.Stop()

	<-sigCh
}

func runCommand(config watch.Config, command []string) {
	execDir, _ := os.Getwd()

	var proc *shell.Process
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	w, err := watch.Start(config, func(e *watch.Event) {
		fmt.Printf("\n>> %s: %s\n", e.Type, e.Path)
		if proc != nil {
			proc.Kill()
		}
		proc, _ = shell.Start(command[0], command[1:], &shell.Options{Dir: execDir, CatchSignal: true})
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, "watch error:", err)
		os.Exit(1)
	}
	defer w.Stop()

	// 启动时先执行一次
	proc, _ = shell.Start(command[0], command[1:], &shell.Options{Dir: execDir, CatchSignal: true})

	<-sigCh
	if proc != nil {
		proc.Kill()
	}
	os.Exit(0)
}

func printHelp() {
	fmt.Print(`Usage: watch [options] [command [args...]]

Watch file changes and execute command on change.
Without a command, just print change events to stdout.

Options:
  -path <paths>          Watch paths (comma-separated), default "."
  -type <types>          Include file types (comma-separated), e.g. -type go,js
  -exclude-type <types>  Exclude file types (comma-separated)
  -exclude-path <pattern> Exclude path pattern (repeatable, gitignore semantics)

Presets:
  -go     Go development       (go | vendor/**, **/node_modules/**)
  -web    Web frontend         (html,css,js,ts,vue,jsx,tsx | node_modules/**, dist/**)
  -js     Node.js development  (js,ts,json | node_modules/**)
  -py     Python development   (py,toml,yaml,yml | __pycache__/**, .venv/**)

Preset values merge with explicit options; they don't replace them.

Examples:
  watch -go go test ./...
  watch -web npm run build
  watch -py python main.py --dev
  watch -path ./src -type go go run .
  watch                   # print change events only
`)
}
