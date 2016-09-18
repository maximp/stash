package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/maximp/stash/client"
)

func main() {
	connection, err := client.Dial("127.0.0.1:7777")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	defer connection.Close()

	fmt.Print("Connected...\nUse 'help' command for help\n> ")

	stdin := bufio.NewScanner(os.Stdin)
	for stdin.Scan() {

		var (
			cmd    string = strings.TrimSpace(stdin.Text())
			code   int    = 0
			result string
			err    error
		)

		switch cmd {
		case "":
		case "help":
			help()
		default:
			code, result, err = connection.Cmd(cmd)
		}

		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			break
		} else if code == 200 {
			fmt.Print(result, "\n> ")
		} else if code == 0 {
			fmt.Print("> ")
		} else {
			fmt.Print(code, " ", result, "\n> ")
		}
	}

	if err := stdin.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading standard input: ", err)
	}
}

// help prints short help on commands
func help() {
	fmt.Println("  set name, [key,] value")
	fmt.Println("  get name [,key]")
	fmt.Println("  push name, value")
	fmt.Println("  pop name")
	fmt.Println("  keys [name]")
	fmt.Println("  ttl name, milliseconds")
	fmt.Println("  remove name [,key]")
	fmt.Println("  nop")
	fmt.Println("  quit")
	fmt.Println("  help")
}
