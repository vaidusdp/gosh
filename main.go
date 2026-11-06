package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)

	for {
		currentDir, err := os.Getwd();

		if err != nil {
			fmt.Println("Unable to direct current working directory!")
			os.Exit(1)
		}

		terminalLine := currentDir + "$" + " "

		fmt.Print(terminalLine)

		check := scanner.Scan()
		
		err = scanner.Err()
		if !check {
			if err != nil {
				fmt.Println("Error reading input:", err)
			}
			return
		}

		code := scanner.Text()
		if code == "" {
			continue
		}

		codeList := strings.Fields(code)
		if len(codeList) == 0 {
			fmt.Println("Invalid Command")
			continue
		}
		codeCommand := codeList[0]

		codeArgList := codeList[1:]

		cmd := exec.Command(codeCommand, codeArgList...)

		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		err = cmd.Run()
		if err != nil {
			_, lookErr := exec.LookPath(codeCommand)
			if lookErr != nil {
				fmt.Println("gosh:", codeCommand, ": command not found")
				continue
			}

			var exitErr *exec.ExitError
			if errors.As(err, &exitErr){
				fmt.Println("Command exited with status:", exitErr.ExitCode())
        		continue
			}

			if errors.Is(err, os.ErrPermission) {
				fmt.Println("Permission denied:", codeCommand)
			} else {
				fmt.Println("Execution error:", err)
			}

			continue
		}
	}
}