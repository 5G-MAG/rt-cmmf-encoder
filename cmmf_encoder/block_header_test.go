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
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func blockHeader_getETSIConfigWithProfile() *ETSIConfig {
	return &ETSIConfig{}
}

func TestBlockHeader_Serialize(t *testing.T) {
	type args struct {
		pCount        byte
		blockSize     uint32
		symbolSize    uint32
		encoderConfig *ETSIConfig
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		{
			name: "Realistic-ish case",
			args: args{255, 5e6, 19608, blockHeader_getETSIConfigWithProfile()},
			want: []byte{
				0x50, 0x0c, // subatom header
				0x00,                   // hardcoded
				0x00, 0x4c, 0x4b, 0x40, // block_size
				0x00, 0x00, 0x4c, 0x98, // block_symbol_size
				0x3f, 0xc0, 0x18, // final hardcoded bytes including pcount, block mask
			},
		},
		{
			name: "Max block size",
			args: args{255, math.MaxUint32, 16843009, blockHeader_getETSIConfigWithProfile()},
			want: []byte{
				0x50, 0x0c, // subatom header
				0x00,                   // hardcoded
				0xff, 0xff, 0xff, 0xff, // block_size
				0x01, 0x01, 0x01, 0x01, // block_symbol_size
				0x3f, 0xc0, 0x18, // final hardcoded bytes including pcount, block mask
			},
		},
		{
			name: "Don't check invalid symbol size (>block)",
			args: args{150, 2e6, 2.5e6, blockHeader_getETSIConfigWithProfile()},
			want: []byte{
				0x50, 0x0c, // subatom header
				0x00,                   // hardcoded
				0x00, 0x1e, 0x84, 0x80, // block_size
				0x00, 0x26, 0x25, 0xa0, // block_symbol_size
				0x25, 0x80, 0x18, // final hardcoded bytes including pcount, block mask
			},
		},
		{
			name: "Don't check insufficient pcount",
			args: args{12, 12e6, 100, blockHeader_getETSIConfigWithProfile()},
			want: []byte{
				0x50, 0x0c, // subatom header
				0x00,                   // hardcoded
				0x00, 0xb7, 0x1b, 0x00, // block_size
				0x00, 0x00, 0x00, 0x64, // block_symbol_size
				0x03, 0x00, 0x18, // final hardcoded bytes including pcount, block mask
			},
		},
		{
			name: "Max all",
			args: args{255, math.MaxUint32, math.MaxUint32, blockHeader_getETSIConfigWithProfile()},
			want: []byte{
				0x50, 0x0c, // subatom header
				0x00,                   // hardcoded
				0xff, 0xff, 0xff, 0xff, // block_size
				0xff, 0xff, 0xff, 0xff, // block_symbol_size
				0x3f, 0xc0, 0x18, // final hardcoded bytes including pcount, block mask
			},
		},
		{
			name: "None all",
			args: args{0, 0, 0, blockHeader_getETSIConfigWithProfile()},
			want: []byte{
				0x50, 0x0c, // subatom header
				0x00,                   // hardcoded
				0x00, 0x00, 0x00, 0x00, // block_size
				0x00, 0x00, 0x00, 0x00, // block_symbol_size
				0x00, 0x00, 0x18, // final hardcoded bytes including pcount, block mask
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			blockHeader := NewBlockHeaderFromEncoderConfig(
				tt.args.encoderConfig,
				tt.args.pCount,
				tt.args.blockSize,
				tt.args.symbolSize,
			)
			serialized := blockHeader.Serialize()
			assert.Equal(t, tt.want, serialized)
		})
	}

}
