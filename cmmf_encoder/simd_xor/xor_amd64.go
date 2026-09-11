//go:build !noasm && !appengine

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

package simd_xor

import "github.com/klauspost/cpuid/v2"

// sSE2Capable says whether or not the current CPU supports the SSE2 instruction set.
var sSE2Capable bool

func init() {
	sSE2Capable = cpuid.CPU.Supports(cpuid.SSE2)
}

//go:noescape
func sSE2XorSlice(in, out []byte, size int)

// xor src into dst.
// It will only xor the number of bytes in the smallest of the two slices.
func XOR(dst, src []byte) {
	size := len(src)
	if len(dst) < size {
		size = len(dst)
	}

	var done int
	if sSE2Capable {
		sSE2XorSlice(src, dst, size)

		// sSE2XorSlice xors in chunks of 16. It will stop if there's not
		// a full 16 bytes left to xor.
		done = (size >> 4) << 4
	}
	remain := size - done
	if remain > 0 {
		for i := done; i < size; i++ {
			dst[i] ^= src[i]
		}
	}
}
