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
)

func Test_calculateEncodingParametersFixedPcount(t *testing.T) {
	tests := []struct {
		name      string
		blockSize int
		pcount    uint8
		want      EncodingParameters
		wantErr   bool
	}{
		// Note: some test cases covered by Test_calculateEncodingParameters - won't duplicate
		{"100KB, split into 200 packets", 100e3, 200, EncodingParameters{529, 200, 100e3, 500, 0}, false},
		{"100KB, split into 3 packets", 100e3, 3, EncodingParameters{33339, 3, 100e3, 33334, 2}, false},
		{"Block size equals pcount", 20, 20, EncodingParameters{7, 20, 20, 1, 0}, false},

		{"Pcount too small", minBlockSize, 1, EncodingParameters{}, true},
		{"Block size smaller than pcount (padding spans multiple packets)", 20, 24, EncodingParameters{}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CalculateEncodingParametersFixedPcount(tt.blockSize, tt.pcount)

			if (err != nil) != tt.wantErr {
				t.Errorf("calculateEncodingParameters() error = %v, wantErr %t", err, tt.wantErr)
			}

			// Don't check other values on error
			if tt.wantErr {
				return
			}

			if got.Pcount != tt.want.Pcount {
				t.Errorf("pcount incorrect = %d, want %d", got.Pcount, tt.want.Pcount)
			}
			if got.Psize != tt.want.Psize {
				t.Errorf("psize incorrect = %d, want %d", got.Psize, tt.want.Psize)
			}
			if got.PaddingSize != tt.want.PaddingSize {
				t.Errorf("paddingSize incorrect: got = %d, want %d", got.PaddingSize, tt.want.PaddingSize)
			}
			if got.SymbolSize != tt.want.SymbolSize {
				t.Errorf("symbolSize incorrect = %d, want %d", got.SymbolSize, tt.want.SymbolSize)
			}
			if got.BlockSize != uint32(tt.want.BlockSize) {
				t.Errorf("blockSize incorrect = %d, want %d", got.BlockSize, tt.want.BlockSize)
			}
		})
	}
}

func Test_calculatePaddingSize(t *testing.T) {
	tests := []struct {
		name       string
		symbolSize uint32
		blockSize  uint32
		pcount     byte
		want       uint32
		wantErr    bool
	}{
		{"Some padding", 1500, 149990, 100, 10, false},
		{"Zero padding", 1500, 150000, 100, 0, false},
		{"Negative padding", 1500, 150001, 100, 0, true},

		{"Capacity too big for a uint32", math.MaxUint32, 1e6, maxPcount, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := calculatePaddingSize(tt.symbolSize, tt.blockSize, tt.pcount)
			if (err != nil) != tt.wantErr {
				t.Errorf("calculatePaddingSize() error = %v, want %t", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if got != tt.want {
				t.Errorf("calculatePaddingSize() = %d, want %d", got, tt.want)
			}
		})
	}
}
