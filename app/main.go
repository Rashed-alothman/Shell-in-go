package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
)

func main() {
	// TODO: immplement REPL loop
	//Read: Display a prompt and wait for user input ..status:done
	//Evaluate: Parse the input and execute the command ..status:done
	//Print: Display the result of the command execution ..status:done
	//Loop: Repeat the process until the user exits the shell ..status:done

	// Define a map of command handlers: key is command name, value is a function that takes arguments as a string
	commands := map[string]func(string){
		"exit": func(args string) { os.Exit(0) },
		// Add more commands here, e.g.:
		"echo": func(args string) { fmt.Println(args) },
		// "pwd": func(args string) { cwd, _ := os.Getwd(); fmt.Println(cwd) },
	}

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("$ ")

		command, err := reader.ReadString('\n')

		if err != nil {

			fmt.Fprintln(os.Stderr, "Error reading input:", err)

			os.Exit(1)

			log.Fatal(err)
		}

		// Trim the newline and any extra whitespace
		trimmed := strings.TrimSpace(command[:len(command)-1])

		// Split the input into command and arguments
		parts := strings.Fields(trimmed)
		if len(parts) == 0 {
			continue // Skip empty input
		}

		cmd := strings.ToLower(parts[0])
		args := strings.Join(parts[1:], " ")

		// Look up the command in the map
		if handler, ok := commands[cmd]; ok {
			handler(args) // Execute the handler with arguments
		} else {
			fmt.Println(trimmed + ": command not found")
		}
	}
}
