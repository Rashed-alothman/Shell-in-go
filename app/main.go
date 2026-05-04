package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
)

func main() {
	// TODO: immplement REPL loop
	//Read: Display a prompt and wait for user input ..status:done
	//Evaluate: Parse the input and execute the command ..status:done
	//Print: Display the result of the command execution ..status:done
	//Loop: Repeat the process until the user exits the shell ..status:done
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("$ ")

		command, err := reader.ReadString('\n')

		if err != nil {

			fmt.Fprintln(os.Stderr, "Error reading input:", err)

			os.Exit(1)

			log.Fatal(err)
		}

		fmt.Println(command[:len(command)-1] + ": command not found")
	}
}
