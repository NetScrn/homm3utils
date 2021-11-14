package lodparse

type LodArchiveType int32
const (
	Base      LodArchiveType = 0x01
	Expansion LodArchiveType = 0x02
	Unknown   LodArchiveType = 0xff
)

func (lat LodArchiveType) IsBaseType() bool {
	if lat == Base {
		return true
	}
	return false
}

func (lat LodArchiveType) IsExpansionType() bool {
	if lat == Expansion {
		return true
	}
	return false
}

func (lat LodArchiveType) IsUnknownType() bool {
	if lat == Unknown {
		return true
	}
	return false
}