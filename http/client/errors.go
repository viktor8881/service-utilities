package simplehttp

import (
	"fmt"
)

type ClientResponseNot200Error struct {
	ClientResponseCode int
	ClientResponseBody string
	Err                error
}

func (e *ClientResponseNot200Error) Error() string {
	return fmt.Sprintf("ClientResponseNot200Error: code: %d, message: %s, error: %s", e.ClientResponseCode, e.ClientResponseBody, e.Err.Error())
}
