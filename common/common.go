/*
 * Copyright (C) 2018 The ontology Authors
 * This file is part of The ontology library.
 *
 * The ontology is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The ontology is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with The ontology.  If not, see <http://www.gnu.org/licenses/>.
 */

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
package common

import (
	"encoding/binary"
	"fmt"
	sdk "github.com/ontio/ontology-go-sdk"
	asdk "github.com/polynetwork/poly-go-sdk"
	"github.com/polynetwork/poly/common/password"
)

func GetAliaAccountByPassword(asdk *asdk.PolySdk, path, pwdStr string) (*asdk.Account, bool) {
	wallet, err := asdk.OpenWallet(path)
	if err != nil {
		fmt.Println("open wallet error:", err)
		return nil, false
	}
	var pwd []byte
	if pwdStr != "" {
		pwd = []byte(pwdStr)
	} else {
		pwd, err = password.GetPassword()
		if err != nil {
			fmt.Println("getPassword error:", err)
			return nil, false
		}
	}
	user, err := wallet.GetDefaultAccount(pwd)
	if err != nil {
		fmt.Println("getDefaultAccount error:", err)
		return nil, false
	}
	return user, true
}

func GetSideAccountByPassword(sdk *sdk.OntologySdk, path, pwdStr string) (*sdk.Account, bool) {
	wallet, err := sdk.OpenWallet(path)
	if err != nil {
		fmt.Println("open wallet error:", err)
		return nil, false
	}

	var pwd []byte
	if pwdStr != "" {
		pwd = []byte(pwdStr)
	} else {
		pwd, err = password.GetPassword()
		if err != nil {
			fmt.Println("getPassword error:", err)
			return nil, false
		}
	}

	user, err := wallet.GetDefaultAccount(pwd)
	if err != nil {
		fmt.Println("getDefaultAccount error:", err)
		return nil, false
	}
	return user, true
}

func ConcatKey(args ...[]byte) []byte {
	temp := []byte{}
	for _, arg := range args {
		temp = append(temp, arg...)
	}
	return temp
}

func GetUint32Bytes(num uint32) []byte {
	var p [4]byte
	binary.LittleEndian.PutUint32(p[:], num)
	return p[:]
}

func GetBytesUint32(b []byte) uint32 {
	if len(b) != 4 {
		return 0
	}
	return binary.LittleEndian.Uint32(b[:])
}

func GetBytesUint64(b []byte) uint64 {
	if len(b) != 8 {
		return 0
	}
	return binary.LittleEndian.Uint64(b[:])
}

func GetUint64Bytes(num uint64) []byte {
	var p [8]byte
	binary.LittleEndian.PutUint64(p[:], num)
	return p[:]
}
