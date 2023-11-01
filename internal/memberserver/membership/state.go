package membership

// State is a string that represents the state of a member
type State string

// Constants for the different states of a member
const (
	ALIVE     State = "ALIVE"
	FAILED    State = "FAILED"
	LEFT      State = "LEFT"
	SUSPECTED State = "SUSPECTED"
)
