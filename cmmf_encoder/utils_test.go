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
	"reflect"
	"testing"
)

func Test_writeUint32NumBytes(t *testing.T) {
	tests := []struct {
		name           string
		buffer         []byte
		number         uint32
		numBytes       int
		expectedBuffer []byte
	}{
		{
			name:           "Write 1 byte",
			buffer:         make([]byte, 1),
			number:         0x2A,
			numBytes:       1,
			expectedBuffer: []byte{0x2A},
		},
		{
			name:           "Write 2 bytes",
			buffer:         make([]byte, 2),
			number:         0x1122,
			numBytes:       2,
			expectedBuffer: []byte{0x11, 0x22},
		},
		{
			name:           "Write 3 bytes",
			buffer:         make([]byte, 3),
			number:         0x1A2b3c,
			numBytes:       3,
			expectedBuffer: []byte{0x1A, 0x2B, 0x3C},
		},
		{
			name:           "Write 4 bytes",
			buffer:         make([]byte, 4),
			number:         0x01020304,
			numBytes:       4,
			expectedBuffer: []byte{0x01, 0x02, 0x03, 0x04},
		},
		{
			name:           "Truncate number",
			buffer:         make([]byte, 2),
			number:         0x01020304,
			numBytes:       2,
			expectedBuffer: []byte{0x03, 0x04},
		},
		{
			name:           "Pad number - obviously this one should work",
			buffer:         make([]byte, 2),
			number:         0x01,
			numBytes:       2,
			expectedBuffer: []byte{0x00, 0x01},
		},
		{
			name:           "Buffer bigger than length of int",
			buffer:         []byte{0xAA, 0xBB},
			number:         0x11,
			numBytes:       1,
			expectedBuffer: []byte{0x11, 0xBB},
		},
		{
			name:           "Buffer bigger than 4 bytes",
			buffer:         []byte{0xAA, 0xBB, 0xCC, 0xDD, 0xEE, 0xFF},
			number:         0x11,
			numBytes:       4,
			expectedBuffer: []byte{0x00, 0x00, 0x00, 0x11, 0xEE, 0xFF},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			writeUint32NumBytes(tt.buffer, tt.number, tt.numBytes)
			if !reflect.DeepEqual(tt.buffer, tt.expectedBuffer) {
				t.Errorf("writeUint32NumBytes() = %v, want %v", tt.buffer, tt.expectedBuffer)
			}
		})
	}
}

func Test_writeUnalignedBytes(t *testing.T) {
	tests := []struct {
		name     string
		existing byte
		offset   byte
		data     []byte
		want     []byte
	}{
		{
			name:     "1 byte to write; offset 1",
			existing: 0x00, // 0b00000000
			offset:   1,
			data:     []byte{0xff},       // 0b11111111
			want:     []byte{0x7f, 0x80}, // 0b01111111 0b10000000
		},
		{
			name:     "1 byte to write; offset (2) matches existing data",
			existing: 0x40, // 0b01000000
			offset:   2,
			data:     []byte{0xbf},       // 0b10111111
			want:     []byte{0x6f, 0xc0}, // 0b01101111 0b11000000
		},
		{
			name:     "1 byte to write; offset (6) matches existing data",
			existing: 0x1c, // 0b00011100
			offset:   6,
			data:     []byte{0xc0},       // 0b11000000
			want:     []byte{0x1f, 0x00}, // 0b00011111 0b00000000
		},
		{
			name:     "1 byte to write; existing data before offset (3) kept, data after overwritten",
			existing: 0xff, // 0b11111111
			offset:   3,
			data:     []byte{0xaa},       // 0b10101010
			want:     []byte{0xf5, 0x40}, // 0b11110101 0b01000000
		},
		{
			name:     "1 byte to write; existing data before offset (4) kept, data after overwritten",
			existing: 0xaa, // 0b10101010
			offset:   4,
			data:     []byte{0x55},       // 0b01010101
			want:     []byte{0xa5, 0x50}, // 0b10100101 0b01010000
		},
		{
			name:     "1 byte to write; existing data before offset (5) kept, data after overwritten",
			existing: 0x07, // 0b00000111
			offset:   5,
			data:     []byte{0x00},       // 0b00000000
			want:     []byte{0x00, 0x00}, // 0b00000000 0b00000000
		},
		{
			name:     "1 byte to write; existing data before offset (7) kept, data after overwritten",
			existing: 0x81, // 0b10000001
			offset:   7,
			data:     []byte{0x01},       // 0b00000001
			want:     []byte{0x80, 0x02}, // 0b10000000 0b00000010
		},
		{
			name:     "0 offset means we overwrite all existing data",
			existing: 0x00, // 0b00000000
			offset:   0,
			data:     []byte{0xff}, // 0b11111111
			want:     []byte{0xff}, // 0b11111111
		},
		{
			name:     "Offset larger 7 - modulo so works as a bit index",
			existing: 0xff,               // 0b11111111
			offset:   133,                // should be the same as offset = 5
			data:     []byte{0x01},       // 0b00000001
			want:     []byte{0xf8, 0x08}, // 0b11111000 0b00001000
		},
		{
			name:     "Multi-byte data, offset = 1",
			existing: 0xff, // 0b11111111
			offset:   1,
			data:     []byte{0xaa, 0xaa},       // 0b10101010 0b10101010
			want:     []byte{0xd5, 0x55, 0x00}, // 0b11010101 0b01010101 0b00000000
		},
		{
			name:     "Multi-byte data, offset = 4",
			existing: 0xab,
			offset:   4,
			data:     []byte{0x12, 0x34},
			want:     []byte{0xa1, 0x23, 0x40},
		},
		{
			name:     "Multi-byte data, offset = 7",
			existing: 0xff, // 0b11111111
			offset:   7,
			data:     []byte{0x55, 0x55},       // 0b01010101 0b01010101
			want:     []byte{0xfe, 0xaa, 0xaa}, // 0b11111110 0b10101010 0b10101010
		},
		{
			name:     "Multi-byte data, offset = 0 (same as input data)",
			existing: 0xff, // 0b11111111
			offset:   0,
			data:     []byte{0xaa, 0xaa}, // 0b10101010 0b10101010
			want:     []byte{0xaa, 0xaa}, // 0b10101010 0b10101010
		},
		{
			name:     "Zero length data will just return the existing",
			existing: 0xf0, // 0b11110000
			offset:   4,
			data:     []byte{},
			want:     []byte{0xf0}, // 0b11110000
		},
		{
			name:     "Nil data will just return the existing",
			existing: 0xf0, // 0b11110000
			offset:   4,
			data:     []byte{},
			want:     []byte{0xf0}, // 0b11110000
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := writeUnalignedBytes(tt.existing, tt.offset, tt.data); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("writeUnalignedBytes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_splitByte(t *testing.T) {
	tests := []struct {
		name      string
		octet     byte
		offset    byte
		wantLeft  byte
		wantRight byte
	}{
		{
			name:      "(0b11111111, 0) -> (0b11111111, 0b00000000)",
			octet:     0xff,
			offset:    0,
			wantLeft:  0xff,
			wantRight: 0x00,
		},
		{
			name:      "(0b11111111, 1) -> (0b01111111, 0b10000000)",
			octet:     0xff,
			offset:    1,
			wantLeft:  0x7f,
			wantRight: 0x80,
		},
		{
			name:      "(0b11111111, 2) -> (0b00111111, 0b11000000)",
			octet:     0xff,
			offset:    2,
			wantLeft:  0x3f,
			wantRight: 0xc0,
		},
		{
			name:      "(0b11111111, 3) -> (0b00011111, 0b11100000)",
			octet:     0xff,
			offset:    3,
			wantLeft:  0x1f,
			wantRight: 0xe0,
		},
		{
			name:      "(0b11111111, 4) -> (0b00001111, 0b11110000)",
			octet:     0xff,
			offset:    4,
			wantLeft:  0xf,
			wantRight: 0xf0,
		},
		{
			name:      "(0b11111111, 5) -> (0b00000111, 0b11111000)",
			octet:     0xff,
			offset:    5,
			wantLeft:  0x7,
			wantRight: 0xf8,
		},
		{
			name:      "(0b11111111, 6) -> (0b00000011, 0b11111100)",
			octet:     0xff,
			offset:    6,
			wantLeft:  0x3,
			wantRight: 0xfc,
		},
		{
			name:      "(0b11111111, 7) -> (0b00000001, 0b11111110)",
			octet:     0xff,
			offset:    7,
			wantLeft:  0x1,
			wantRight: 0xfe,
		},
		{
			name:      "(0b10111101, 4) -> (0b00001011, 0b11010000)",
			octet:     0xbd,
			offset:    4,
			wantLeft:  0x0b,
			wantRight: 0xd0,
		},
		{
			name:      "(0b10101111, 6) -> (0b00000010, 0b10111100)",
			octet:     0xaf,
			offset:    6,
			wantLeft:  0x2,
			wantRight: 0xbc,
		},
		{
			name:      "offset could be more than 7, but only the modulo should matter",
			octet:     0xff,
			offset:    97, // should be the same as offset = 1
			wantLeft:  0x7f,
			wantRight: 0x80,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotLeft, gotRight := splitByte(tt.octet, tt.offset)
			if gotLeft != tt.wantLeft {
				t.Errorf("splitByte() gotLeft = %v, want %v", gotLeft, tt.wantLeft)
			}
			if gotRight != tt.wantRight {
				t.Errorf("splitByte() gotRight = %v, want %v", gotRight, tt.wantRight)
			}
		})
	}
}

func Test_ceilDivision(t *testing.T) {
	tests := []struct {
		numerator int
		divisor   int
		want      int
	}{
		{2, 2, 1},
		{3, 2, 2},
		{0, 2, 0},
		{1e6, 100, 1e4},
		{1e6, 99, 10102},
		{999999, 99, 10101},

		// Very po-or.
		{1, 0, 0},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("ceilDivision(%d/%d) == %d", tt.numerator, tt.divisor, tt.want), func(t *testing.T) {
			if got := ceilDivision(tt.numerator, tt.divisor); got != tt.want {
				t.Errorf("ceilDivision() = %d, want %d", got, tt.want)
			}
		})
	}
}
