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

func TestPacket_writePacketHeader(t *testing.T) {
	tests := []struct {
		name            string
		psize           int
		pcount          byte
		symbolSize      uint32
		want            []byte // The headers written to the start of the packet
		wantCoeffOffset int
		wantErr         bool
	}{
		{
			name:       "Normal case, big psize",
			psize:      1e6,
			pcount:     maxPcount,
			symbolSize: 999963,
			want: []byte{
				0x62, 0x0f, 0x42, 0x3c, // subatom header
				0x10,                                           // up to coeff_vector (includes first bit)
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // coeff vector ...
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // ... and byte align
			},
			wantCoeffOffset: 40,
		},
		{
			name:       "Max pcount",
			psize:      1500,
			pcount:     maxPcount,
			symbolSize: 1464,
			want: []byte{
				0x61, 0x05, 0xd9, // subatom header
				0x10,                                           // up to coeff_vector (includes first bit)
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // coeff vector ...
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // ... and byte align
			},
			wantCoeffOffset: 32,
		},
		{
			name:       "Smallest possible psize for headers (does not check valid pcount/payload size)",
			psize:      2 + 1,
			pcount:     0,
			symbolSize: 0,
			want: []byte{
				0x60, 0x01, // subatom header
				0x10, // whole packet header (since len(coeff_vector) == 0)
			},
			wantCoeffOffset: 24,
		},
		{
			name:       "Valid pcount, smallest possible psize for headers",
			psize:      2 + 14,
			pcount:     100,
			symbolSize: 0,
			want: []byte{
				0x60, 0x0e, // subatom header
				0x10,                                           // up to coeff_vector (includes first bit)
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // coeff vector ...
				0x00, 0x00, 0x00, 0x00, 0x00, // ... and byte align
			},
			wantCoeffOffset: 24,
		},
		{
			name:       "No byte align required",
			psize:      2 + 4,
			pcount:     17,
			symbolSize: 0,
			want: []byte{
				0x60, 0x04, // subatom header
				0x10,             // up to coeff_vector (includes first bit)
				0x00, 0x00, 0x00, // coeff vector
			},
			wantCoeffOffset: 24,
		},
		{
			name:       "Packet too small for headers",
			psize:      2 + 12,
			pcount:     100,
			symbolSize: 0,
			wantErr:    true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pkt := NewPacketFromEncoderConfig(tt.psize, nil)

			written, gotCoeffOffset, err := pkt.writePacketHeader(tt.pcount, tt.symbolSize)

			actualBytes := pkt.Serialize() // []byte

			if (err != nil) != tt.wantErr {
				t.Errorf("packet.writePacketHeader() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr {
				return
			}

			if written != len(tt.want) {
				t.Errorf("packet.writePacketHeader() wrote %d bytes, want %d", written, len(tt.want))
			}

			if gotCoeffOffset != tt.wantCoeffOffset {
				t.Errorf("coeff vector starts at bit %d, expected %d", gotCoeffOffset, tt.wantCoeffOffset)
			}

			// Some test cases have the header as the whole packet. Need to
			// handle them separately to avoid indexing issues.
			if len(tt.want) == len(actualBytes) {
				if !reflect.DeepEqual(actualBytes, tt.want) {
					t.Errorf(
						"packet.writePacketHeader() packet headers did not match expected. Want %X; got %X",
						tt.want, actualBytes,
					)
				}
				return
			}

			if gotHeader := actualBytes[:written]; !reflect.DeepEqual(gotHeader, tt.want) {
				t.Errorf(
					"packet.writePacketHeader() packet headers did not match expected. Want %X; got %X",
					tt.want, gotHeader,
				)
			}

			if !reflect.DeepEqual(actualBytes[written:], make([]byte, tt.psize-written)) {
				t.Error("packet.writePacketHeader() body was modified")
			}
		})
	}
}

func TestPacket_AddSourceSymbol(t *testing.T) {
	type offsets struct {
		codedSymbol int
		coeffVec    int
	}

	tests := []struct {
		name      string
		pkt       *Packet
		offsets   offsets
		symbolIdx byte
		symbol    []byte
		want      []byte
		wantErr   bool
	}{
		{
			name:      "Fresh new packet",
			pkt:       NewPacketFromEncoderConfig(10, nil),
			offsets:   offsets{2, 7},
			symbolIdx: 0,
			symbol:    []byte{0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA},
			want: []byte{
				0x01, 0x00,
				0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA,
			},
		},
		{
			name:    "Add symbol 0 to packet with existing symbols in it",
			offsets: offsets{1, 2},
			pkt: &Packet{
				Data: []byte{
					0xdf, // 0b11011111
					0xff, 0xff, 0xff,
				},
			},
			symbolIdx: 0,
			symbol:    []byte{0xff, 0xff, 0xff},
			want: []byte{
				0xff,
				0x00, 0x00, 0x00,
			},
		},
		{
			name:    "Add symbol 5 to packet with existing symbols in it",
			offsets: offsets{1, 0},
			pkt: &Packet{
				Data: []byte{
					0xaa, // 0b10101010
					0xaa, 0xaa, 0xaa,
				},
			},
			symbolIdx: 5,
			symbol:    []byte{0xff, 0x00, 0xff},
			want: []byte{
				0xae,             // 0b10101110
				0x55, 0xaa, 0x55, // 0x55 = 0b01010101
			},
		},
		{
			name:    "Add symbol 35 to packet with existing symbols in it; starts in 2nd byte",
			offsets: offsets{6, 9},
			pkt: &Packet{
				Data: []byte{
					0xde, 0xad, 0xbe, 0xef, 0xfe, 0xe5, // 0xe5 = 0b11100101
					0xf0, // 0b11110000
				},
			},
			symbolIdx: 35,
			symbol:    []byte{0xc3}, // 0b11000011
			want: []byte{
				0xde, 0xad, 0xbe, 0xef, 0xfe, 0xed, // 0xed = 0b11101101
				0x33, // 0b00110011
			},
		},
		{
			name:    "Symbol larger than coded_symbol",
			offsets: offsets{1, 6},
			pkt: &Packet{
				Data: []byte{
					0x02,
					0xf0,
				},
			},
			symbolIdx: 1,
			symbol:    []byte{0xff, 0xff, 0xff},
			want: []byte{
				0x03,
				0x0f,
			},
		},
		{
			name:    "Symbol smaller than coded_symbol",
			offsets: offsets{1, 6},
			pkt: &Packet{
				Data: []byte{
					0x02,
					0xff, 0xff, 0xff,
				},
			},
			symbolIdx: 1,
			symbol:    []byte{0xf0},
			want: []byte{
				0x03,
				0x0f, 0xff, 0xff,
			},
		},
		{
			name:    "coded_symbol size 0",
			offsets: offsets{3, 7},
			pkt: &Packet{
				Data: []byte{0xfa, 0xce, 0x42},
			}, // length == codedSymbolStartByte
			symbolIdx: 0,
			symbol:    []byte{0xff},
			want:      []byte{0xfa, 0xce, 0x42}, // untouched
			wantErr:   true,
		},
		{
			name:    "Symbol already in packet",
			offsets: offsets{1, 4},
			pkt: &Packet{
				Data: []byte{0x01, 0x00},
			},
			symbolIdx: 3,
			symbol:    []byte{0xff},
			want:      []byte{0x01, 0x00}, // untouched
			wantErr:   true,
		},
		{
			name:    "Negative coeff_vector offset",
			offsets: offsets{1, -1},
			pkt: &Packet{
				Data: []byte{0xaa, 0xaa, 0xaa, 0xaa},
			},
			symbolIdx: 5,
			symbol:    []byte{0xff, 0x00, 0xff},
			want:      []byte{0xaa, 0xaa, 0xaa, 0xaa}, // untouched
			wantErr:   true,
		},
		{
			name:    "Negative coded_symbol offset",
			offsets: offsets{-1, 0},
			pkt: &Packet{
				Data: []byte{0xaa, 0xaa, 0xaa, 0xaa},
			},
			symbolIdx: 5,
			symbol:    []byte{0xff, 0x00, 0xff},
			want:      []byte{0xaa, 0xaa, 0xaa, 0xaa}, // untouched
			wantErr:   true,
		},
		{
			name:    "coded_symbol offset larger than packet",
			offsets: offsets{5, 7},
			pkt: &Packet{
				Data: []byte{0xaa, 0xaa, 0xaa, 0xaa},
			},
			symbolIdx: 5,
			symbol:    []byte{0xff, 0x00, 0xff},
			want:      []byte{0xaa, 0xaa, 0xaa, 0xaa}, // untouched
			wantErr:   true,
		},
		{
			name:    "coded_symbol starts before coeff_vector ends",
			offsets: offsets{1, 7},
			pkt: &Packet{
				Data: []byte{0x00, 0x00, 0x00, 0x00},
			},
			symbolIdx: 1, // offset + idx == 8. Not in range [0,7], therefore needs another byte
			symbol:    []byte{0xff, 0x00, 0xff},
			want:      []byte{0x00, 0x00, 0x00, 0x00}, // untouched
			wantErr:   true,
		},
		{
			name:    "Empty packet",
			offsets: offsets{1, 0},
			pkt: &Packet{
				Data: []byte{},
			},
			symbolIdx: 5,
			symbol:    []byte{0xff, 0x00, 0xff},
			want:      []byte{}, // untouched
			wantErr:   true,
		},
		{
			name:    "Empty symbol",
			offsets: offsets{1, 0},
			pkt: &Packet{
				Data: []byte{0xaa, 0xaa, 0xaa, 0xaa},
			},
			symbolIdx: 5,
			symbol:    []byte{},
			want:      []byte{0xae, 0xaa, 0xaa, 0xaa}, // coded_symbol untouched
		},
		{
			name:    "Nil symbol",
			offsets: offsets{1, 0},
			pkt: &Packet{
				Data: []byte{0xaa, 0xaa, 0xaa, 0xaa},
			},
			symbolIdx: 5,
			symbol:    nil,
			want:      []byte{0xae, 0xaa, 0xaa, 0xaa}, // coded_symbol untouched
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.pkt.AddSourceSymbol(tt.offsets.codedSymbol, tt.offsets.coeffVec, tt.symbolIdx, tt.symbol)

			if (err != nil) != tt.wantErr {
				t.Errorf("packet.AddSourceSymbol() error = %v, wantErr %t", err, tt.wantErr)
				return
			}

			if !reflect.DeepEqual(tt.pkt.Data, tt.want) {
				t.Errorf("packet.AddSourceSymbol() result = %X, want = %X", tt.pkt.Data, tt.want)
			}
		})
	}
}

func Test_addCoeffToVector(t *testing.T) {
	tests := []struct {
		name          string
		packet        *Packet
		coeffVecStart int
		symbolIdx     byte
		want          *Packet
		wantErr       bool
	}{
		{
			"Symbol 0",
			&Packet{Data: []byte{0x00}},
			0, 0,
			&Packet{Data: []byte{0x80}},
			false,
		},
		{
			"Symbol 1",
			&Packet{Data: []byte{0x00}},
			0, 1,
			&Packet{Data: []byte{0x40}},
			false,
		},
		{
			"Symbol 2",
			&Packet{Data: []byte{0x00}},
			0, 2,
			&Packet{Data: []byte{0x20}},
			false,
		},
		{
			"Symbol 3",
			&Packet{Data: []byte{0x00}},
			0, 3,
			&Packet{Data: []byte{0x10}},
			false,
		},
		{
			"Symbol 4",
			&Packet{Data: []byte{0x00}},
			0, 4,
			&Packet{Data: []byte{0x08}},
			false,
		},
		{
			"Symbol 5",
			&Packet{Data: []byte{0x00}},
			0, 5,
			&Packet{Data: []byte{0x04}},
			false,
		},
		{
			"Symbol 6",
			&Packet{Data: []byte{0x00}},
			0, 6,
			&Packet{Data: []byte{0x02}},
			false,
		},
		{
			"Symbol 7",
			&Packet{Data: []byte{0x00}},
			0, 7,
			&Packet{Data: []byte{0x01}},
			false,
		},
		{
			"Symbol 8",
			&Packet{Data: []byte{0x00, 0x00}},
			0, 8,
			&Packet{Data: []byte{0x00, 0x80}},
			false,
		},
		{
			"Symbol 30",
			&Packet{Data: []byte{0x00, 0x00, 0x00, 0x00}},
			0, 30,
			&Packet{Data: []byte{0x00, 0x00, 0x00, 0x02}},
			false,
		},
		{
			name:          "Existing data untouched",
			packet:        &Packet{Data: []byte{0xde, 0x8d, 0xbe, 0xef}},
			coeffVecStart: 0,
			symbolIdx:     10,
			want:          &Packet{Data: []byte{0xde, 0xad, 0xbe, 0xef}},
		},
		{
			name:          "Non-zero coeffVecStart",
			packet:        &Packet{Data: []byte{0xde, 0x8d, 0xbe, 0xef}},
			coeffVecStart: 6,
			symbolIdx:     4,
			want:          &Packet{Data: []byte{0xde, 0xad, 0xbe, 0xef}},
		},
		{
			name: "Max symbolIdx",
			packet: &Packet{
				Data: []byte{
					0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA,
					0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA,
					0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA,
					0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA,
					0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA,
				}},
			coeffVecStart: 5,
			symbolIdx:     maxPcount - 1,
			want: &Packet{
				Data: []byte{
					0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA,
					0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA,
					0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA,
					0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA,
					0xbA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA,
				}},
		},
		{
			name:          "Coeff already in vector: error",
			packet:        &Packet{Data: []byte{0xde, 0xad, 0x8, 0xef}},
			coeffVecStart: 20,
			symbolIdx:     0,
			want:          &Packet{Data: []byte{0xde, 0xad, 0x8, 0xef}},
			wantErr:       true,
		},
		{
			name:    "Empty packet",
			packet:  &Packet{},
			wantErr: true,
			want:    &Packet{}, // unchanged
		},
		{
			name:          "coeffVecStart negative",
			packet:        &Packet{Data: []byte{0xde, 0xad, 0xbe, 0xef}},
			coeffVecStart: -1,
			wantErr:       true,
			want:          &Packet{Data: []byte{0xde, 0xad, 0xbe, 0xef}}, // unchanged
		},
		// symbolIdx too low not possible with an unsigned type as '0' is a valid value.
		{
			name: "symbolIdx too high",
			packet: &Packet{
				Data: []byte{
					0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA,
					0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA,
					0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA,
					0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA,
					0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA,
				}},
			symbolIdx: maxPcount,
			wantErr:   true,
			want: &Packet{ // unchanged
				Data: []byte{0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA,
					0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA,
					0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA,
					0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA,
					0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA,
				}},
		},
		{
			name:          "Packet too small",
			packet:        &Packet{Data: []byte{0xde, 0xad, 0xbe, 0xef}},
			coeffVecStart: 31,
			symbolIdx:     1,
			wantErr:       true,
			want:          &Packet{Data: []byte{0xde, 0xad, 0xbe, 0xef}}, // unchanged
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.packet.addCoeffToVector(tt.coeffVecStart, tt.symbolIdx)

			if (err != nil) != tt.wantErr {
				t.Errorf("packet.addCoeffToVector() error = %v, wantErr %t", err, tt.wantErr)
			}

			if !reflect.DeepEqual(tt.packet.Data, tt.want.Data) {
				t.Errorf("packet.addCoeffToVector() result = %X, want = %X", tt.packet.Data, tt.want.Data)
			}
		})
	}
}

func Test_calculatePacketHeaderSize(t *testing.T) {
	tests := []struct {
		name       string
		pcount     byte
		symbolSize uint32
		want       int
	}{
		{"Medium-sized symbols", 88, 60e3, 3 + 12},
		{"Big symbols", 167, 500e3, 4 + 22},
		{"Massive symbols", 255, 17e6, 5 + 33},
		{"Zero everything", 0, 0, 2 + 1},

		// Cross-check constants we've defined with this func
		{"Min everything", minPcount, minSymbolSize, 2 + minPacketHeaderSize},
		{"Max everything", maxPcount, maxSymbolSize, 4 + maxPacketHeaderSize},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := calculatePacketHeaderSize(tt.pcount, tt.symbolSize); got != tt.want {
				t.Errorf("calculatePacketHeaderSize() = %d, want %d", got, tt.want)
			}
		})
	}
}

func Test_calculateCoeffVectorSize(t *testing.T) {
	tests := []struct {
		name   string
		pcount byte
		want   int
	}{
		{"all fit in 1 byte", 8, 8},
		{"spill over into 2 (take ceil)", 9, 9},
		{"zero pcount", 0, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := calculateCoeffVectorSize(tt.pcount); got != tt.want {
				t.Errorf("calculateCoeffVectorSize() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_calculatePsize(t *testing.T) {
	tests := []struct {
		name       string
		pcount     byte
		symbolSize uint32
		want       int
		wantErr    bool
	}{
		{"Medium-sized symbols", 88, 60e3, 3 + 12 + 60e3, false},
		{"Big symbols", 167, 500e3, 4 + 22 + 500e3, false},
		{"Massive symbols", 255, maxSymbolSize, 4 + 33 + maxSymbolSize, false},
		{"Zero everything", 0, 0, 0, true},
		{"Psize would be too large", 255, math.MaxUint32, 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := calculatePsize(tt.pcount, tt.symbolSize)

			if (err != nil) != tt.wantErr {
				t.Errorf("calculatePsize() wantErr %t, got error: %v", tt.wantErr, err)
			}

			// don't check reported value on errors
			if err != nil {
				return
			}

			if got != tt.want {
				t.Errorf("calculatePsize() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_calculateSymbolSize(t *testing.T) {
	tests := []struct {
		name      string
		pcount    byte
		blockSize uint32
		want      uint32
		wantErr   bool
	}{
		{"Divisible", 10, 1000, 100, false},
		{"Not divisible: take ceil", 10, 1001, 101, false},
		{"Audio segment-ish", 88, 60e3, 682, false},
		{"720p segment-ish", 167, 500e3, 2995, false},
		{"4K segment-ish", 255, 17e6, 66667, false},
		{"pcount == blockSize", 60, 60, 1, false},
		{"Minimum all", minPcount, minBlockSize, 1, false},
		{"Maximum all", maxPcount, maxBlockSize, 4210715, false},

		{"pcount > blockSize", 10, 9, 0, true},
		{"blockSize == 0", 10, 0, 0, true},
		{"pcount == 0", 0, 1000, 0, true},
		{"blockSize too large for given pcount", 1, maxSymbolSize + 1, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := calculateSymbolSize(tt.pcount, tt.blockSize)
			if (err != nil) != tt.wantErr {
				t.Errorf("calculateSymbolSize() wantErr %t, got %v", tt.wantErr, err)
			}

			// don't check reported value on errors
			if tt.wantErr {
				return
			}

			if got != tt.want {
				t.Errorf("calculateSymbolSize() = %v, want %v", got, tt.want)
			}
		})
	}
}
