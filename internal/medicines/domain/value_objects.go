package domain

import "strconv"

type (
	MedicineName string
	DoseQuantity float64
	ResolvedBy   string
	DoseNotes    string
)

// PatientKind distinguishes the two kinds of family members that receive
// medication. It doubles as the species field: no further taxonomy drives any
// behaviour in this module.
type PatientKind string

const (
	PatientKindHuman  PatientKind = "human"
	PatientKindAnimal PatientKind = "animal"
)

func (k PatientKind) Validate() error {
	switch k {
	case PatientKindHuman, PatientKindAnimal:
		return nil
	default:
		return ErrInvalidPatientKind
	}
}

// DoseUnit is how a dose is measured when it is administered.
type DoseUnit string

const (
	DoseUnitDrop     DoseUnit = "drop"
	DoseUnitPill     DoseUnit = "pill"
	DoseUnitSpoonful DoseUnit = "spoonful"
)

func (u DoseUnit) Validate() error {
	switch u {
	case DoseUnitDrop, DoseUnitPill, DoseUnitSpoonful:
		return nil
	default:
		return ErrInvalidDoseUnit
	}
}

// Label returns the Spanish noun for the unit, pluralised to match quantity.
func (u DoseUnit) Label(quantity DoseQuantity) string {
	singular := quantity >= -1 && quantity <= 1

	switch u {
	case DoseUnitDrop:
		if singular {
			return "gota"
		}
		return "gotas"
	case DoseUnitPill:
		if singular {
			return "pastilla"
		}
		return "pastillas"
	case DoseUnitSpoonful:
		if singular {
			return "cucharada"
		}
		return "cucharadas"
	default:
		return string(u)
	}
}

// String renders the quantity without trailing zeros, so 15 reads as "15" and
// 0.5 as "0.5".
func (q DoseQuantity) String() string {
	return strconv.FormatFloat(float64(q), 'f', -1, 64)
}

// DoseStatus is the lifecycle of a single scheduled dose.
type DoseStatus string

const (
	DoseStatusPending      DoseStatus = "pending"
	DoseStatusAdministered DoseStatus = "administered"
	DoseStatusSkipped      DoseStatus = "skipped"
)

func (s DoseStatus) Validate() error {
	switch s {
	case DoseStatusPending, DoseStatusAdministered, DoseStatusSkipped:
		return nil
	default:
		return ErrInvalidDoseStatus
	}
}
