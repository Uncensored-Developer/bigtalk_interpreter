package main

import (
	"BigTalk_Interpreter/repl"
	"fmt"
	"os"
	"os/user"
)

func main() {
	osUser, err := user.Current()
	if err != nil {
		panic(err)
	}

	fmt.Printf("Hello %s, this is the BigTalk programming language.", osUser.Username)
	fmt.Printf("Type in commands\n")
	repl.Start(os.Stdin, os.Stdout)
}
