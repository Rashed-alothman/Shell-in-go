# Shell in Go

A Unix shell built from scratch in Go. Started as part of the [CodeCrafters "Build Your Own Shell"](https://app.codecrafters.io/courses/shell/overview) challenge and extended with additional builtins, proper error handling, and a clean command architecture.

## Features

- Interactive REPL loop with a `$ ` prompt
- Extensible command system — each command is a self-describing struct with a type, description, and handler
- PATH resolution for external executables
- Proper error reporting to stderr

## Built-in Commands

| Command | Description |
|---------|-------------|
| `echo <text>` | Print text to the screen |
| `pwd` | Print the current working directory |
| `cd [dir]` | Change directory. No argument goes to the home directory |
| `ls [dir]` | List directory contents. Defaults to current directory |
| `grep <pattern> <file>` | Search for a regex pattern in a file |
| `type <command>` | Describe how a command would be interpreted (builtin or PATH) |
| `clear` | Clear the terminal screen |
| `help` | List all available commands with descriptions |
| `exit` | Exit the shell |

## Getting Started

### Requirements

- Go 1.24 or higher

### Build and Run

```sh
git clone https://github.com/Rashed-alothman/Shell-in-go.git
cd Shell-in-go
go run ./app/main.go
```

Or build a binary:

```sh
go build -o shell ./app/main.go
./shell
```

## Usage

```
$ echo hello world
hello world

$ pwd
/home/user/projects

$ cd /tmp
$ pwd
/tmp

$ ls
file1.txt  file2.txt

$ grep error log.txt
error: connection refused
error: timeout

$ type echo
echo is a shell builtin

$ type git
git is /usr/bin/git

$ help
Available commands:
  cd         Change the current directory
  clear      Clear the terminal screen
  echo       Print text to the screen
  grep       Search for a pattern in a file
  help       Show all available commands
  ls         List directory contents
  pwd        Print the current working directory
  type       Describe how a command would be interpreted
  exit       Exit the shell
```

## Project Structure

```
app/
  main.go    Entry point and all builtin command definitions
go.mod
```

## How Commands Are Added

Each command is registered as an entry in a `map[string]Command`. The `Command` struct carries the command's type, a human-readable description, and its handler function. Adding a new command means adding one block — no other code needs to change.

```go
"mycommand": {
    Type:        "shell builtin",
    Description: "Does something useful",
    Handler: func(args string) {
        // implementation
    },
},
```

## License

MIT
