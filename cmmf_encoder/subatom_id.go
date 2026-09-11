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
	"fmt"
)

// SubatomID is an enum containing enumerations for each type of subatom.
type SubatomID uint8

const (
	SUBATOM_ID_RESERVED                 SubatomID = iota
	SUBATOM_ID_SYNC                               = 0x1
	SUBATOM_ID_BITSTREAM_HEADER                   = 0x2
	SUBATOM_ID_ENCODER_CONTENT_INFO               = 0x3
	SUBATOM_ID_MEDIA_SEGMENT_INFO                 = 0x4
	SUBATOM_ID_BLOCK_HEADER                       = 0x5
	SUBATOM_ID_PACKET                             = 0x6
	SUBATOM_ID_CHUNKED_SUBATOM                    = 0x7
	SUBATOM_ID_BLOCK_GROUP_DIRECTORY              = 0x8
	SUBATOM_ID_PACKET_HEADER_ONLY                 = 0x9
	SUBATOM_ID_PACKET_GROUP                       = 0x10
	SUBATOM_ID_MULTI_BLOCK_PACKET_GROUP           = 0x11
)

func (sID SubatomID) String() string {
	switch sID {
	case SUBATOM_ID_RESERVED:
		return "reserved"
	case SUBATOM_ID_SYNC:
		return "sync"
	case SUBATOM_ID_BITSTREAM_HEADER:
		return "bitstream_header"
	case SUBATOM_ID_ENCODER_CONTENT_INFO:
		return "encoder_content_info"
	case SUBATOM_ID_MEDIA_SEGMENT_INFO:
		return "media_segment_info"
	case SUBATOM_ID_BLOCK_HEADER:
		return "block_header"
	case SUBATOM_ID_PACKET:
		return "packet"
	case SUBATOM_ID_CHUNKED_SUBATOM:
		return "chunked_subatom"
	case SUBATOM_ID_BLOCK_GROUP_DIRECTORY:
		return "block_group_directory"
	case SUBATOM_ID_PACKET_HEADER_ONLY:
		return "packet_header_only"
	case SUBATOM_ID_PACKET_GROUP:
		return "packet_group"
	case SUBATOM_ID_MULTI_BLOCK_PACKET_GROUP:
		return "multi_block_packet_group"
	}
	return "unknown"
}

func SubatomIDFromUint32(val uint32) (SubatomID, error) {
	switch val {
	case 0x00:
		return SUBATOM_ID_RESERVED, nil
	case 0x1:
		return SUBATOM_ID_SYNC, nil
	case 0x2:
		return SUBATOM_ID_BITSTREAM_HEADER, nil
	case 0x3:
		return SUBATOM_ID_ENCODER_CONTENT_INFO, nil
	case 0x4:
		return SUBATOM_ID_MEDIA_SEGMENT_INFO, nil
	case 0x5:
		return SUBATOM_ID_BLOCK_HEADER, nil
	case 0x6:
		return SUBATOM_ID_PACKET, nil
	case 0x7:
		return SUBATOM_ID_CHUNKED_SUBATOM, nil
	case 0x8:
		return SUBATOM_ID_BLOCK_GROUP_DIRECTORY, nil
	case 0x9:
		return SUBATOM_ID_PACKET_HEADER_ONLY, nil
	case 0x10:
		return SUBATOM_ID_PACKET_GROUP, nil
	case 0x11:
		return SUBATOM_ID_MULTI_BLOCK_PACKET_GROUP, nil
	}

	return SUBATOM_ID_RESERVED, fmt.Errorf("invalid value: %v", val)
}
