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
	"unicode"
)

const (
	StatusSuccess       = 0
	StatusGeneralError  = 1
	StatusUsageError    = 2
	StatusNotFound      = 127
	StatusNotExecutable = 126
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

func tokenize(input string) ([]string, error) {
	var res []string
	var word strings.Builder

	var quote byte
	tokenStarted := false

	for i := 0; i < len(input); i++ {
		c := input[i]

		if quote == 0 {
			op := ""

			if i+1 < len(input) && input[i:i+2] == "2>" && (i == 0 || input[i-1] == ' ' || input[i-1] == '\t') {
				op = "2>"
			} else if i+1 < len(input) && input[i:i+2] == ">>" {
				op = ">>"
			} else if c == '>' || c == '<' {
				op = string(c)
			}

			if op != "" {
				if tokenStarted {
					res = append(res, word.String())
					word.Reset()
					tokenStarted = false
				}

				res = append(res, op)

				if len(op) == 2 {
					i++
				}
				continue
			}
		}

		if c == '"' || c == '\'' {
			if quote == 0 {
				quote = c
				tokenStarted = true
			} else if quote == c {
				quote = 0
			} else {
				word.WriteByte(c)
			}
			continue
		}

		if unicode.IsSpace(rune(c)) && quote == 0 {
			if tokenStarted {
				res = append(res, word.String())
				word.Reset()
				tokenStarted = false
			}
			continue
		}

		word.WriteByte(c)
		tokenStarted = true
	}

	if quote != 0 {
		return nil, errors.New("unclosed quote")
	}

	if tokenStarted {
		res = append(res, word.String())
	}

	return res, nil
}

func parseRedirection(tokens []string) ([]string, map[string]string, error) {
	args := []string{}
	mp := map[string]string{}

	op := ""
	filename := ""

	for i := 0; i < len(tokens); i++ {
		c := tokens[i]

		if c == "2>" || c == ">>" || c == ">" || c == "<" {
			if i+1 >= len(tokens) {
				return args, mp, errors.New("gosh: missing redirection filename")
			}
			op = c
			filename = tokens[i+1]
			i++
			continue
		} else {
			args = append(args, c)
		}
	}

	if op == "" {
		return args, mp, nil
	}

	if filename == "" {
		return args, mp, errors.New("gosh: no filename")
	}

	mp[op] = filename

	return args, mp, nil
}

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
		currentDir, err := os.Getwd()

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

		codeList, err := tokenize(code)
		if err != nil {
			fmt.Println("gosh:", err)
			lastExitStatus = StatusGeneralError
			continue
		}

		var outputFile *os.File
		var inputFile *os.File
		var errorFile *os.File
		codeList, redirections, err := parseRedirection(codeList)

		if err != nil {
			fmt.Println(err)
			continue
		}

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

		if len(redirections) > 0 {
			var op string
			var file string

			for key := range redirections {
				op = key
				file = redirections[key]
				break
			}

			if op == ">" {
				outputFile, err = os.Create(file)
				if err != nil {
					fmt.Println("gosh:", err)
					lastExitStatus = StatusGeneralError
					continue
				}
				cmd.Stdout = outputFile
			} else if op == ">>" {
				outputFile, err = os.OpenFile(file, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
				if err != nil {
					fmt.Println("gosh:", err)
					lastExitStatus = StatusGeneralError
					continue
				}
				cmd.Stdout = outputFile
			} else if op == "<" {
				inputFile, err = os.Open(file)
				if err != nil {
					fmt.Println("gosh:", err)
					lastExitStatus = StatusGeneralError
					continue
				}
				cmd.Stdin = inputFile
			} else if op == "2>" {
				errorFile, err = os.Create(file)
				if err != nil {
					fmt.Println("gosh:", err)
					lastExitStatus = StatusGeneralError
					continue
				}
				cmd.Stderr = errorFile
			}
		}

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
		if outputFile != nil {
			outputFile.Close()
		}
		if inputFile != nil {
			inputFile.Close()
		}
		if errorFile != nil {
			errorFile.Close()
		}

		if err != nil {
			_, lookErr := exec.LookPath(codeCommand)
			if lookErr != nil {
				fmt.Println("gosh:", codeCommand, ": command not found")
				lastExitStatus = StatusNotFound
				continue
			}

			var exitErr *exec.ExitError
			if errors.As(err, &exitErr) {
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
