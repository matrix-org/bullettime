package types

type Error interface {
	error
	Status() int
	Code() string
}

type apiError struct {
	ErrorCode    string `json:"errcode"`
	ErrorMessage string `json:"error"`
	status       int
}

func (e apiError) Code() string {
	return e.ErrorCode
}

func (e apiError) Status() int {
	return e.status
}

func (e apiError) Error() string {
	return e.ErrorMessage
}

func UnrecognizedError(message string) Error {
	return apiError{
		ErrorCode:    "M_UNRECOGNIZED",
		ErrorMessage: message,
		status:       404,
	}
}

func NotFoundError(message string) Error {
	return apiError{
		ErrorCode:    "NOT_FOUND",
		ErrorMessage: message,
		status:       404,
	}
}

func UserInUseError(message string) Error {
	return apiError{
		ErrorCode:    "M_USER_IN_USE",
		ErrorMessage: message,
		status:       400,
	}
}

func RoomInUseError(message string) Error {
	return apiError{
		ErrorCode:    "M_ROOM_IN_USE",
		ErrorMessage: message,
		status:       400,
	}
}

func ForbiddenError(message string) Error {
	return apiError{
		ErrorCode:    "M_FORBIDDEN",
		ErrorMessage: message,
		status:       403,
	}
}

func UnknownTokenError(message string) Error {
	return apiError{
		ErrorCode:    "M_UNKNOWN_TOKEN",
		ErrorMessage: message,
		status:       403,
	}
}

func BadJsonError(message string) Error {
	return apiError{
		ErrorCode:    "M_BAD_JSON",
		ErrorMessage: message,
		status:       400,
	}
}

func BadParamError(message string) Error {
	return apiError{
		ErrorCode:    "M_BAD_PARAM",
		ErrorMessage: message,
		status:       400,
	}
}

func BadQueryError(message string) Error {
	return apiError{
		ErrorCode:    "M_BAD_QUERY",
		ErrorMessage: message,
		status:       400,
	}
}

func NotJsonError(message string) Error {
	return apiError{
		ErrorCode:    "M_NOT_JSON",
		ErrorMessage: message,
		status:       400,
	}
}

func MissingTokenError(message string) Error {
	return apiError{
		ErrorCode:    "M_MISSING_TOKEN",
		ErrorMessage: message,
		status:       403,
	}
}

func UnkownError(message string) Error {
	return apiError{
		ErrorCode:    "M_UNKNOWN",
		ErrorMessage: message,
		status:       400,
	}
}

func ServerError(message string) Error {
	return apiError{
		ErrorCode:    "M_SERVER_ERROR",
		ErrorMessage: message,
		status:       500,
	}
}

var DefaultUnrecognizedError = UnrecognizedError("unrecognized request")
var DefaultMissingTokenError = MissingTokenError("Missing access token")
var DefaultUnknownTokenError = UnknownTokenError("Unrecognised access token")
