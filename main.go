package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"strings"
)

func showHelp() {
	fmt.Println("Available Build In Commands:")
	fmt.Println()
	fmt.Println("  cd       - Change Current Directory")
	fmt.Println("  pwd      - Print CurrentWorking Directory")
	fmt.Println("  export   - Set enviornment variables")
	fmt.Println("  unset    - Remove environment variables")
	fmt.Println("  help     - Provide details of built-in commands")
	fmt.Println("  exit     - Exit The Shell")
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)

	currentUser, userErr := user.Current()

	if userErr != nil {
		fmt.Println("Unable To Find Current User!")
		os.Exit(1)
	}


	for {
		currentDir, err := os.Getwd();

		if err != nil {
			fmt.Println("Unable to direct current working directory!")
			os.Exit(1)
		}

		terminalLine := currentUser.Username + ":" +currentDir + "$" + " "

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
		if len(code) == 0 {
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

		switch codeCommand {
			case "cd":
				if len(codeArgList) == 0 {
					homeDir, er := os.UserHomeDir()
					if er != nil {
						fmt.Println("cd:", er)
					}
					err := os.Chdir(homeDir)
					if err != nil {
						fmt.Println("cd:", err)
					}
					continue
				}

				dir := strings.Join(codeArgList, "")
				er := os.Chdir(dir)
				if er != nil {
					fmt.Println("cd:", er)
				}
				continue
			case "pwd":
				dir, er := os.Getwd() 
				if er != nil {
					fmt.Println("pwd:", er)
					continue
				}
				fmt.Println(dir)
				continue
			case "exit":
				os.Exit(0)
			case "export":
				if len(codeArgList) != 1 {
					fmt.Println("export: usage: export KEY=value")
					continue
				}

				keyValueList := strings.SplitN(codeArgList[0], "=", 2)

				if len(keyValueList) != 2 || keyValueList[0] == "" {
					fmt.Println("export: invalid assignment")
					continue
				}
				key := keyValueList[0]
				value := keyValueList[1]

				os.Setenv(key, value)
				continue
			case "unset":
				if len(codeArgList) == 0 {
					fmt.Println("unset: usage: unset KEY...")
					continue
				}

				for _, key := range codeArgList {
					err := os.Unsetenv(key)
					if err != nil {
						fmt.Println("unset:", err)
					}
				}
				continue
			case "help":
				showHelp()
				continue
		}

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