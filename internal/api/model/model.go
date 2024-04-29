package model

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

type Response struct {
	Message string `json:"msg,omitempty"`
	Error   string `json:"error,omitempty"`
}
