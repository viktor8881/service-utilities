package client

import (
	"fmt"
)

type ClientResponseNot200Error struct {
	ClientResponseCode int
	ClientResponseBody string
	Err                error
}

func (e *ClientResponseNot200Error) Error() string {
	mess := fmt.Sprintf("ClientResponseNot200Error: code: %d, message: %s", e.ClientResponseCode, e.ClientResponseBody)
	if e.Err != nil {
		mess += "; error: " + e.Err.Error()
	}

	return mess
}
