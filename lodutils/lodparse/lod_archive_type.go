package lodparse

type LodArchiveType int32
const (
	Base      LodArchiveType = 0x01
	Expansion LodArchiveType = 0x02
	Unknown   LodArchiveType = 0xff
)

func (lat LodArchiveType) IsBaseType() bool {
	return lat == Base
}

func (lat LodArchiveType) IsExpansionType() bool {
	return lat == Expansion
}

func (lat LodArchiveType) IsUnknownType() bool {
	return lat == Unknown
}