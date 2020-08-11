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
	"fmt"
	"github.com/polynetwork/poly/common"
)

type Retry struct {
	Height uint32
	Key    string
}

func (this *Retry) Serialization(sink *common.ZeroCopySink) {
	sink.WriteUint32(this.Height)
	sink.WriteString(this.Key)
}

func (this *Retry) Deserialization(source *common.ZeroCopySource) error {
	height, eof := source.NextUint32()
	if eof {
		return fmt.Errorf("Waiting deserialize height error")
	}
	key, eof := source.NextString()
	if eof {
		return fmt.Errorf("Waiting deserialize key error")
	}

	this.Height = height
	this.Key = key
	return nil
}
