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

import (
	"strconv"
	"time"
)

type Content interface{}

type Timestamp struct {
	time.Time
}

type Event interface {
	GetContent() interface{}
	GetEventType() string
	GetRoomId() *RoomId
	GetUserId() *UserId
	GetEventKey() Id
}

type IndexedEvent interface {
	Event() Event
	Index() uint64
}

type TypedContent interface {
	GetEventType() string
}

func (ts Timestamp) MarshalJSON() ([]byte, error) {
	ms := ts.UnixNano() / int64(time.Millisecond)
	return []byte(strconv.FormatInt(ms, 10)), nil
}
