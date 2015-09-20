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

import (
	"sync"

	"github.com/matrix-org/bullettime/core/interfaces"
	"github.com/matrix-org/bullettime/core/types"
)

type idDataCache struct {
	sync.RWMutex
	data map[types.Id]*idDataFields
}

type idDataFields struct {
	sync.RWMutex
	fields []interface{}
}

func NewIdDataCache() (interfaces.IdDataCache, error) {
	return &idDataCache{
		data: map[types.Id]*idDataFields{},
	}, nil
}

func (c *idDataCache) Put(id types.Id, fieldId int, data interface{}) {
	c.Lock()
	idData := c.data[id]
	if idData == nil {
		idData = &idDataFields{
			fields: make([]interface{}, fieldId+1),
		}
		c.data[id] = idData
	}
	idData.Lock()
	defer idData.Unlock()
	c.Unlock()
	if len(idData.fields) <= fieldId {
		oldFields := idData.fields
		idData.fields = make([]interface{}, fieldId+1)
		copy(idData.fields, oldFields)
	}
	idData.fields[fieldId] = data
}

func (c *idDataCache) LockedTransform(id types.Id, fieldId int, fun interfaces.DataTransformFunc) {
	c.Lock()
	idData := c.data[id]
	if idData == nil {
		idData = &idDataFields{
			fields: make([]interface{}, fieldId+1),
		}
		c.data[id] = idData
	}
	idData.Lock()
	defer idData.Unlock()
	c.Unlock()
	if len(idData.fields) <= fieldId {
		oldFields := idData.fields
		idData.fields = make([]interface{}, fieldId+1)
		copy(idData.fields, oldFields)
	}
	data := idData.fields[fieldId]
	data = fun(data)
	idData.fields[fieldId] = data
}

func (c *idDataCache) Lookup(id types.Id, fieldId int) interface{} {
	c.Lock()
	defer c.Unlock()
	idData := c.data[id]
	if idData == nil {
		return nil
	}
	idData.Lock()
	defer idData.Unlock()
	if len(idData.fields) <= fieldId {
		return nil
	}
	return idData.fields[fieldId]
}
