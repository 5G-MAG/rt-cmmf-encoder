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
	"reflect"
	"testing"
)

func Test_getSubatomHeader_AtomTypes(t *testing.T) {
	tests := []struct {
		name        string
		subatomType SubatomID
		subatomSize uint32
		want        []byte
	}{
		{
			name:        "Sync atom",
			subatomType: SUBATOM_ID_SYNC,
			subatomSize: 1,
			want:        []byte{0x10, 0x01},
		},
		{
			name:        "Bitstream header",
			subatomType: SUBATOM_ID_BITSTREAM_HEADER,
			subatomSize: 10,
			want:        []byte{0x20, 0x0a},
		},
		{
			name:        "Encoder content info",
			subatomType: SUBATOM_ID_ENCODER_CONTENT_INFO,
			subatomSize: 30,
			want:        []byte{0x30, 0x1e},
		},
		{
			name:        "Media Segment Info",
			subatomType: SUBATOM_ID_MEDIA_SEGMENT_INFO,
			subatomSize: 25,
			want:        []byte{0x40, 0x19},
		},
		{
			name:        "Block header",
			subatomType: SUBATOM_ID_BLOCK_HEADER,
			subatomSize: 15,
			want:        []byte{0x50, 0x0f},
		},
		{
			name:        "Packet",
			subatomType: SUBATOM_ID_PACKET,
			subatomSize: 42000,
			want:        []byte{0x61, 0xa4, 0x10},
		},
		{
			name:        "Chunked subatom",
			subatomType: SUBATOM_ID_CHUNKED_SUBATOM,
			subatomSize: 1<<32 - 1,
			want:        []byte{0x73, 0xFF, 0xFF, 0xFF, 0xFF},
		},
		{
			name:        "Block directory",
			subatomType: SUBATOM_ID_BLOCK_GROUP_DIRECTORY,
			subatomSize: 40,
			want:        []byte{0x80, 0x28},
		},
		{
			name:        "Packet header only",
			subatomType: SUBATOM_ID_PACKET_HEADER_ONLY,
			subatomSize: 25,
			want:        []byte{0x90, 0x19},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotHeader := getSubatomHeader(tt.subatomType, tt.subatomSize)
			if !reflect.DeepEqual(gotHeader, tt.want) {
				t.Errorf("getSubatomHeader() gotHeader = %v, want %v", gotHeader, tt.want)
			}
		})
	}
}

func Test_writeSubatomHeader_AtomTypes(t *testing.T) {
	tests := []struct {
		name        string
		buffer      []byte
		subatomType SubatomID
		subatomSize uint32
		want        []byte
	}{
		// The 'want' values were worked out by hand using the bitstream.
		{
			name:        "Sync atom",
			buffer:      make([]byte, 2),
			subatomType: SUBATOM_ID_SYNC,
			subatomSize: 1,
			want:        []byte{0x10, 0x01},
		},
		{
			name:        "Bitstream header",
			buffer:      make([]byte, 2),
			subatomType: SUBATOM_ID_BITSTREAM_HEADER,
			subatomSize: 10,
			want:        []byte{0x20, 0x0a},
		},
		{
			name:        "Encoder content info",
			buffer:      make([]byte, 2),
			subatomType: SUBATOM_ID_ENCODER_CONTENT_INFO,
			subatomSize: 30,
			want:        []byte{0x30, 0x1e},
		},
		{
			name:        "Media Segment Info",
			buffer:      make([]byte, 2),
			subatomType: SUBATOM_ID_MEDIA_SEGMENT_INFO,
			subatomSize: 25,
			want:        []byte{0x40, 0x19},
		},
		{
			name:        "Block header",
			buffer:      make([]byte, 2),
			subatomType: SUBATOM_ID_BLOCK_HEADER,
			subatomSize: 15,
			want:        []byte{0x50, 0x0f},
		},
		{
			name:        "Packet",
			buffer:      make([]byte, 3),
			subatomType: SUBATOM_ID_PACKET,
			subatomSize: 42000,
			want:        []byte{0x61, 0xa4, 0x10},
		},
		{
			name:        "Chunked subatom",
			buffer:      make([]byte, 5),
			subatomType: SUBATOM_ID_CHUNKED_SUBATOM,
			subatomSize: 1<<32 - 1,
			want:        []byte{0x73, 0xFF, 0xFF, 0xFF, 0xFF},
		},
		{
			name:        "Block directory",
			buffer:      make([]byte, 2),
			subatomType: SUBATOM_ID_BLOCK_GROUP_DIRECTORY,
			subatomSize: 40,
			want:        []byte{0x80, 0x28},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := writeSubatomHeader(tt.buffer, tt.subatomType, tt.subatomSize)
			if n != len(tt.want) {
				t.Errorf("wrote fewer bytes (%d) than expected (%d)", n, len(tt.want))
			} else if !reflect.DeepEqual(tt.buffer, tt.want) {
				t.Errorf("writeSubatomHeader() = %v, want %v", tt.buffer, tt.want)
			}
		})
	}
}

func Test_getSubatomHeader_AtomSizes(t *testing.T) {
	tests := []struct {
		name        string
		subatomType SubatomID
		subatomSize uint32
		want        []byte
	}{
		{
			name:        "Size 0",
			subatomType: SUBATOM_ID_PACKET,
			subatomSize: 0,
			want:        []byte{0x60, 0x00},
		},
		{
			name:        "Max uint8",
			subatomType: SUBATOM_ID_PACKET,
			subatomSize: 255,
			want:        []byte{0x60, 0xFF},
		},
		{
			name:        "Just over uint8",
			subatomType: SUBATOM_ID_PACKET,
			subatomSize: 256,
			want:        []byte{0x61, 0x01, 0x00},
		},
		{
			name:        "Max uint16",
			subatomType: SUBATOM_ID_PACKET,
			subatomSize: 1<<16 - 1,
			want:        []byte{0x61, 0xFF, 0xFF},
		},
		{
			name:        "Just over uint16",
			subatomType: SUBATOM_ID_PACKET,
			subatomSize: 1 << 16,
			want:        []byte{0x62, 0x01, 0x00, 0x00},
		},
		{
			name:        "Max uint24",
			subatomType: SUBATOM_ID_PACKET,
			subatomSize: 1<<24 - 1,
			want:        []byte{0x62, 0xFF, 0xFF, 0xFF},
		},
		{
			name:        "Just over uint24",
			subatomType: SUBATOM_ID_PACKET,
			subatomSize: 1 << 24,
			want:        []byte{0x63, 0x01, 0x00, 0x00, 0x00},
		},
		{
			name:        "Max uint32",
			subatomType: SUBATOM_ID_PACKET,
			subatomSize: 1<<32 - 1,
			want:        []byte{0x63, 0xFF, 0xFF, 0xFF, 0xFF},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotHeader := getSubatomHeader(tt.subatomType, tt.subatomSize)
			if !reflect.DeepEqual(gotHeader, tt.want) {
				t.Errorf("getSubatomHeader() = %v, want %v", gotHeader, tt.want)
			}
		})
	}
}

func Test_writeSubatomHeader_AtomSizes(t *testing.T) {
	tests := []struct {
		name        string
		buffer      []byte
		subatomType SubatomID
		subatomSize uint32
		want        []byte
	}{
		{
			name:        "Size 0",
			buffer:      make([]byte, 2),
			subatomType: SUBATOM_ID_PACKET,
			subatomSize: 0,
			want:        []byte{0x60, 0x00},
		},
		{
			name:        "Max uint8",
			buffer:      make([]byte, 2),
			subatomType: SUBATOM_ID_PACKET,
			subatomSize: 255,
			want:        []byte{0x60, 0xFF},
		},
		{
			name:        "Just over uint8",
			buffer:      make([]byte, 3),
			subatomType: SUBATOM_ID_PACKET,
			subatomSize: 256,
			want:        []byte{0x61, 0x01, 0x00},
		},
		{
			name:        "Max uint16",
			buffer:      make([]byte, 3),
			subatomType: SUBATOM_ID_PACKET,
			subatomSize: 1<<16 - 1,
			want:        []byte{0x61, 0xFF, 0xFF},
		},
		{
			name:        "Just over uint16",
			buffer:      make([]byte, 4),
			subatomType: SUBATOM_ID_PACKET,
			subatomSize: 1 << 16,
			want:        []byte{0x62, 0x01, 0x00, 0x00},
		},
		{
			name:        "Max uint24",
			buffer:      make([]byte, 4),
			subatomType: SUBATOM_ID_PACKET,
			subatomSize: 1<<24 - 1,
			want:        []byte{0x62, 0xFF, 0xFF, 0xFF},
		},
		{
			name:        "Just over uint24",
			buffer:      make([]byte, 5),
			subatomType: SUBATOM_ID_PACKET,
			subatomSize: 1 << 24,
			want:        []byte{0x63, 0x01, 0x00, 0x00, 0x00},
		},
		{
			name:        "Max uint32",
			buffer:      make([]byte, 5),
			subatomType: SUBATOM_ID_PACKET,
			subatomSize: 1<<32 - 1,
			want:        []byte{0x63, 0xFF, 0xFF, 0xFF, 0xFF},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := writeSubatomHeader(tt.buffer, tt.subatomType, tt.subatomSize)
			if n != len(tt.want) {
				t.Errorf("wrote fewer bytes (%d) than expected (%d)", n, len(tt.want))
			} else if !reflect.DeepEqual(tt.buffer, tt.want) {
				t.Errorf("writeSubatomHeader() = %v, want %v", tt.buffer, tt.want)
			}
		})
	}
}

func Test_writeSubatomHeader_bufferTooBig(t *testing.T) {
	buffer := []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}

	writeSubatomHeader(buffer, SUBATOM_ID_PACKET, math.MaxUint32)
	want := []byte{0x63, 0xFF, 0xFF, 0xFF, 0xFF, 0x06, 0x07, 0x08}

	if !reflect.DeepEqual(buffer, want) {
		t.Errorf("writeSubatomHeader() with too big buffer: got = %v, want %v", buffer, want)
	}
}

func Test_writeSubatomHeader_PanicIfBufferTooSmall(t *testing.T) {
	defer func() {
		if panicVal := recover(); panicVal == nil {
			t.Errorf("writeSubatomHeader() did not panic when expected")
		}
	}()

	// Try to write 2 bytes to a 1 byte buffer
	writeSubatomHeader(make([]byte, 1), SUBATOM_ID_SYNC, 5)
}

func Test_calculateBytesRequiredToStoreNum(t *testing.T) {
	tests := []struct {
		name   string
		number uint32
		want   int
	}{
		{"0", 0, 1},
		{"max uint8", 1<<8 - 1, 1},
		{"just over uint8", 1 << 8, 2},
		{"max uint16", 1<<16 - 1, 2},
		{"just over uint16", 1 << 16, 3},
		{"max uint24", 1<<24 - 1, 3},
		{"just over uint24", 1 << 24, 4},
		{"max uint32", 1<<32 - 1, 4},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := calculateBytesRequiredToStoreNum(tt.number); got != tt.want {
				t.Errorf("calculateBytesRequiredToStoreNum() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_calculateSubatomHeaderSize(t *testing.T) {
	tests := []struct {
		name        string
		subatomSize uint32
		want        int
	}{
		{"0", 0, 2},
		{"max uint8", 1<<8 - 1, 2},
		{"just over uint8", 1 << 8, 3},
		{"max uint16", 1<<16 - 1, 3},
		{"just over uint16", 1 << 16, 4},
		{"max uint24", 1<<24 - 1, 4},
		{"just over uint24", 1 << 24, 5},
		{"max uint32", 1<<32 - 1, 5},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := calculateSubatomHeaderSize(tt.subatomSize); got != tt.want {
				t.Errorf("calculateSubatomHeaderSize() = %v, want %v", got, tt.want)
			}
		})
	}
}
