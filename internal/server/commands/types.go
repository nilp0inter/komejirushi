package commands

import "encoding/json"

type ServerCommand struct {
	command string
	payload json.RawMessage
}

type SearchPayload struct {
	term string
}
