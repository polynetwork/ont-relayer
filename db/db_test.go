/*
* Copyright (C) 2020 The poly network Authors
* This file is part of The poly network library.
*
* The poly network is free software: you can redistribute it and/or modify
* it under the terms of the GNU Lesser General Public License as published by
* the Free Software Foundation, either version 3 of the License, or
* (at your option) any later version.
*
* The poly network is distributed in the hope that it will be useful,
* but WITHOUT ANY WARRANTY; without even the implied warranty of
* MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
* GNU Lesser General Public License for more details.
* You should have received a copy of the GNU Lesser General Public License
* along with The poly network . If not, see <http://www.gnu.org/licenses/>.
*/
package db

import (
	"encoding/hex"
	"fmt"
	"testing"

	acommon "github.com/polynetwork/poly/common"
	"github.com/stretchr/testify/assert"
)

func TestRetryDB(t *testing.T) {
	db, err := NewBoltDB("../testdb")
	assert.NoError(t, err)
	txHash, err := hex.DecodeString("253488b641eb25509bbd6bf7a744d130d2e7be24016144ae3a7049a9d2760cf0")
	assert.NoError(t, err)
	for i := 0; i < 10; i++ {
		retry := &Retry{
			TxHash: txHash,
			Height: uint32(i),
			Key:    "0000000000000000000000000000000000000009726571756573740000000000000000253488b641eb25509bbd6bf7a744d130d2e7be24016144ae3a7049a9d2760cf0",
		}
		sink1 := acommon.NewZeroCopySink(nil)
		retry.Serialization(sink1)
		err = db.PutRetry(sink1.Bytes())
		assert.NoError(t, err)
	}

	retryList, err := db.GetAllRetry()
	assert.NoError(t, err)
	for _, v := range retryList {
		fmt.Printf("####: %x \n", v)
		err := db.PutCheck(hex.EncodeToString(v), v)
		assert.NoError(t, err)
	}
}
