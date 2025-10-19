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
	"regexp"
)

// Service Status Codes
const (
	StatusCode_OK = "2000"

	// Client error responses
	StatusCode_BadRequest = "4000"
	StatusCode_UnAuth     = "4010"
	StatusCode_NotFound   = "4040"
	StatusCode_Timeout    = "4080"
	StatusCode_Conflict   = "4090"
	StatusCode_RateLimit  = "4290"

	// Server error responses
	StatusCode_InternalServerError = "5000"
	StatusCode_NotImplemented      = "5010"
	StatusCode_ServiceUnavailable  = "5030"

	StatusCode_Unknown = "0000"
)

type DataMerge_Type int

const (
	DataMerge_UnSpec DataMerge_Type = iota
	DataMerge_Create
	DataMerge_Update
	DataMerge_Delete
)

const RequestSpecNameInContext = "lynkdb.lynkapi.Context.RequestSpec"

var (
	// name identifier
	NameIdentifier = regexp.MustCompile("^[a-zA-Z][a-zA-Z0-9_]{0,63}$")

	// length range [2, 16]
	NamespaceIdentifier = regexp.MustCompile("^[a-z][a-z0-9]{1,15}$")

	//
	ObjectIdentifier = regexp.MustCompile("^[0-9a-f]{4,32}$")
)
