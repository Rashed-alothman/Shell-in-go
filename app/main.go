package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

type Command struct {
	Type        string
	Description string
	Handler     func(string)
}

func main() {
	// TODO: implement a type to determine how the command would be interpreted if it were used.
	// Continue the TODO: it checks whether the command is a built-in command, an alias, or an external command ,an executable file, or unrecognized.

	// Define a map of command handlers: key is command name, value is a function that takes arguments as a string
	// This allows us to easily add new commands by simply adding new entries to the map without changing the main loop logic.
	var commands = make(map[string]Command)
	commands = map[string]Command{
		"exit": {
			Type:        "shell builtin",
			Description: "Exit the shell",
			Handler: func(args string) {
				os.Exit(0)
			},
		},
		"echo": {
			Type:        "shell builtin",
			Description: "Print text to the screen",
			Handler: func(args string) {
				fmt.Println(args)
			},
		},
		"pwd": {
			Type:        "shell builtin",
			Description: "Print the current working directory",
			Handler: func(args string) {
				cwd, err := os.Getwd()
				if err != nil {
					fmt.Fprintln(os.Stderr, "pwd:", err)
					return
				}
				fmt.Println(cwd)
			},
		},
		"help": {
			Type:        "shell builtin",
			Description: "Show all available commands",
			Handler: func(args string) {
				fmt.Println("Available commands:")
				keys := make([]string, 0, len(commands))
				for cmd := range commands {
					keys = append(keys, cmd)
				}
				sort.Strings(keys)
				for _, cmd := range keys {
					fmt.Printf("  %-10s %s\n", cmd, commands[cmd].Description)
				}
			},
		},
		"type": {
			Type:        "shell builtin",
			Description: "Describe how a command would be interpreted",
			Handler: func(args string) {
				name := strings.TrimSpace(args)
				if name == "" {
					fmt.Fprintln(os.Stderr, "type: missing argument")
					return
				}
				if cmd, ok := commands[name]; ok {
					fmt.Printf("%s is a %s\n", name, cmd.Type)
					return
				}
				pathEnv := os.Getenv("PATH")
				for dir := range strings.SplitSeq(pathEnv, string(os.PathListSeparator)) {
					fullPath := filepath.Join(dir, name)
					if info, err := os.Stat(fullPath); err == nil && !info.IsDir() && info.Mode()&0111 != 0 {

						fmt.Println(name + " is " + fullPath)
						return
					}
				}
				fmt.Println(name + ": not found")
			},
		},
		"ls": {
			Type:        "shell builtin",
			Description: "List directory contents",
			Handler: func(args string) {
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
					fmt.Println(file.Name())
				}
			},
		},
		"clear": {
			Type:        "shell builtin",
			Description: "Clear the terminal screen",
			Handler: func(args string) {
				fmt.Print("\033[H\033[2J")
			},
		},
		"cd": {
			Type:        "shell builtin",
			Description: "Change the current directory",
			Handler: func(args string) {
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
			Description: "Search for a pattern in a file",
			Handler: func(args string) {
				parts := strings.Fields(args)
				if len(parts) < 2 {
					fmt.Fprintln(os.Stderr, "grep: usage: grep <pattern> <file>")
					return
				}
				pattern := parts[0]
				filename := strings.Join(parts[1:], " ")

				re, err := regexp.Compile(pattern)
				if err != nil {
					fmt.Fprintln(os.Stderr, "grep: invalid pattern:", err)
					return
				}

				file, err := os.Open(filename)
				if err != nil {
					fmt.Fprintln(os.Stderr, "grep:", err)
					return
				}
				defer file.Close()

				scanner := bufio.NewScanner(file)
				for scanner.Scan() {
					line := scanner.Text()
					if re.MatchString(line) {
						fmt.Println(line)
					}
				}
				if err := scanner.Err(); err != nil {
					fmt.Fprintln(os.Stderr, "grep:", err)
				}
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

		// Trim the newline and any extra whitespace
		trimmed := strings.TrimSpace(command)

		// Split the input into command and arguments
		parts := strings.Fields(trimmed)
		if len(parts) == 0 {
			continue // Skip empty input
		}

		cmd := strings.ToLower(parts[0])
		args := strings.Join(parts[1:], " ")

		// Look up the command in the map
		if handler, ok := commands[cmd]; ok {
			handler.Handler(args)
		} else {
			fmt.Println(trimmed + ": command not found")
		}
	}
}
