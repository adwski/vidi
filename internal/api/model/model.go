package model

// Prepared http api common responses.
var (
	ResponseUnauthorized = &Response{
		Error: "unauthorized",
	}
	ResponseInternalError = &Response{
		Error: "internal error",
	}
	ResponseIncorrectStatus = &Response{
		Error: "incorrect status",
	}
	ResponseIncorrectParams = &Response{
		Error: "incorrect params",
	}
	ResponseIncorrectCredentials = &Response{
		Error: "incorrect credentials",
	}
	ResponseRegistrationComplete = &Response{
		Message: "registration complete",
	}
	ResponseOK = &Response{
		Message: "ok",
	}
)

// Response is a common response struct that is used by APIs.
type Response struct {
	Message string `json:"msg,omitempty"`
	Error   string `json:"error,omitempty"`
}
