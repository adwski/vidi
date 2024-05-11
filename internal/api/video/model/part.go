package model

const (
	PartStatusInProgress = iota
	PartStatusInvalid
	PartStatusOK
)

type Part struct {
	Checksum string `json:"checksum"`
	Num      uint   `json:"num"`
	Size     uint64 `json:"size"`
	Status   int    `json:"status"`
}
