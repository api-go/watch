package main

import (
	"bufio"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"apigo.cc/go/shell"
	"apigo.cc/go/watch"
)

// ── ANSI 颜色常量 ──────────────────────────────────────────────────────────────

const (
	reset  = "\033[0m"
	bold   = "\033[1m"
	dim    = "\033[2m"
	red    = "\033[31m"
	green  = "\033[32m"
	yellow = "\033[33m"
	blue   = "\033[34m"
	cyan   = "\033[36m"
	white  = "\033[37m"
	gray   = "\033[90m"

	bgBlue  = "\033[44m"
	bgGreen = "\033[42m"
)

const debounceTime = 300 * time.Millisecond

// ── Preset 定义 ────────────────────────────────────────────────────────────────

type preset struct {
	Types          []string
	ExcludePaths   []string
	DefaultCommand []string
}

var presets = map[string]preset{
	"go": {
		Types:          []string{"go"},
		ExcludePaths:   []string{"vendor/**", "**/node_modules/**"},
		DefaultCommand: []string{"go", "run", "."},
	},
	"gotest": {
		Types:          []string{"go"},
		ExcludePaths:   []string{"vendor/**", "**/node_modules/**"},
		DefaultCommand: []string{"go", "test", "-v", "-count=1", "./..."},
	},
	"web": {
		Types:          []string{"html", "css", "js", "ts", "vue", "jsx", "tsx"},
		ExcludePaths:   []string{"node_modules/**", "dist/**"},
		DefaultCommand: []string{"npm", "run", "dev"},
	},
	"js": {
		Types:          []string{"js", "ts", "json"},
		ExcludePaths:   []string{"node_modules/**"},
		DefaultCommand: []string{"npm", "start"},
	},
	"py": {
		Types:          []string{"py", "toml", "yaml", "yml"},
		ExcludePaths:   []string{"__pycache__/**", ".venv/**"},
		DefaultCommand: []string{"python", "main.py"},
	},
}

// ── 入口 ───────────────────────────────────────────────────────────────────────

func main() {
	args := os.Args[1:]
	// Built-in control options must be handled before loading .watch; otherwise
	// a task definition can consume -h as task arguments.
	if hasHelpFlag(args) {
		printHelp()
		return
	}
	if handleReload(args) {
		return
	}
	args = tryLoadWatchFile(args)

	if len(args) == 0 {
		printHelp()
		return
	}

	var paths []string
	var types []string
	var excludeTypes []string
	var excludePaths []string
	var command []string
	var clearScreen bool

	// 自动读取 .gitignore 和 .ignore
	excludePaths = append(excludePaths, parseIgnoreFiles()...)

	i := 0
	for i < len(args) {
		arg := args[i]

		switch {
		case arg == "-h" || arg == "--help":
			printHelp()
			return
		case arg == "-c" || arg == "--clear":
			clearScreen = true
		case arg == "-w" || arg == "--watch":
			i++
			if i < len(args) {
				paths = append(paths, strings.Split(args[i], ",")...)
			}
		case arg == "-e" || arg == "--ext":
			i++
			if i < len(args) {
				types = append(types, strings.Split(args[i], ",")...)
			}
		case arg == "-E" || arg == "--exclude-ext":
			i++
			if i < len(args) {
				excludeTypes = append(excludeTypes, strings.Split(args[i], ",")...)
			}
		case arg == "-i" || arg == "--ignore":
			i++
			if i < len(args) {
				excludePaths = append(excludePaths, args[i])
			}
		case strings.HasPrefix(arg, "--") && presets[strings.TrimPrefix(arg, "--")].Types != nil:
			name := strings.TrimPrefix(arg, "--")
			p := presets[name]
			types = append(types, p.Types...)
			excludePaths = append(excludePaths, p.ExcludePaths...)
			if len(command) == 0 {
				command = append([]string{}, p.DefaultCommand...) // 设定为预设默认命令
			}
		default:
			// 如果走到这里，说明是遇到了位置参数（command 或 参数）
			if len(command) > 0 { // 已经有预设的默认命令
				// 判断输入的是否像是一个新命令（例如 `go`, `npm`, `python` 等）
				base := args[i]
				if idx := strings.LastIndexAny(base, "/\\"); idx >= 0 {
					base = base[idx+1:]
				}
				if base == "go" || base == "go.exe" || base == "npm" || base == "npm.cmd" || base == "python" || base == "python.exe" || base == "node" || base == "node.exe" {
					command = args[i:] // 替换为用户显式输入的命令
				} else {
					// 否则将其视为预设命令的参数，智能追加
					// 针对 go run . 自动追加 -- 避免参数被 go run 吃掉
					if command[0] == "go" && len(command) >= 3 && command[1] == "run" {
						command = append(command, "--")
					}
					command = append(command, args[i:]...)
				}
			} else {
				command = args[i:] // 没有预设，直接使用
			}
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
	pidPath := filepath.Join(".", "watch.pid")
	if err := os.WriteFile(pidPath, []byte(fmt.Sprintf("%d\n", os.Getpid())), 0644); err != nil {
		logError("failed to write watch.pid: " + err.Error())
	} else {
		defer os.Remove(pidPath)
	}

	printBanner(config, command, clearScreen)

	if len(command) == 0 {
		runMonitor(config)
	} else {
		runCommand(config, command, clearScreen)
	}
}

func hasHelpFlag(args []string) bool {
	for _, arg := range args {
		if arg == "-h" || arg == "--help" {
			return true
		}
	}
	return false
}

// handleReload sends SIGUSR1 to the process recorded in a PID file.
func handleReload(args []string) bool {
	for i, arg := range args {
		if arg != "-r" && arg != "--reload" {
			continue
		}
		pidPath := "watch.pid"
		if i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
			pidPath = args[i+1]
		}
		data, err := os.ReadFile(pidPath)
		if err != nil {
			if !os.IsNotExist(err) {
				logError("failed to read PID file: " + err.Error())
			}
			return true
		}
		var pid int
		if _, err = fmt.Sscanf(strings.TrimSpace(string(data)), "%d", &pid); err != nil || pid <= 0 {
			logError("invalid PID in " + pidPath)
			return true
		}
		if err = syscall.Kill(pid, syscall.SIGUSR1); err != nil {
			logError(fmt.Sprintf("failed to reload process %d: %v", pid, err))
		} else {
			fmt.Printf("sent reload signal to process %d (%s)\n", pid, pidPath)
		}
		return true
	}
	return false
}

// tryLoadWatchFile 尝试从 .watch 读取配置并处理 --task
func tryLoadWatchFile(args []string) []string {
	contentBytes, err := os.ReadFile(".watch")
	if err != nil {
		return args
	}
	content := string(contentBytes)

	tasks := make(map[string]string)
	var firstTask string

	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			k := strings.TrimSpace(parts[0])
			v := strings.TrimSpace(parts[1])
			tasks[k] = v
			if firstTask == "" {
				firstTask = k
			}
		}
	}

	if len(tasks) == 0 {
		return args
	}

	targetTask := firstTask // 默认第一个
	if len(args) > 0 && strings.HasPrefix(args[0], "--") {
		taskName := strings.TrimPrefix(args[0], "--")
		if _, ok := tasks[taskName]; ok && presets[taskName].Types == nil {
			targetTask = taskName
			args = args[1:] // 消费掉这个参数
		}
	}

	if cmdStr, ok := tasks[targetTask]; ok {
		fmt.Printf("%s> Using task from .watch: --%s (%s)%s\n", dim, targetTask, cmdStr, reset)
		// 简单按空格切割组合
		taskArgs := strings.Fields(cmdStr)
		return append(taskArgs, args...)
	}

	return args
}

// parseIgnoreFiles 解析当前目录的忽略文件
func parseIgnoreFiles() []string {
	var ignores []string
	files := []string{".gitignore", ".ignore"}
	for _, f := range files {
		contentBytes, err := os.ReadFile(f)
		if err == nil {
			content := string(contentBytes)
			lines := strings.Split(content, "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if line != "" && !strings.HasPrefix(line, "#") {
					ignores = append(ignores, line)
				}
			}
		}
	}
	return ignores
}

// ── 监控模式（无命令）──────────────────────────────────────────────────────────

func runMonitor(config watch.Config) {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	w, err := watch.Start(config, func(e *watch.Event) {
		printEvent(e)
	})
	if err != nil {
		logError("failed to start watcher: " + err.Error())
		os.Exit(1)
	}
	defer w.Stop()

	<-sigCh
	fmt.Printf("\n%s%s Stopped.%s\n", dim, gray, reset)
}

// ── 命令模式 ───────────────────────────────────────────────────────────────────

func runCommand(config watch.Config, command []string, clearScreen bool) {
	execDir, _ := os.Getwd()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGUSR1)
	restartCh := make(chan *watch.Event, 1)
	stopCh := make(chan struct{})
	doneCh := make(chan struct{})

	requestRestart := func(e *watch.Event) {
		select {
		case restartCh <- e:
		default:
			// A restart is already pending; one restart will include all recent changes.
		}
	}

	w, err := watch.Start(config, requestRestart)
	if err != nil {
		logError("failed to start watcher: " + err.Error())
		os.Exit(1)
	}
	defer w.Stop()

	// 监听控制台输入 `rs` 手动重启
	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			text := strings.TrimSpace(scanner.Text())
			if text == "rs" {
				requestRestart(&watch.Event{Path: "manual restart", Type: watch.Change})
			}
		}
	}()

	// A single worker owns proc so file events, rs, and signals cannot restart concurrently.
	go func() {
		defer close(doneCh)
		var proc *shell.Process
		start := func() {
			printRestart(command)
			var err error
			proc, err = shell.Start(command[0], command[1:], &shell.Options{Dir: execDir, CatchSignal: true})
			if err != nil {
				proc = nil
				logError("failed to start command: " + err.Error())
			}
		}
		if clearScreen {
			fmt.Print("\033[2J\033[H")
		}
		start()
		for {
			select {
			case e := <-restartCh:
				if clearScreen {
					fmt.Print("\033[2J\033[H")
				}
				printEvent(e)
				if proc != nil {
					if err := proc.Kill(); err != nil {
						logError("failed to stop command: " + err.Error())
						continue
					}
				}
				start()
			case <-stopCh:
				if proc != nil {
					if err := proc.Kill(); err != nil {
						logError("failed to stop command: " + err.Error())
					}
				}
				return
			}
		}
	}()

	for sig := range sigCh {
		if sig == syscall.SIGUSR1 {
			requestRestart(&watch.Event{Path: "manual reload", Type: watch.Change})
			continue
		}
		break
	}
	close(stopCh)
	<-doneCh
	fmt.Printf("\n%s%s Stopped.%s\n", dim, gray, reset)
}

// ── 输出函数 ───────────────────────────────────────────────────────────────────

func printBanner(config watch.Config, command []string, clearScreen bool) {
	var parts []string

	parts = append(parts, fmt.Sprintf("%spaths:%s %s", dim, reset, strings.Join(config.Paths, ",")))

	if len(config.Types) > 0 {
		types := make([]string, len(config.Types))
		for i, t := range config.Types {
			if !strings.HasPrefix(t, ".") {
				t = "." + t
			}
			types[i] = t
		}
		parts = append(parts, fmt.Sprintf("%sexts:%s %s", dim, reset, strings.Join(types, ",")))
	}

	if len(config.ExcludeTypes) > 0 {
		exTypes := make([]string, len(config.ExcludeTypes))
		for i, t := range config.ExcludeTypes {
			if !strings.HasPrefix(t, ".") {
				t = "." + t
			}
			exTypes[i] = t
		}
		parts = append(parts, fmt.Sprintf("%s-exts:%s %s", dim, reset, strings.Join(exTypes, ",")))
	}

	if len(config.Excludes) > 0 {
		showCount := 2
		if len(config.Excludes) > showCount {
			parts = append(parts, fmt.Sprintf("%signore:%s %s...", dim, reset, strings.Join(config.Excludes[:showCount], ",")))
		} else {
			parts = append(parts, fmt.Sprintf("%signore:%s %s", dim, reset, strings.Join(config.Excludes, ",")))
		}
	}

	if len(command) > 0 {
		parts = append(parts, fmt.Sprintf("%scmd:%s %s%s%s", dim, reset, bold, strings.Join(command, " "), reset))
	}

	fmt.Printf("\n%s%swatch%s ❖ %s\n", bold, cyan, reset, strings.Join(parts, fmt.Sprintf(" %s|%s ", dim, reset)))
	fmt.Printf("%s(Type `rs` then Enter to restart manually)%s\n\n", gray, reset)
}

func printEvent(e *watch.Event) {
	ts := time.Now().Format("15:04:05")
	typeStr, typeColor := eventStyle(e.Type)

	dirMark := ""
	if e.IsDir {
		dirMark = gray + "  [dir]" + reset
	}

	fmt.Printf("%s%s%s  %s%s%s%s  %s%s%s\n",
		gray, ts, reset,
		bold, typeColor, typeStr, reset,
		white, e.Path, reset+dirMark)
}

func printRestart(command []string) {
	ts := time.Now().Format("15:04:05")
	cmdStr := strings.Join(command, " ")
	line := strings.Repeat("─", 48)
	fmt.Printf("\n%s%s%s  %s%s$ %s%s%s\n",
		gray, line, reset,
		gray, ts, reset,
		bold+green, cmdStr+reset)
}

func logError(msg string) {
	fmt.Fprintf(os.Stderr, "%s%s✗ %s%s\n", bold, red, msg, reset)
}

func eventStyle(et watch.EventType) (string, string) {
	switch et {
	case watch.Create:
		return "create", green
	case watch.Change:
		return "change", blue
	case watch.Remove:
		return "remove", red
	case watch.Rename:
		return "rename", yellow
	default:
		return string(et), white
	}
}

// ── 帮助信息 ───────────────────────────────────────────────────────────────────

func printHelp() {
	fmt.Print(`
` + bold + bgBlue + `  watch  ` + reset + `  File watcher & command runner
` + dim + `  Recursive by default. Respects .gitignore automatically.` + reset + `

` + bold + `Usage` + reset + `
  watch [options] [command [args...]]
  watch                  # execute default task in .watch or print help

` + bold + `Options` + reset + `
  ` + cyan + `-w, --watch <path>` + reset + `     Watch paths (comma-separated), default "."
  ` + cyan + `-e, --ext <exts>` + reset + `       Include extensions (comma-separated), e.g., ` + yellow + `go,ts` + reset + `
  ` + cyan + `-E, --exclude-ext <exts>` + reset + ` Exclude extensions (comma-separated)
  ` + cyan + `-i, --ignore <pattern>` + reset + ` Ignore glob patterns (repeatable), e.g., ` + yellow + `node_modules/**` + reset + `
  ` + cyan + `-c, --clear` + reset + `            Clear screen before restarting command
  ` + cyan + `-r, --reload [pidfile]` + reset + `  Send SIGUSR1 to the PID in watch.pid (or a specified file)
  ` + cyan + `-h, --help` + reset + `             Show this help

` + bold + `Presets` + reset + ` (Auto-injects default commands if no command provided)
  ` + green + `--go` + reset + `       Go dev       — ext: ` + dim + `go` + reset + ` | cmd: ` + dim + `go run .` + reset + `
  ` + green + `--gotest` + reset + `   Go test      — ext: ` + dim + `go` + reset + ` | cmd: ` + dim + `go test -v -count=1 ./...` + reset + `
  ` + green + `--web` + reset + `      Web frontend — ext: ` + dim + `html,css,js,ts,vue,jsx...` + reset + ` | cmd: ` + dim + `npm run dev` + reset + `
  ` + green + `--js` + reset + `       Node.js dev  — ext: ` + dim + `js,ts,json` + reset + ` | cmd: ` + dim + `npm start` + reset + `
  ` + green + `--py` + reset + `       Python dev   — ext: ` + dim + `py,toml,yaml...` + reset + ` | cmd: ` + dim + `python main.py` + reset + `

` + bold + `Tasks File (.watch)` + reset + `
  You can create a .watch file with key=value format:
  ` + dim + `dev = --clear -w src -e go go run .` + reset + `
  ` + dim + `test = --gotest` + reset + `
  Run ` + yellow + `watch --dev` + reset + ` to execute a task, or just ` + yellow + `watch` + reset + ` for the first task.

` + bold + `Interactive` + reset + `
  Type ` + yellow + `rs` + reset + ` and press Enter while running to force a manual restart.
`)
}
