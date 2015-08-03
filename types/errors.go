package types

type Error struct {
	Code    string `json:"errcode"`
	Message string `json:"error"`
	status  int
}

func (e *Error) Status() int {
	return e.status
}

func (e *Error) Error() string {
	return e.Message
}

func UnrecognizedError(message string) *Error {
	return &Error{
		Code:    "M_UNRECOGNIZED",
		Message: message,
		status:  404,
	}
}

func NotFoundError(message string) *Error {
	return &Error{
		Code:    "NOT_FOUND",
		Message: message,
		status:  404,
	}
}

func UserInUseError(message string) *Error {
	return &Error{
		Code:    "M_USER_IN_USE",
		Message: message,
		status:  400,
	}
}

func ForbiddenError(message string) *Error {
	return &Error{
		Code:    "M_FORBIDDEN",
		Message: message,
		status:  403,
	}
}

func UnknownTokenError(message string) *Error {
	return &Error{
		Code:    "M_UNKNOWN_TOKEN",
		Message: message,
		status:  403,
	}
}

func BadJsonError(message string) *Error {
	return &Error{
		Code:    "M_BAD_JSON",
		Message: message,
		status:  400,
	}
}

func MissingTokenError(message string) *Error {
	return &Error{
		Code:    "M_MISSING_TOKEN",
		Message: message,
		status:  403,
	}
}

func UnkownError(message string) *Error {
	return &Error{
		Code:    "M_UNKNOWN",
		Message: message,
		status:  400,
	}
}

func ServerError(message string) *Error {
	return &Error{
		Code:    "M_SERVER_ERROR",
		Message: message,
		status:  500,
	}
}

var DefaultUnrecognizedError = UnrecognizedError("unrecognized request")
var DefaultMissingTokenError = MissingTokenError("Missing access token")
var DefaultUnknownTokenError = UnknownTokenError("Unrecognised access token")
