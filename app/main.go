package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strings"
	"sync"
)

type Command struct {
	Type        string
	Description string
	Handler     func(string, io.Reader, io.Writer)
}

type stageInfo struct {
	handler Command
	args    string
}

func runPipeline(input string, commands map[string]Command) {
	stages := strings.Split(input, " | ")
	if len(stages) < 2 {
		fmt.Fprintln(os.Stderr, "Usage: <cmd1> [args] | <cmd2> [args] | ...")
		return
	}

	parsed := make([]stageInfo, 0, len(stages))
	for i, stage := range stages {
		stage = strings.TrimSpace(stage)
		parts := strings.Fields(stage)
		if len(parts) == 0 {
			fmt.Fprintf(os.Stderr, "empty stage %d\n", i)
			return
		}
		cmdName := strings.ToLower(parts[0])
		cmdArgs := ""
		if len(parts) > 1 {
			cmdArgs = strings.Join(parts[1:], " ")
		}
		handler, ok := commands[cmdName]
		if !ok {
			fmt.Fprintf(os.Stderr, "error: command not found: %s\n", cmdName)
			return
		}
		parsed = append(parsed, stageInfo{handler, cmdArgs})
	}

	readers := make([]*os.File, len(parsed)-1)
	writers := make([]*os.File, len(parsed)-1)
	for i := range readers {
		r, w, err := os.Pipe()
		if err != nil {
			fmt.Fprintf(os.Stderr, "pipe error: %v\n", err)
			return
		}
		readers[i] = r
		writers[i] = w
	}

	var wg sync.WaitGroup
	for i, s := range parsed {
		wg.Add(1)

		var in io.Reader
		var out io.Writer

		if i == 0 {
			in = os.Stdin
		} else {
			in = readers[i-1]
		}

		if i == len(parsed)-1 {
			out = os.Stdout
		} else {
			out = writers[i]
		}

		go func(h Command, a string, r io.Reader, w io.Writer, idx int) {
			defer wg.Done()
			h.Handler(a, r, w)
			if idx < len(writers) {
				writers[idx].Close()
			}
		}(s.handler, s.args, in, out, i)
	}

	wg.Wait()
}

func main() {
	var commands map[string]Command
	commands = map[string]Command{
		"exit": {
			Type:        "shell builtin",
			Description: "Exit the shell",
			Handler: func(args string, in io.Reader, out io.Writer) {
				os.Exit(0)
			},
		},
		"echo": {
			Type:        "shell builtin",
			Description: "Print text to the screen",
			Handler: func(args string, in io.Reader, out io.Writer) {
				fmt.Fprintln(out, args)
			},
		},
		"pwd": {
			Type:        "shell builtin",
			Description: "Print the current working directory",
			Handler: func(args string, in io.Reader, out io.Writer) {
				cwd, err := os.Getwd()
				if err != nil {
					fmt.Fprintln(os.Stderr, "pwd:", err)
					return
				}
				fmt.Fprintln(out, cwd)
			},
		},
		"help": {
			Type:        "shell builtin",
			Description: "Show all available commands",
			Handler: func(args string, in io.Reader, out io.Writer) {
				fmt.Fprintln(out, "Available commands:")
				keys := make([]string, 0, len(commands))
				for cmd := range commands {
					keys = append(keys, cmd)
				}
				sort.Strings(keys)
				for _, cmd := range keys {
					fmt.Fprintf(out, "  %-10s %s\n", cmd, commands[cmd].Description)
				}
			},
		},
		"type": {
			Type:        "shell builtin",
			Description: "Describe how a command would be interpreted",
			Handler: func(args string, in io.Reader, out io.Writer) {
				name := strings.TrimSpace(args)
				if name == "" {
					fmt.Fprintln(os.Stderr, "type: missing argument")
					return
				}
				if cmd, ok := commands[name]; ok {
					fmt.Fprintf(out, "%s is a %s\n", name, cmd.Type)
					return
				}
				if path, err := exec.LookPath(name); err == nil {
					fmt.Fprintln(out, name+" is "+path)
					return
				}
				fmt.Fprintln(out, name+": not found")
			},
		},
		"ls": {
			Type:        "shell builtin",
			Description: "List directory contents",
			Handler: func(args string, in io.Reader, out io.Writer) {
				dir := "."
				if args != "" {
					dir = args
				}
				files, err := os.ReadDir(dir)
				if err != nil {
					fmt.Fprintln(os.Stderr, "ls:", err)
					return
				}
				for _, file := range files {
					fmt.Fprintln(out, file.Name())
				}
			},
		},
		"clear": {
			Type:        "shell builtin",
			Description: "Clear the terminal screen",
			Handler: func(args string, in io.Reader, out io.Writer) {
				fmt.Fprint(out, "\033[H\033[2J")
			},
		},
		"cd": {
			Type:        "shell builtin",
			Description: "Change the current directory",
			Handler: func(args string, in io.Reader, out io.Writer) {
				if args == "" {
					home, err := os.UserHomeDir()
					if err != nil {
						fmt.Fprintln(os.Stderr, "cd: cannot find home directory:", err)
						return
					}
					args = home
				}
				if err := os.Chdir(args); err != nil {
					fmt.Fprintln(os.Stderr, "cd:", err)
				}
			},
		},
		"grep": {
			Type:        "shell builtin",
			Description: "Search for a pattern in a file or stdin",
			Handler: func(args string, in io.Reader, out io.Writer) {
				parts := strings.Fields(args)
				if len(parts) == 0 {
					fmt.Fprintln(os.Stderr, "grep: usage: grep <pattern> [file]")
					return
				}
				pattern := parts[0]
				re, err := regexp.Compile(pattern)
				if err != nil {
					fmt.Fprintln(os.Stderr, "grep: invalid pattern:", err)
					return
				}

				// No file — read from pipe/stdin
				var reader io.Reader
				if len(parts) < 2 {
					reader = in
				} else {
					f, err := os.Open(strings.Join(parts[1:], " "))
					if err != nil {
						fmt.Fprintln(os.Stderr, "grep:", err)
						return
					}
					defer f.Close()
					reader = f
				}

				scanner := bufio.NewScanner(reader)
				for scanner.Scan() {
					line := scanner.Text()
					if re.MatchString(line) {
						fmt.Fprintln(out, line)
					}
				}
			},
		},
		"del": {
			Type:        "builtin",
			Description: "Delete a file",
			Handler: func(args string, in io.Reader, out io.Writer) {
				if args == "" {
					fmt.Fprintln(out, "Usage: del <filename>")
					return
				}
				if err := os.Remove(args); err != nil {
					fmt.Fprintln(os.Stderr, "del:", err)
				}
			},
		},
		"mkfile": {
			Type:        "builtin",
			Description: "Create a new file",
			Handler: func(args string, in io.Reader, out io.Writer) {
				if args == "" {
					fmt.Fprintln(out, "Usage: mkfile <filename>")
					return
				}
				if _, err := os.Create(args); err != nil {
					fmt.Fprintln(os.Stderr, "mkfile:", err)
				}
			},
		},
		"mkdir": {
			Type:        "builtin",
			Description: "Create a new directory",
			Handler: func(args string, in io.Reader, out io.Writer) {
				if args == "" {
					fmt.Fprintln(out, "Usage: mkdir <dirname>")
					return
				}
				if err := os.Mkdir(args, 0755); err != nil {
					fmt.Fprintln(os.Stderr, "mkdir:", err)
				}
			},
		},
		"cat": {
			Type:        "builtin",
			Description: "Print file contents",
			Handler: func(args string, in io.Reader, out io.Writer) {
				if args == "" {
					fmt.Fprint(os.Stderr, "Usage: cat <filename>")
					return
				}
				data, err := os.ReadFile(args)
				if err != nil {
					fmt.Fprint(os.Stderr, "cat:", err)
					return
				}
				fmt.Fprint(out, string(data))
			},
		},
	}

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("$ ")

		command, err := reader.ReadString('\n')
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error reading input:", err)
			os.Exit(1)
		}

		trimmed := strings.TrimSpace(command)

		if strings.Contains(trimmed, " | ") {
			runPipeline(trimmed, commands)
			continue
		}

		parts := strings.Fields(trimmed)
		if len(parts) == 0 {
			continue
		}

		cmd := strings.ToLower(parts[0])
		args := strings.Join(parts[1:], " ")

		if handler, ok := commands[cmd]; ok {
			handler.Handler(args, os.Stdin, os.Stdout)
		} else {
			if path, err := exec.LookPath(cmd); err == nil {
				extCmd := exec.Command(path, parts[1:]...)
				extCmd.Stdin = os.Stdin
				extCmd.Stdout = os.Stdout
				extCmd.Stderr = os.Stderr
				if err := extCmd.Run(); err != nil {
					fmt.Fprintln(os.Stderr, err)
				}
			} else {
				fmt.Fprintln(os.Stderr, trimmed+": command not found")
			}
		}
	}
}
