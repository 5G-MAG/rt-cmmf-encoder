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

import "encoding/binary"

type BlockHeader struct {
	SubatomID     SubatomID
	pCount        uint8
	blockSize     uint32
	symbolSize    uint32
	encoderConfig *ETSIConfig
}

func NewBlockHeaderFromEncoderConfig(encoderConfig *ETSIConfig, pCount uint8, blockSize, symbolSize uint32) *BlockHeader {
	return &BlockHeader{
		SubatomID:     SUBATOM_ID_BLOCK_HEADER,
		pCount:        pCount,
		blockSize:     blockSize,
		symbolSize:    symbolSize,
		encoderConfig: encoderConfig,
	}
}

func (b *BlockHeader) Serialize() []byte {
	// In this implementation, the block header is of a fixed size.
	// We set a total of 93 bits, where the `block_num_symbols_bits_code` field is set to 0,
	// and the 6 `block_mask` bits are set to 000001
	blockHeaderSize := 12

	header := make([]byte, blockHeaderSize+calculateSubatomHeaderSize(uint32(blockHeaderSize)))

	offset := writeSubatomHeader(header, b.SubatomID, uint32(blockHeaderSize))

	// Bits 0-7: `block_index`, only 1 block, so set to 0.
	offset += 1

	// Bits 8-39: `block_size`
	binary.BigEndian.PutUint32(header[offset:], b.blockSize)
	offset += 4
	// Bits 40-71: `block_symbol_size`
	binary.BigEndian.PutUint32(header[offset:], b.symbolSize)
	offset += 4

	// Bits 72-73: `bns_bits`. Our current max pcount is 255, so need only 1
	// byte to store the `block_num_symbols`. Hence, set to 0.

	// Bits 74-81: `block_num_symbols`.
	copy(header[offset:], writeUnalignedBytes(0x00, 2, []byte{b.pCount}))
	offset += 1

	// Bit 82: `b_block_max_symbol_index_present`. Set to 0 for now.
	// Bits 83-84: `b_block_content_source_index_present` & `b_block_composite_sources`. Set to 0
	// Bit 85: `b_addl_block_coding_info_present`. Set to 0.

	// Bits 86-91: block_mask. Within the mask, we will only set the last bit (bit 91): symbols present.
	// Bit 92: b_sufficient_symbols_present. Since we put all packets needed to decode in a block, this is set to 1.
	offset += 1
	header[offset] = 0x18 // = 0001_1000

	return header
}
