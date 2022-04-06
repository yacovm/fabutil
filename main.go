/*
Copyright IBM Corp. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"crypto/sha256"
	"encoding/asn1"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"math/big"
	"os"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-protos-go/common"
	"gopkg.in/alecthomas/kingpin.v2"
)

type BlockMetadataIndex int32

const (
	BlockMetadataIndex_SIGNATURES          BlockMetadataIndex = 0
	BlockMetadataIndex_LAST_CONFIG         BlockMetadataIndex = 1
	BlockMetadataIndex_TRANSACTIONS_FILTER BlockMetadataIndex = 2
	BlockMetadataIndex_ORDERER             BlockMetadataIndex = 3
)

type TxValidationFlags []uint8

var TxValidationCode_name = map[int32]string{
	0:   "VALID",
	1:   "NIL_ENVELOPE",
	2:   "BAD_PAYLOAD",
	3:   "BAD_COMMON_HEADER",
	4:   "BAD_CREATOR_SIGNATURE",
	5:   "INVALID_ENDORSER_TRANSACTION",
	6:   "INVALID_CONFIG_TRANSACTION",
	7:   "UNSUPPORTED_TX_PAYLOAD",
	8:   "BAD_PROPOSAL_TXID",
	9:   "DUPLICATE_TXID",
	10:  "ENDORSEMENT_POLICY_FAILURE",
	11:  "MVCC_READ_CONFLICT",
	12:  "PHANTOM_READ_CONFLICT",
	13:  "UNKNOWN_TX_TYPE",
	14:  "TARGET_CHAIN_NOT_FOUND",
	15:  "MARSHAL_TX_ERROR",
	16:  "NIL_TXACTION",
	17:  "EXPIRED_CHAINCODE",
	18:  "CHAINCODE_VERSION_CONFLICT",
	19:  "BAD_HEADER_EXTENSION",
	20:  "BAD_CHANNEL_HEADER",
	21:  "BAD_RESPONSE_PAYLOAD",
	22:  "BAD_RWSET",
	23:  "ILLEGAL_WRITESET",
	24:  "INVALID_WRITESET",
	25:  "INVALID_CHAINCODE",
	254: "NOT_VALIDATED",
	255: "INVALID_OTHER_REASON",
}

func (obj TxValidationFlags) Flag(txIndex int) string {
	return TxValidationCode_name[int32(obj[txIndex])]
}

var (
	f = kingpin.Flag("filename", "filename").Short('f').ExistingFile()
)

func main() {
	kingpin.Parse()
	if f == nil || *f == "" {
		fmt.Fprint(os.Stderr, "Must specify filename")
		return
	}
	block := &common.Block{}
	file, err := ioutil.ReadFile(*f)
	if err != nil {
		fmt.Fprint(os.Stderr, "Error opening file", *f, ":", err)
		return
	}

	if err := proto.Unmarshal(file, block); err != nil {
		fmt.Fprint(os.Stderr, "File", *f, "isn't a block file")
		return
	}

	currHash := hex.EncodeToString(BlockHeaderHash(block.Header))
	prevHash := hex.EncodeToString(block.Header.PreviousHash)

	fmt.Println("Current hash:", currHash)
	fmt.Println("Previous hash:", prevHash)
	txFilter := TxValidationFlags(block.Metadata.Metadata[BlockMetadataIndex_TRANSACTIONS_FILTER])
	for i := 0; i < len(block.Metadata.Metadata[BlockMetadataIndex_TRANSACTIONS_FILTER]); i++ {
		fmt.Println("Transaction", i, "status:", txFilter.Flag(i), "size:", len(block.Data.Data[i]))
	}
	if len(block.Metadata.Metadata[BlockMetadataIndex_TRANSACTIONS_FILTER]) == 0 {
		fmt.Println("TRANSACTIONS_FILTER section is empty")
	}
	signatureMD := block.Metadata.Metadata[BlockMetadataIndex_SIGNATURES]
	if len(signatureMD) == 0 {
		fmt.Println("No signatures in block")
	} else {
		md := &common.Metadata{}
		if err := proto.Unmarshal(signatureMD, md); err !=nil {
			fmt.Println("Signature metadata is malformed:", err)
		} else {
			fmt.Println("Blocked contains", len(md.Signatures), "signatures")
		}
	}
}

type asn1Header struct {
	Number       *big.Int
	PreviousHash []byte
	DataHash     []byte
}

func BlockHeaderBytes(b *common.BlockHeader) []byte {
	asn1Header := asn1Header{
		PreviousHash: b.PreviousHash,
		DataHash:     b.DataHash,
		Number:       new(big.Int).SetUint64(b.Number),
	}
	result, err := asn1.Marshal(asn1Header)
	if err != nil {
		// Errors should only arise for types which cannot be encoded, since the
		// BlockHeader type is known a-priori to contain only encodable types, an
		// error here is fatal and should not be propogated
		panic(err)
	}
	return result
}

func BlockHeaderHash(b *common.BlockHeader) []byte {
	sum := sha256.Sum256(BlockHeaderBytes(b))
	return sum[:]
}
