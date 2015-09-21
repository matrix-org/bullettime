// Copyright 2015  Ericsson AB
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package types

import ct "github.com/matrix-org/bullettime/core/types"

type Error interface {
	ct.Error
	Status() int
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

func InternalError(err ct.Error) Error {
	if err == nil {
		return nil
	}
	return apiError{
		ErrorCode:    err.Code(),
		ErrorMessage: err.Error(),
		status:       500,
	}
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
