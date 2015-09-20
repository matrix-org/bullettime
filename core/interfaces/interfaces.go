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

package interfaces

import "github.com/matrix-org/bullettime/core/types"

type IdMapStore interface {
	// Does nothing and returns false if the mapping already exists
	Insert(key types.Id, value types.Id) (inserted bool, err types.Error)
	// Does nothing and returns false if the mapping doesn't already exist
	Replace(key types.Id, value types.Id) (replaced bool, err types.Error)
	// Inserts or replaces as needed
	Put(key types.Id, value types.Id) types.Error
	// Does noting and returns false if the mapping doesn't exist
	Delete(key types.Id, value types.Id) (deleted bool, err types.Error)
	Lookup(key types.Id) (*types.Id, types.Error)
	ReverseLookup(value types.Id) ([]types.Id, types.Error)
}
