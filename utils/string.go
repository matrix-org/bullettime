// Copyright 2015 OpenMarket Ltd
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

package utils

import (
	"crypto/rand"
	"math/big"
)

const randomStringCharset = "abcdefghijklmnopqrstuvxyzABCDEFGHIJKLMNOPQRSTUVXYZ0123456789"

var randomStringCharsetLength = big.NewInt(int64(len(randomStringCharset)))

func RandomString(length int) string {
	result := make([]byte, length)
	for i := 0; i < length; i++ {
		value, err := rand.Int(rand.Reader, randomStringCharsetLength)
		if err != nil {
			panic(err)
		}
		result[i] = randomStringCharset[value.Int64()]
	}
	return string(result)
}

func StripQuotes(str string) string {
	if len(str) > 1 && str[0] == '"' && str[len(str)-1] == '"' {
		return str[1 : len(str)-1]
	}
	return str
}
