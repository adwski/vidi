package model

import (
	"encoding/json"
	"errors"
	"fmt"
)

// Video statuses.
// TODO: May be transitive statuses are redundant?
// TODO: How these statuses will map to DB enums? (or don't use enums may be?)
const (
	VideoStatusError = iota - 1
	VideoStatusCreated
	VideoStatusUploading
	VideoStatusUploaded
	VideoStatusProcessing
	VideoStatusReady
)

type Status int

var GetStatusName = func(s Status) string { return statusNames[s] }

var statusNames = map[Status]string{
	VideoStatusError:      "error",
	VideoStatusCreated:    "created",
	VideoStatusUploading:  "uploading",
	VideoStatusUploaded:   "uploaded",
	VideoStatusProcessing: "processing",
	VideoStatusReady:      "ready",
}

var statusFromName = make(map[string]Status)

func init() {
	for k, v := range statusNames {
		statusFromName[v] = k
	}
}

func GetStatusFromName(name string) (Status, error) {
	if status, ok := statusFromName[name]; ok {
		return status, nil
	}
	return 0, errors.New("incorrect status name")
}

func (s *Status) String() string {
	return statusNames[*s]
}

func (s *Status) UnmarshalJSON(b []byte) (err error) {
	var v interface{}
	if err = json.Unmarshal(b, &v); err != nil {
		return
	}
	switch value := v.(type) {
	case string:
		*s, err = GetStatusFromName(value)
	case int:
		*s = Status(value)
		if s.String() == "" {
			err = fmt.Errorf("unknown status num: %d", value)
		}
	default:
		err = errors.New("invalid status")
	}
	return
}
