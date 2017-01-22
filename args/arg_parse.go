package args

import (
	"errors"
	"github.com/apflieger/tie/commands"
)

var (
	NoArgsError        = errors.New("")
	NoSuchCommandError = errors.New("")
)

func ParseArgs(args []string) (commands.Command, []string, error) {
	if len(args) < 2 {
		return nil, nil, NoArgsError
	}

	command, err := buildCommand(args[1])

	if err != nil {
		return nil, nil, err
	}

	return command, args[2:], nil
}

func buildCommand(verb string) (command commands.Command, err error) {
	switch verb {
	case "help":
		command = commands.HelpCommand
		break
	case "select":
		command = commands.SelectCommand
	}

	if command == nil {
		err = NoSuchCommandError
	}

	return
}
