package domain

type ID string

func (vo ID) String() string {
	return string(vo)
}

type Name string
type Index uint8
