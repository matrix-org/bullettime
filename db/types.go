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

package db

import "github.com/matrix-org/bullettime/types"

type State interface {
	Key() string
	Value() []byte
}

type StateStore interface {
	CreateBucket(types.Id) (bool, types.Error)
	BucketExists(types.Id) (bool, types.Error)
	SetState(id types.Id, key string, value []byte) (oldValue []byte, err types.Error)
	State(id types.Id, key string) (value []byte, err types.Error)
	States(id types.Id) ([]State, types.Error)
}

/*

presence:

Id: <userId>
Key: presence|status_msg
Value:  "online"


profile:

Id: <userId>
Key: name|avatar_url
Value:  "arne"


typing:

Id: <roomId>
Key: <userId>
Value:  "1"


room-state:

Id: <roomId>
Key: <eventType>:<stateKey>
Value:  "{\"herp\":\"derp\"}"


*/
