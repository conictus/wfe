package wfe

import "encoding/gob"

func init() {
	gob.Register(Error{})
}

type Error struct {
	E string
}

func (e Error) Error() string {
	return e.E
}
