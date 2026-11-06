package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"strconv"
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

const (
    StatusSuccess      = 0
    StatusGeneralError = 1
    StatusUsageError   = 2
    StatusNotFound     = 127
    StatusNotExecutable = 126
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)

	currentUser, userErr := user.Current()

	lastExitStatus := StatusSuccess

	if userErr != nil {
		fmt.Println("Unable To Find Current User!")
		lastExitStatus = StatusGeneralError
		os.Exit(lastExitStatus)
	}


	for {
		currentDir, err := os.Getwd();

		if err != nil {
			fmt.Println("Unable to direct current working directory!")
			lastExitStatus = StatusGeneralError
			os.Exit(lastExitStatus)
		}

		terminalLine := currentUser.Username + ":" + currentDir + "$" + " "

		fmt.Print(terminalLine)

		exitStatus := strconv.Itoa(lastExitStatus)

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

		code = strings.ReplaceAll(code, "$?", exitStatus)
		code = os.ExpandEnv(code)

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
						lastExitStatus = StatusUsageError
						continue
					}
					err := os.Chdir(homeDir)
					if err != nil {
						fmt.Println("cd:", err)
						lastExitStatus = StatusUsageError
					}
					continue
				}

				dir := strings.Join(codeArgList, "")
				er := os.Chdir(dir)
				if er != nil {
					fmt.Println("cd:", er)
					lastExitStatus = StatusUsageError
				}
				lastExitStatus = StatusSuccess
				continue
			case "pwd":
				dir, er := os.Getwd() 
				if er != nil {
					fmt.Println("pwd:", er)
					lastExitStatus = StatusUsageError
					continue
				}
				fmt.Println(dir)
				lastExitStatus = StatusSuccess
				continue
			case "exit":
				os.Exit(lastExitStatus)
			case "export":
				if len(codeArgList) != 1 {
					fmt.Println("export: usage: export KEY=value")
					lastExitStatus = StatusUsageError
					continue
				}

				keyValueList := strings.SplitN(codeArgList[0], "=", 2)

				if len(keyValueList) != 2 || keyValueList[0] == "" {
					fmt.Println("export: invalid assignment")
					lastExitStatus = StatusUsageError
					continue
				}
				key := keyValueList[0]
				value := keyValueList[1]

				err := os.Setenv(key, value)
				if err != nil {
					fmt.Println("export:", err)
					lastExitStatus = StatusUsageError
					continue
				}

				lastExitStatus = StatusSuccess
				continue
			case "unset":
				if len(codeArgList) == 0 {
					fmt.Println("unset: usage: unset KEY...")
					lastExitStatus = StatusUsageError
					continue
				}

				for _, key := range codeArgList {
					err := os.Unsetenv(key)
					if err != nil {
						fmt.Println("unset:", err)	
						lastExitStatus = StatusUsageError
					}
				}

				lastExitStatus = StatusSuccess
				continue
			case "help":
				showHelp()
				lastExitStatus = StatusSuccess
				continue
		}

		err = cmd.Run()
		if err != nil {
			_, lookErr := exec.LookPath(codeCommand)
			if lookErr != nil {
				fmt.Println("gosh:", codeCommand, ": command not found")
				lastExitStatus = StatusNotFound
				continue
			}

			var exitErr *exec.ExitError
			if errors.As(err, &exitErr){
				fmt.Println("Command exited with status:", exitErr.ExitCode())
				lastExitStatus = exitErr.ExitCode()
        		continue
			}

			if errors.Is(err, os.ErrPermission) {
				fmt.Println("Permission denied:", codeCommand)
			} else {
				fmt.Println("Execution error:", err)
			}

			lastExitStatus = StatusNotExecutable
			continue
		}

		lastExitStatus = StatusSuccess
	}
}