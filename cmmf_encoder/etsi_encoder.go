/******************************************************************************
 * 5G-MAG Reference Tools: CMMF Encoder
 ******************************************************************************
 * Copyright: (C)2024-2026 Dolby Laboratories Inc.
 * License: 5G-MAG Public License v1
 *
 * Licensed under the License terms and conditions for use, reproduction, and
 * distribution of 5G-MAG software (the “License”).  You may not use this file
 * except in compliance with the License.  You may obtain a copy of the License at
 * https://www.5g-mag.com/reference-tools.  Unless required by applicable law or
 * agreed to in writing, software distributed under the License is distributed on
 * an “AS IS” BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
 * or implied.
 *
 * See the License for the specific language governing permissions and limitations
 * under the License.
 */

package cmmf_encoder

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"math"
	"strings"
)

// ErrNoContent is returned by NewETSIEncoder when the source reader is empty.
//
// This is only a variable because Go will not allow it to be a constant.
// DO NOT MODIFY.
var ErrNoContent = errors.New("no content was provided")

// EncodingParameters contains information about the way in which a block will
// be encoded by the encoder.
type EncodingParameters struct {
	// Psize is the number of bytes in each encoded packet, including headers.
	Psize int

	// Pcount is the number of source symbols into which the source block was
	// split. It also represents the minimum number of packets a decoder needs
	// to receive to decode.
	Pcount byte

	// BlockSize is the number of bytes in the source block.
	BlockSize uint32

	// SymbolSize is the number of bytes in each source symbol into which we
	// split the block.
	SymbolSize uint32

	// PaddingSize is the number of bytes of padding we add to the source block
	// so that it would be evenly divisible by `Pcount`.
	PaddingSize uint32
}

// ETSIEncoder implements the Encoder interface. It will read the encoder configuration
// from a "CMMF Encoder configuration" file - refer to section E of the ETSI Technical
// specification document
// (see https://www.etsi.org/deliver/etsi_ts/103900_103999/103973/01.02.02_60/ts_103973v010202p.pdf)
// In addition, it will read all data from the source block before it starts encoding packets.
type ETSIEncoder struct {
	// psize is the number of bytes in each encoded packet, including headers.
	psize int

	// pcount is the number of source symbols into which we split the source block.
	// It also represents the minimum number of packets a decoder needs to
	// receive to decode.
	pcount byte

	// paddingSize specifies the number of bytes of padding we add to source
	// block so that it is evenly divisible by pcount.
	paddingSize uint32

	// The number of bytes in each source symbol into which we split the block.
	symbolSize uint32

	// sourceBlock is the original data (block) that is to be encoded.
	sourceBlock []byte

	// sourceSymbols store the original block in its packetised form. It
	// provides simple access a required source symbol.
	sourceSymbols [][]byte

	// coeffLookupByVariantID
	coeffLookupByVariantID map[string][][]byte

	etsiConfig *ETSIConfig
}

// getPacketIDsFromBitString returns a list of packet indices from a bit-string
func getPacketIDsFromBitString(bitString string) []byte {
	bstr := strings.ReplaceAll(bitString, "b", "")
	pktIds := make([]byte, 0)

	for i, r := len(bstr)-1, 0; i >= 0; i, r = i-1, r+1 {
		if bstr[i] == '1' {
			pktIds = append(pktIds, byte(r))
		}
	}
	return pktIds
}

// makeCoefficientLUTPerVariantID creates a lookup table of variantID to a list of packet indices per symbol
func (enc *ETSIEncoder) makeCoefficientLUTPerVariantID() {
	for _, info := range enc.etsiConfig.XCDInfo.CoefficientInfo.CoefficientVectorMaps {
		if _, ok := enc.coeffLookupByVariantID[*info.VariantID]; !ok {
			listOfPktIDs := make([][]byte, 0)
			for _, coeffVector := range info.CoefficientVectors {
				pktIds := getPacketIDsFromBitString(coeffVector)
				listOfPktIDs = append(listOfPktIDs, pktIds)
			}
			enc.coeffLookupByVariantID[*info.VariantID] = listOfPktIDs
		}
	}
}

// EncodeVariant encodes a bitstream for a given variant ID
func (enc *ETSIEncoder) EncodeVariant(variantID string, w io.Writer) error {
	coeffs, ok := enc.coeffLookupByVariantID[variantID]
	if !ok {
		return fmt.Errorf("no coefficient for variant %s", variantID)
	}

	if len(coeffs) != len(enc.sourceSymbols) {
		return fmt.Errorf("mismatch in number of coefficient symbols")
	}

	written := 0
	// first, write the preamble - which consists of the Sync subatom, the Bitstream Header, and the Block Header
	syncSubatom := NewSyncFromEncoderConfig(enc.etsiConfig)
	n, err := w.Write(syncSubatom.Serialize())
	if err != nil {
		return fmt.Errorf("error writing Sync subatom: %w", err)
	}
	written += n

	bitstreamHeader := NewBitstreamHeaderFromEncoderConfig(enc.etsiConfig, uint64(enc.symbolSize)*uint64(enc.pcount))
	n, err = w.Write(bitstreamHeader.Serialize())
	if err != nil {
		return fmt.Errorf("error writing Bitstream header: %w", err)
	}
	written += n

	blockHeader := NewBlockHeaderFromEncoderConfig(enc.etsiConfig, enc.pcount, uint32(len(enc.sourceBlock)), enc.symbolSize)
	n, err = w.Write(blockHeader.Serialize())
	if err != nil {
		return fmt.Errorf("error writing Block header: %w", err)
	}

	// now, encode the packets
	for i, c := range coeffs {
		slog.Info("encoding packet", "id", i)
		encoded, err := enc.Encode(c...)
		if err != nil {
			return fmt.Errorf("error encoding packet: %d", i)
		}
		n, err = w.Write(encoded)
		if err != nil {
			return fmt.Errorf("error writing encoded packet %d: %w", i, err)
		}
		slog.Info("wrote encoded packet bytes", "id", i, "size", n)
	}

	return nil
}

func (enc *ETSIEncoder) Encode(sourceSymbols ...byte) ([]byte, error) {
	packet := NewPacketFromEncoderConfig(enc.psize, enc.etsiConfig)

	err := enc.encodeInPlace(packet, sourceSymbols...)
	if err != nil {
		return nil, err
	}

	return packet.Data, nil
}

func (enc *ETSIEncoder) encodeInPlace(packet *Packet, sourceSymbols ...byte) error {
	if len(packet.Data) != enc.psize {
		return fmt.Errorf("packet is not the correct size (%d): expected %d", len(packet.Data), enc.psize)
	}

	codedSymbolStartByte, coeffVecStartBit, err := packet.writePacketHeader(enc.pcount, enc.symbolSize)
	if err != nil {
		return err
	}

	for _, symbol := range sourceSymbols {
		if int(symbol) >= len(enc.sourceSymbols) {
			return fmt.Errorf("source symbol index out of bounds: %d", symbol)
		}
		data := enc.sourceSymbols[symbol]
		err := packet.AddSourceSymbol(codedSymbolStartByte, coeffVecStartBit, symbol, data)
		if err != nil {
			return err
		}
	}

	return nil
}

// GetEncodingParameters returns key encoding configuration values that this
// encoder will use to encode the block.
func (enc *ETSIEncoder) GetEncodingParameters() EncodingParameters {
	return EncodingParameters{
		Psize:       enc.psize,
		Pcount:      enc.pcount,
		BlockSize:   uint32(len(enc.sourceBlock)),
		SymbolSize:  enc.symbolSize,
		PaddingSize: enc.paddingSize,
	}
}

func (enc *ETSIEncoder) GetParsedETSIConfig() *ETSIConfig {
	return enc.etsiConfig
}

// _newBasicEncoderReadSource reads data from `source` into `enc.sourceBlock`.
//
// It is intended to only be used by a BasicEncoder constructor - do not use
// on a fully initialised encoder.
func (enc *ETSIEncoder) _newBasicEncoderReadSource(source io.Reader) error {
	data, err := io.ReadAll(source)
	if err != nil {
		return fmt.Errorf("failed to read all data from source reader: %w", err)
	} else if len(data) == 0 {
		return ErrNoContent
	}
	enc.sourceBlock = data

	return nil
}

// _newBasicEncoderAssignSourceSymbols instantiates `enc`'s `sourceSymbol` slice
// based on the contents of `enc.sourceBlock`.
//
// It is intended to only be used by a BasicEncoder constructor - do not use
// on a fully initialised encoder.
func (enc *ETSIEncoder) _newBasicEncoderAssignSourceSymbols() error {
	if len(enc.sourceBlock) == 0 {
		return fmt.Errorf("source block is uninitialised or empty")
	}
	if enc.pcount == 0 || enc.symbolSize == 0 {
		return fmt.Errorf("encoding parameters are uninitialised")
	}
	enc.sourceSymbols = make([][]byte, enc.pcount)

	// hook up the pointers to the source symbols.
	for i := 0; i < len(enc.sourceSymbols); i++ {
		start := i * int(enc.symbolSize)
		end := (i + 1) * int(enc.symbolSize)
		if end > len(enc.sourceBlock) {
			end = len(enc.sourceBlock)
		}
		enc.sourceSymbols[i] = enc.sourceBlock[start:end]
	}

	return nil
}

// NewETSIEncoder creates an ETSIEncoder
func NewETSIEncoder(sourceBytesReader, encoderConfigReader io.Reader) (*ETSIEncoder, error) {
	encoder := &ETSIEncoder{}

	etsiEncoderConfig, err := NewValidatedETSIConfig(encoderConfigReader)
	if err != nil {
		return nil, fmt.Errorf("error parsing the ETSI encoder config file: %w", err)
	}
	encoder.etsiConfig = etsiEncoderConfig

	if err := encoder._newBasicEncoderReadSource(sourceBytesReader); err != nil {
		return nil, err
	}

	encParams, err := CalculateEncodingParameters(len(encoder.sourceBlock), encoder.etsiConfig)
	if err != nil {
		return nil, fmt.Errorf("error calculating encoding parameters: %w", err)
	}
	encoder.psize = encParams.Psize
	encoder.pcount = encParams.Pcount
	encoder.paddingSize = encParams.PaddingSize
	encoder.symbolSize = encParams.SymbolSize

	if err := encoder._newBasicEncoderAssignSourceSymbols(); err != nil {
		return nil, fmt.Errorf("error assigning source symbols: %w", err)
	}

	encoder.coeffLookupByVariantID = make(map[string][][]byte)
	encoder.makeCoefficientLUTPerVariantID()

	return encoder, nil
}

func CalculateEncodingParameters(blockSize int, encoderConfig *ETSIConfig) (EncodingParameters, error) {
	// pcount lookup
	pcount := 0
	for _, blockSizeMap := range encoderConfig.BlockSizeToBlockNumSymbolsMaps {
		if blockSize > *blockSizeMap.BlockSizeMin && blockSize < *blockSizeMap.BlockSizeMax {
			pcount = *blockSizeMap.BlockNumSymbols
			break
		}
	}

	if pcount > 255 {
		return EncodingParameters{}, fmt.Errorf("pcount (%d) exceeds maximum allowed value (255)", pcount)
	}

	return CalculateEncodingParametersFixedPcount(blockSize, uint8(pcount))
}

// CalculateEncodingParametersFixedPcount calculates the coding parameters
// given the size of the original data block and a fixed `pcount`.
func CalculateEncodingParametersFixedPcount(blockSize int, pcount uint8) (EncodingParameters, error) {
	if err := checkValidBlockSize(blockSize); err != nil {
		return EncodingParameters{}, err
	}
	blockSizeUint32 := uint32(blockSize)

	if err := checkValidPcount(int(pcount)); err != nil {
		return EncodingParameters{}, fmt.Errorf("calculating pcount: %w", err)
	}

	symbolSize, err := calculateSymbolSize(pcount, blockSizeUint32)
	if err != nil {
		return EncodingParameters{}, fmt.Errorf("calculating symbolSize: %w", err)
	}

	psize, err := calculatePsize(pcount, symbolSize)
	if err != nil {
		return EncodingParameters{}, fmt.Errorf("calculating psize: %w", err)
	}

	paddingSize, err := calculatePaddingSize(symbolSize, blockSizeUint32, pcount)
	if err != nil {
		return EncodingParameters{}, fmt.Errorf("calculating paddingSize: %w", err)
	}

	encParams := EncodingParameters{
		Psize:       psize,
		Pcount:      pcount,
		BlockSize:   blockSizeUint32,
		SymbolSize:  symbolSize,
		PaddingSize: paddingSize,
	}
	return encParams, nil
}

// calculatePaddingSize returns the number of bytes that are needed to be added
// to a source block (length `blockSize`) in order to ensure that it can evenly
// split into `pcount` symbols of `symbolSize` bytes each.
func calculatePaddingSize(symbolSize, blockSize uint32, pcount byte) (uint32, error) {
	capacityOfAllSymbols := int(symbolSize) * int(pcount)
	paddingSize := capacityOfAllSymbols - int(blockSize)

	if paddingSize < 0 {
		return 0, fmt.Errorf("paddingSize was negative: %d", paddingSize)
	} else if paddingSize > math.MaxUint32 {
		return 0, fmt.Errorf("paddingSize too big for a uint32: %d", paddingSize)
	}

	return uint32(paddingSize), nil
}
