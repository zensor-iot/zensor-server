package domain

type ID string
type Version int

func (vo ID) String() string {
	return string(vo)
}

type Name string
type Index uint8
