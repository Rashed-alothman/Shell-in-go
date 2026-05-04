package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
)

func main() {

	fmt.Print("$ ")
	// TODO: immplement a support for printing error message when the user input is invalid
	//read the user input
	command, err := bufio.NewReader(os.Stdin).ReadString('\n')
	//check if there is an error reading the user input
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error reading input:", err)
		os.Exit(1)
		log.Fatal(err)
	}
	//check if the user input is valid
	fmt.Println(command[:len(command)-1] + ": command not found")

}
