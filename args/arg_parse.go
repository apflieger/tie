package args

func ParseArgs(args []string) (string, []string, error) {
	if len(args) < 2 {
		return "", nil, NoArgsError{}
	}

	return args[1], args[2:], nil
}

type NoArgsError struct{}

func (e NoArgsError) Error() string {
	return ""
}
