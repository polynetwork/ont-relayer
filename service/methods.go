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
package service

import (
	"encoding/hex"
	"fmt"
	common2 "github.com/ontio/ontology/common"
	"os"
	"strings"
	"time"

	"github.com/ontio/ontology/smartcontract/service/native/cross_chain/cross_chain_manager"
	"github.com/ontio/ontology/smartcontract/service/native/cross_chain/header_sync"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
	"github.com/polynetwork/ont-relayer/common"
	"github.com/polynetwork/ont-relayer/db"
	"github.com/polynetwork/ont-relayer/log"
	acommon "github.com/polynetwork/poly/common"
	hscommon "github.com/polynetwork/poly/native/service/header_sync/common"
	autils "github.com/polynetwork/poly/native/service/utils"
)

var codeVersion = byte(0)

func (this *SyncService) GetSideChainID() uint64 {
	return this.config.SideChainID
}

func (this *SyncService) GetGasPrice() uint64 {
	return this.config.GasPrice
}

func (this *SyncService) GetGasLimit() uint64 {
	return this.config.GasLimit
}

func (this *SyncService) GetCurrentSideChainSyncHeight(aliaChainID uint64) (uint32, error) {
	contractAddress := utils.HeaderSyncContractAddress
	aliaChainIDBytes := common.GetUint64Bytes(aliaChainID)
	key := common.ConcatKey([]byte(header_sync.CURRENT_HEIGHT), aliaChainIDBytes)
	value, err := this.sideSdk.ClientMgr.GetStorage(contractAddress.ToHexString(), key)
	if err != nil {
		return 0, fmt.Errorf("getStorage error: %s", err)
	}
	height, err := utils.GetBytesUint32(value)
	if err != nil {
		return 0, fmt.Errorf("GetBytesUint32, get height error: %s", err)
	}
	if dbh := this.db.GetPolyHeight(); dbh > height {
		height = dbh
	}

	return height, nil
}

func (this *SyncService) GetCurrentAliaChainSyncHeight(sideChainID uint64) (uint32, error) {
	contractAddress := autils.HeaderSyncContractAddress
	sideChainIDBytes := common.GetUint64Bytes(sideChainID)

	key := common.ConcatKey([]byte(hscommon.CURRENT_MSG_HEIGHT), sideChainIDBytes)
	value, err := this.aliaSdk.ClientMgr.GetStorage(contractAddress.ToHexString(), key)
	if err != nil {
		return 0, fmt.Errorf("getStorage error: %s", err)
	}
	height := autils.GetBytesUint32(value)

	if height == 0 {
		key = common.ConcatKey([]byte(hscommon.CURRENT_HEADER_HEIGHT), sideChainIDBytes)
		value, err := this.aliaSdk.ClientMgr.GetStorage(contractAddress.ToHexString(), key)
		if err != nil {
			return 0, fmt.Errorf("getStorage error: %s", err)
		}
		height = autils.GetBytesUint32(value)
	}
	dbh := this.db.GetOntHeight()
	if dbh > height {
		height = dbh
	}

	return height, nil
}

func (this *SyncService) syncHeaderToAlia(height uint32) error {
	chainIDBytes := common.GetUint64Bytes(this.GetSideChainID())
	heightBytes := common.GetUint32Bytes(height)
	v, err := this.aliaSdk.GetStorage(autils.HeaderSyncContractAddress.ToHexString(),
		common.ConcatKey([]byte(hscommon.HEADER_INDEX), chainIDBytes, heightBytes))
	if len(v) != 0 {
		return nil
	}
	block, err := this.sideSdk.GetBlockByHeight(height)
	if err != nil {
		return fmt.Errorf("[syncHeaderToAlia] this.sideSdk.GetBlockByHeight error: %s", err)
	}
	txHash, err := this.aliaSdk.Native.Hs.SyncBlockHeader(this.GetSideChainID(), this.aliaAccount.Address, [][]byte{block.Header.ToArray()},
		this.aliaAccount)
	if err != nil {
		return fmt.Errorf("[syncHeaderToAlia] invokeNativeContract error: %s", err)
	}
	log.Infof("[syncHeaderToAlia] syncHeaderToAlia txHash is :", txHash.ToHexString())
	this.waitForAliaBlock()
	return nil
}

func (this *SyncService) syncCrossChainMsgToAlia(height uint32) error {
	chainIDBytes := common.GetUint64Bytes(this.GetSideChainID())
	heightBytes := common.GetUint32Bytes(height)
	v, err := this.aliaSdk.GetStorage(autils.HeaderSyncContractAddress.ToHexString(),
		common.ConcatKey([]byte(hscommon.CROSS_CHAIN_MSG), chainIDBytes, heightBytes))
	if len(v) != 0 {
		return nil
	}
	crossChainMsg, err := this.sideSdk.GetCrossChainMsg(height)
	if err != nil {
		return fmt.Errorf("[syncCrossChainMsgToAlia] this.sideSdk.GetCrossChainMsg error: %s", err)
	}
	params, err := hex.DecodeString(crossChainMsg)
	if err != nil {
		return fmt.Errorf("[syncCrossChainMsgToAlia] hex.DecodeString error: %s", err)
	}
	txHash, err := this.aliaSdk.Native.Hs.SyncCrossChainMsg(this.GetSideChainID(), this.aliaAccount.Address,
		[][]byte{params}, this.aliaAccount)
	if err != nil {
		return fmt.Errorf("[syncCrossChainMsgToAlia] invokeNativeContract error: %s", err)
	}
	log.Infof("[syncCrossChainMsgToAlia] syncHeaderToAlia txHash is :", txHash.ToHexString())
	this.waitForAliaBlock()
	return nil
}

func (this *SyncService) syncProofToAlia(key string, height uint32) (acommon.Uint256, error) {
	chainIDBytes := common.GetUint64Bytes(this.GetSideChainID())
	heightBytes := common.GetUint32Bytes(height)
	params := []byte{}
	v, err := this.aliaSdk.GetStorage(autils.HeaderSyncContractAddress.ToHexString(),
		common.ConcatKey([]byte(hscommon.CROSS_CHAIN_MSG), chainIDBytes, heightBytes))
	if len(v) == 0 {
		crossChainMsg, err := this.sideSdk.GetCrossChainMsg(height)
		if err != nil {
			return acommon.UINT256_EMPTY, fmt.Errorf("[syncProofToAlia] this.sideSdk.GetCrossChainMsg error: %s", err)
		}
		params, err = hex.DecodeString(crossChainMsg)
		if err != nil {
			return acommon.UINT256_EMPTY, fmt.Errorf("[syncProofToAlia] hex.DecodeString error: %s", err)
		}
	}

	k, err := hex.DecodeString(key)
	if err != nil {
		return acommon.UINT256_EMPTY, fmt.Errorf("[syncProofToAlia] hex.DecodeString error: %s", err)
	}
	proof, err := this.sideSdk.GetCrossStatesProof(height, k)
	if err != nil {
		return acommon.UINT256_EMPTY, fmt.Errorf("[syncProofToAlia] this.sideSdk.GetCrossStatesProof error: %s", err)
	}
	auditPath, err := hex.DecodeString(proof.AuditPath)
	if err != nil {
		return acommon.UINT256_EMPTY, fmt.Errorf("[syncProofToAlia] hex.DecodeString error: %s", err)
	}

	retry := &db.Retry{
		Height: height,
		Key:    key,
	}
	sink := acommon.NewZeroCopySink(nil)
	retry.Serialization(sink)

	txHash, err := this.aliaSdk.Native.Ccm.ImportOuterTransfer(this.GetSideChainID(), nil, height, auditPath,
		this.aliaAccount.Address[:], params, this.aliaAccount)
	if err != nil {
		if strings.Contains(err.Error(), "chooseUtxos, current utxo is not enough") {
			log.Infof("[syncProofToAlia] invokeNativeContract error: %s", err)

			err = this.db.PutRetry(sink.Bytes())
			if err != nil {
				return acommon.UINT256_EMPTY, fmt.Errorf("[syncProofToAlia] this.db.PutRetry error: %s", err)
			}
			log.Infof("[syncProofToAlia] put tx into retry db, height %d, key %s", height, key)
			return acommon.UINT256_EMPTY, nil
		} else {
			return acommon.UINT256_EMPTY, err
		}
	}

	err = this.db.PutCheck(txHash.ToHexString(), sink.Bytes())
	if err != nil {
		return acommon.UINT256_EMPTY, fmt.Errorf("[syncProofToAlia] this.db.PutCheck error: %s", err)
	}

	return txHash, nil
}

func (this *SyncService) retrySyncProofToAlia(v []byte) error {
	retry := new(db.Retry)
	err := retry.Deserialization(acommon.NewZeroCopySource(v))
	if err != nil {
		return fmt.Errorf("[retrySyncProofToAlia] retry.Deserialization error: %s", err)
	}
	k, err := hex.DecodeString(retry.Key)
	if err != nil {
		return fmt.Errorf("[retrySyncProofToAlia] hex.DecodeString error: %s", err)
	}
	proof, err := this.sideSdk.GetCrossStatesProof(retry.Height, k)
	if err != nil {
		return fmt.Errorf("[retrySyncProofToAlia] this.sideSdk.GetCrossStatesProof error: %s", err)
	}
	auditPath, err := hex.DecodeString(proof.AuditPath)
	if err != nil {
		return fmt.Errorf("[retrySyncProofToAlia] hex.DecodeString error: %s", err)
	}
	chainIDBytes := common.GetUint64Bytes(this.GetSideChainID())
	heightBytes := common.GetUint32Bytes(retry.Height)
	params := []byte{}
	s, err := this.aliaSdk.GetStorage(autils.HeaderSyncContractAddress.ToHexString(),
		common.ConcatKey([]byte(hscommon.CROSS_CHAIN_MSG), chainIDBytes, heightBytes))
	if len(s) == 0 {
		crossChainMsg, err := this.sideSdk.GetCrossChainMsg(retry.Height)
		if err != nil {
			return fmt.Errorf("[retrySyncProofToAlia] this.sideSdk.GetCrossChainMsg error: %s", err)
		}
		params, err = hex.DecodeString(crossChainMsg)
		if err != nil {
			return fmt.Errorf("[retrySyncProofToAlia] hex.DecodeString error: %s", err)
		}
	}

	txHash, err := this.aliaSdk.Native.Ccm.ImportOuterTransfer(this.GetSideChainID(),
		nil, retry.Height, auditPath, this.aliaAccount.Address[:], params, this.aliaAccount)
	if err != nil {
		if strings.Contains(err.Error(), "chooseUtxos, current utxo is not enough") {
			log.Infof("[retrySyncProofToAlia] invokeNativeContract error: %s", err)
			return nil
		} else {
			if err := this.db.DeleteRetry(v); err != nil {
				return fmt.Errorf("[retrySyncProofToAlia] this.db.DeleteRetry error: %s", err)
			}
			return fmt.Errorf("[retrySyncProofToAlia] invokeNativeContract error: %s", err)
		}
	}

	err = this.db.PutCheck(txHash.ToHexString(), v)
	if err != nil {
		return fmt.Errorf("[retrySyncProofToAlia] this.db.PutCheck error: %s", err)
	}
	err = this.db.DeleteRetry(v)
	if err != nil {
		return fmt.Errorf("[retrySyncProofToAlia] this.db.DeleteRetry error: %s", err)
	}

	log.Infof("[retrySyncProofToAlia] syncProofToAlia txHash is :", txHash.ToHexString())
	return nil
}

func (this *SyncService) syncHeaderToSide(height uint32) error {
	chainIDBytes := common.GetUint64Bytes(this.aliaSdk.ChainId)
	heightBytes := common.GetUint32Bytes(height)
	v, err := this.sideSdk.GetStorage(utils.HeaderSyncContractAddress.ToHexString(),
		common.ConcatKey([]byte(header_sync.HEADER_INDEX), chainIDBytes, heightBytes))
	if len(v) != 0 {
		return nil
	}
	contractAddress := utils.HeaderSyncContractAddress
	method := header_sync.SYNC_BLOCK_HEADER
	blockHeader, err := this.aliaSdk.GetHeaderByHeight(height)
	if err != nil {
		log.Errorf("[syncHeaderToSide] this.mainSdk.GetHeaderByHeight error:%s", err)
	}
	param := &header_sync.SyncBlockHeaderParam{
		Headers: [][]byte{blockHeader.ToArray()},
	}
	txHash, err := this.sideSdk.Native.InvokeNativeContract(this.GetGasPrice(), this.GetGasLimit(), this.sideAccount,
		this.sideAccount, codeVersion, contractAddress, method, []interface{}{param})
	if err != nil {
		return fmt.Errorf("[syncHeaderToSide] invokeNativeContract error: %s", err)
	}
	log.Infof("[syncHeaderToSide] syncHeaderToSide txHash is :", txHash.ToHexString())
	this.waitForSideBlock()
	return nil
}

func (this *SyncService) syncProofToSide(key string, height uint32) (common2.Uint256, error) {
	chainIDBytes := common.GetUint64Bytes(this.aliaSdk.ChainId)
	heightBytes := common.GetUint32Bytes(height + 1)
	proof, err := this.aliaSdk.ClientMgr.GetCrossStatesProof(height, key)
	if err != nil {
		return common2.UINT256_EMPTY, fmt.Errorf("[syncProofToSide] this.sideSdk.GetMptProof error: %s", err)
	}
	param := &cross_chain_manager.ProcessCrossChainTxParam{
		Address:     this.sideAccount.Address,
		FromChainID: this.aliaSdk.ChainId,
		Height:      height + 1,
		Proof:       proof.AuditPath,
	}
	v, err := this.sideSdk.GetStorage(utils.HeaderSyncContractAddress.ToHexString(),
		common.ConcatKey([]byte(header_sync.HEADER_INDEX), chainIDBytes, heightBytes))
	if len(v) == 0 {
		blockHeader, err := this.aliaSdk.GetHeaderByHeight(height + 1)
		if err != nil {
			log.Errorf("[syncHeaderToSide] this.mainSdk.GetHeaderByHeight error:%s", err)
		}
		param.Header = blockHeader.ToArray()
	}

	contractAddress := utils.CrossChainContractAddress
	method := cross_chain_manager.PROCESS_CROSS_CHAIN_TX
	txHash, err := this.sideSdk.Native.InvokeNativeContract(this.GetGasPrice(), this.GetGasLimit(),
		this.sideAccount, this.sideAccount, codeVersion, contractAddress, method, []interface{}{param})
	if err != nil {
		return common2.UINT256_EMPTY, err
	}
	return txHash, nil
}

func (this *SyncService) checkDoneTx() error {
	checkMap, err := this.db.GetAllCheck()
	if err != nil {
		return fmt.Errorf("[checkDoneTx] this.db.GetAllCheck error: %s", err)
	}
	for k, v := range checkMap {
		event, err := this.aliaSdk.GetSmartContractEvent(k)
		if err != nil {
			return fmt.Errorf("[checkDoneTx] this.aliaSdk.GetSmartContractEvent error: %s", err)
		}
		if event == nil {
			log.Infof("[checkDoneTx] can not find event of hash %s", k)
			continue
		}
		if event.State != 1 {
			log.Infof("[checkDoneTx] state of tx %s is not success", k)
			err := this.db.PutRetry(v)
			if err != nil {
				log.Errorf("[checkDoneTx] this.db.PutRetry error:%s", err)
			}
		}
		err = this.db.DeleteCheck(k)
		if err != nil {
			log.Errorf("[checkDoneTx] this.db.DeleteCheck error:%s", err)
		}
	}

	return nil
}

func (this *SyncService) retryTx() error {
	retryList, err := this.db.GetAllRetry()
	if err != nil {
		return fmt.Errorf("[retryTx] this.db.GetAllRetry error: %s", err)
	}
	for _, v := range retryList {
		err = this.retrySyncProofToAlia(v)
		if err != nil {
			log.Errorf("[retryTx] this.retrySyncProofToAlia error:%s", err)
		}
		time.Sleep(time.Duration(this.config.RetryInterval) * time.Second)
	}

	return nil
}

func (this *SyncService) waitForAliaBlock() {
	_, err := this.aliaSdk.WaitForGenerateBlock(90*time.Second, 3)
	if err != nil {
		log.Errorf("waitForAliaBlock error:%s", err)
	}
}

func (this *SyncService) waitForSideBlock() {
	_, err := this.sideSdk.WaitForGenerateBlock(90*time.Second, 3)
	if err != nil {
		log.Errorf("waitForSideBlock error:%s", err)
	}
}

func checkIfExist(dir string) bool {
	_, err := os.Stat(dir)
	if err != nil && !os.IsExist(err) {
		return false
	}
	return true
}
