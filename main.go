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
	"github.com/hyperledger/fabric/protos/common"
	"gopkg.in/alecthomas/kingpin.v2"
)

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
