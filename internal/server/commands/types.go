package commands

import "encoding/json"

type ServerCommand struct {
	Command string
	Payload json.RawMessage
}

type SearchPayload struct {
	Term string
}
