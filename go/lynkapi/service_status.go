// Copyright 2024 Eryx <evorui at gmail dot com>, All rights reserved.
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

package lynkapi

import (
	"errors"
)

func NewClientError(msg string) error {
	return NewError(StatusCode_BadRequest, msg)
}

func NewBadRequestError(msg string) error {
	return NewError(StatusCode_BadRequest, msg)
}

func NewUnAuthError(msg string) error {
	return NewError(StatusCode_UnAuth, msg)
}

func NewNotFoundError(msg string) error {
	return NewError(StatusCode_NotFound, msg)
}

func NewTimeoutError(msg string) error {
	return NewError(StatusCode_Timeout, msg)
}

func NewConflictError(msg string) error {
	return NewError(StatusCode_Conflict, msg)
}

func NewRateLimitError(msg string) error {
	return NewError(StatusCode_RateLimit, msg)
}

func NewInternalServerError(msg string) error {
	return NewError(StatusCode_InternalServerError, msg)
}

func NewNotImplementedError(msg string) error {
	return NewError(StatusCode_NotImplemented, msg)
}

func NewServerUnavailableError(msg string) error {
	return NewError(StatusCode_ServiceUnavailable, msg)
}

func NewError(code, msg string) error {
	return errors.New("#" + code + " " + msg)
}

func NewResponseError(code, msg string) *Response {
	return &Response{
		Kind: "Error",
		Status: &ServiceStatus{
			Code:    code,
			Message: msg,
		},
	}
}

func (it *ServiceStatus) OK() bool {
	return it.Code == StatusCode_OK
}

func (it *ServiceStatus) Err() error {
	if it != nil && it.Code != StatusCode_OK {
		return errors.New("#" + it.Code + " " + it.Message)
	}
	return nil
}

func NewServiceStatus(code, msg string) *ServiceStatus {
	return &ServiceStatus{
		Code:    code,
		Message: msg,
	}
}

func NewServiceStatusOK() *ServiceStatus {
	return &ServiceStatus{
		Code: StatusCode_OK,
	}
}

func NewServiceStatusClientError(msg string) *ServiceStatus {
	return &ServiceStatus{
		Code:    StatusCode_BadRequest,
		Message: msg,
	}
}

func NewServiceStatusServerError(msg string) *ServiceStatus {
	return &ServiceStatus{
		Code:    StatusCode_InternalServerError,
		Message: msg,
	}
}

func ParseError(err error) *ServiceStatus {
	if err == nil {
		return &ServiceStatus{
			Code: StatusCode_OK,
		}
	}
	if len(err.Error()) >= 6 &&
		err.Error()[0] == '#' &&
		err.Error()[5] == ' ' {
		return &ServiceStatus{
			Code:    err.Error()[1:5],
			Message: err.Error()[6:],
		}
	}
	return &ServiceStatus{
		Code:    StatusCode_InternalServerError,
		Message: "unknown error",
	}
}
