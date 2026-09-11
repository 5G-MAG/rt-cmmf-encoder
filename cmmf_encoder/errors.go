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

// checkValidBlockSize checks if `blockSize` is a valid number of bytes for a
// source block. After you call this function it'll be safe to cast `blockSize`
// to a uint32.
//
// For now, this function only implements a subset of the spec.
func checkValidBlockSize(blockSize int) error {
	if blockSize < minBlockSize || blockSize > maxBlockSize {
		return fmt.Errorf("invalid blockSize: %d not in range [%d, %d]",
			blockSize, minBlockSize, maxBlockSize)
	}
	return nil
}

// checkValidPcount checks if `pcount` is a valid number of symbols into which
// to split a source block. After you call this function it'll be safe to cast
// `pcount` to a uint8.
//
// For now, this function only implements a subset of the spec.
func checkValidPcount(pcount int) error {
	if pcount < minPcount || pcount > maxPcount {
		return fmt.Errorf("invalid pcount: %d not in range [%d, %d]",
			pcount, minPcount, maxPcount)
	}
	return nil
}

// checkValidSymbolSize checks if `symbolSize` is a valid number of bytes for a
// symbol. After you call this function it'll be safe to cast `symbolSize` to a
// uint32.
//
// For now, this function only implements a subset of the spec.
func checkValidSymbolSize(symbolSize int) error {
	if symbolSize < minSymbolSize || symbolSize > maxSymbolSize {
		return fmt.Errorf("invalid symbolSize: %d not in range [%d, %d]",
			symbolSize, minSymbolSize, maxSymbolSize)
	}
	return nil
}

// checkValidPsize returns an error if the `psize` is invalid.
//
// It will check that the `psize` is within the range of possible values and
// that it is large enough to contain all headers and the coded_symbol.
//
// For now, this function only implements a subset of the spec.
func checkValidPsize(psize int, pcount byte, symbolSize uint32) error {
	if psize < minPsize || psize > maxPsize {
		return fmt.Errorf("psize %d not in range [%d, %d]", psize, minPsize, maxPsize)
	}

	headSize := calculatePacketHeaderSize(pcount, symbolSize)

	if psize < headSize+int(symbolSize) {
		return fmt.Errorf("psize %d too small to fit pcount %d/symbolSize %d", psize, pcount, symbolSize)
	}

	return nil
}
