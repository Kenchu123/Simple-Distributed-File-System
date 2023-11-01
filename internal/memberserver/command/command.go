package command

// State is a string that represents the state of a member
type Command string

// Constants for the different states of a member
const (
	// member
	JOIN  Command = "JOIN"
	LEAVE Command = "LEAVE"
	FAIL  Command = "FAIL"
	LIST  Command = "LIST"
	ID    Command = "ID"
	// config
	SUSPICION Command = "SUSPICION"
	DROPRATE  Command = "DROPRATE"
	VERBOSE   Command = "VERBOSE"
)
