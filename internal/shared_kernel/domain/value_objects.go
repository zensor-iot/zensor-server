package domain

type (
	ID      string
	Version int
)

func (vo ID) String() string {
	return string(vo)
}

type (
	Name        string
	Index       uint8
	DisplayName string
	Description string
)
