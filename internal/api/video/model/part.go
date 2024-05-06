package model

const (
	PartStatusInProgress = iota
	PartStatusInvalid
	PartStatusOK
)

type Part struct {
	Num      uint   `json:"num"`
	Size     uint64 `json:"size"`
	Status   int    `json:"status"`
	Checksum string `json:"checksum"`
}
