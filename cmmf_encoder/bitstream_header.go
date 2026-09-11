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
	"encoding/binary"
	"fmt"
	"log/slog"
)

type BitstreamHeader struct {
	SubatomID         SubatomID
	contentSourceSize uint64
	encoderConfig     *ETSIConfig
}

func NewBitstreamHeaderFromEncoderConfig(encoderConfig *ETSIConfig, contentSourceSize uint64) *BitstreamHeader {
	return &BitstreamHeader{
		SubatomID:         SUBATOM_ID_BITSTREAM_HEADER,
		contentSourceSize: contentSourceSize,
		encoderConfig:     encoderConfig,
	}
}

// prepareProfileInfo writes profile information, and optionally profile description, into a buffer
func (b *BitstreamHeader) prepareProfileInfo(profileType string, profileDescription *string) ([]byte, error) {
	totalBits := 1 + 4 + (8 * len(profileType)) + 32
	totalBytes := ceilDivision(totalBits, 8)

	profileTypeBytes := []byte(profileType)
	profileDescriptionBytes := make([]byte, 4)
	if profileDescription == nil || *profileDescription == "" {
		slog.Debug("BitstreamHeader: profile description is empty")
	} else {
		n := copy(profileDescriptionBytes, []byte(*profileDescription)[:min(4, len(*profileDescription))])
		slog.Debug("BitstreamHeader: read profile description bytes", "n", n, "bytes", profileDescriptionBytes)
	}
	profileTypeBytes = append(profileTypeBytes, profileDescriptionBytes...)

	// write b_profile_information_present
	profileMetadata := uint8(0)
	profileMetadata = profileMetadata | 0x80
	// write the size of the profile type
	profileMetadata = profileMetadata | (uint8(len(profileType)&0x0F) << 3)

	written := writeUnalignedBytes(profileMetadata, 5, profileTypeBytes)
	if len(written) != totalBytes {
		return nil, fmt.Errorf("expected to write %v bytes, wrote %v", totalBytes, written)
	}

	return written, nil
}

func (b *BitstreamHeader) Serialize() []byte {
	totalBits := 1 + 4 + (8 * len(*b.encoderConfig.Profile)) + 32
	bitstreamHeaderSizeInBits := 84 + totalBits
	bitstreamHeaderSize := ceilDivision(bitstreamHeaderSizeInBits, 8)

	header := make([]byte, bitstreamHeaderSize+calculateSubatomHeaderSize(uint32(bitstreamHeaderSize)))

	offset := writeSubatomHeader(header, b.SubatomID, uint32(bitstreamHeaderSize))

	// NOTE: all bit addresses in comments below are for the bitstream_header only.
	// They do not consider the subatom header, so also don't reflect the offset within the header slice.
	// Ranges are inclusive.

	// Bits 0-63: content_source_size
	binary.BigEndian.PutUint64(header[offset:], b.contentSourceSize)
	offset += 8

	// Bits 64-66: `content_source_type` (3 bits).
	// Bits 67-68: `reserved` & `b_content_source_split`. Both set to 0
	// Bits 69-72: `code_type`. As of ETSI v1.1.1, xCD-1 code type is represented as 0.
	// Bit 73: `b_rfc5052` - set to 0.
	header[offset] = 1 << 5
	offset += 1 // offset within the byte now 2

	// Bits 74-81: `block_count_minus1`. We only allow for 1 block, so set to 0.
	// Bit 82: `b_content_block_separate_sources`. Set to 0.
	offset += 1 // offset within the byte now 3

	// Add the profile information
	dolbyProfileInfo, err := b.prepareProfileInfo(*b.encoderConfig.Profile, b.encoderConfig.ProfileDescription)
	if err != nil {
		panic(err)
	}
	offset += copy(header[offset:], writeUnalignedBytes(header[offset], 3, dolbyProfileInfo))

	// Bit 160: `b_block_cc_encrypted`. Set to 0.
	offset += 1

	return header
}
