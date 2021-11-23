package lodparse

type LodArchiveType uint32
const (
	Base      LodArchiveType = 0x01
	Expansion LodArchiveType = 0x02
)
var unknownTypes = [2]LodArchiveType{0xff, 0x1F4}

func (lat LodArchiveType) IsBaseType() bool {
	return lat == Base
}

func (lat LodArchiveType) IsExpansionType() bool {
	return lat == Expansion
}

func (lat LodArchiveType) IsUnknownType() bool {
	for _, t := range unknownTypes {
		if lat == t {
			return true
		}
	}
	return false
}