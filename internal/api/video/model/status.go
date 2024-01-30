package model

import (
	"encoding/json"
	"errors"
	"fmt"
)

// Video statuses.
// TODO: May be transitive statuses are redundant?
const (
	StatusError Status = iota - 1
	StatusCreated
	StatusUploading
	StatusUploaded
	StatusProcessing
	StatusReady
)

type Status int

var statusNames = map[Status]string{
	StatusError:      "error",
	StatusCreated:    "created",
	StatusUploading:  "uploading",
	StatusUploaded:   "uploaded",
	StatusProcessing: "processing",
	StatusReady:      "ready",
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
			err = fmt.Errorf("unknown status num: %v", value)
		}
	case float64:
		*s = Status(int(value))
		if s.String() == "" {
			err = fmt.Errorf("unknown status num: %f", value)
		}
	default:
		err = fmt.Errorf("status unmarshal error, invalid type: %T", v)
	}
	return
}
