//go:build !noasm && !appengine

// ******************************************************************************
// * 5G-MAG Reference Tools: CMMF Encoder
// ******************************************************************************
// * Copyright: (C)2024-2026 Dolby Laboratories Inc.
// * License: 5G-MAG Public License v1
// *
// * Licensed under the License terms and conditions for use, reproduction, and
// * distribution of 5G-MAG software (the “License”).  You may not use this file
// * except in compliance with the License.  You may obtain a copy of the License at
// * https://www.5g-mag.com/reference-tools.  Unless required by applicable law or
// * agreed to in writing, software distributed under the License is distributed on
// * an “AS IS” BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
// * or implied.
// *
// * See the License for the specific language governing permissions and limitations
// * under the License.
//

// Modified from code which is Copyright 2015, Klaus Post.
// Licenced under MIT Licence. See LICENCE_MIT.txt file for details.

// func sSE2XorSlice(in, out []byte, size int)
// XOR `in` and `out` 16 bytes at a time and store result in `out`.
// `size` is the length of the smallest of the two slices.
TEXT ·sSE2XorSlice(SB), 7, $0
	MOVQ in+0(FP), SI     // SI: &in
	MOVQ out+24(FP), DX   // DX: &out
	MOVQ size+48(FP), R9   // R9: size
	SHRQ $4, R9           // size / 16
	CMPQ R9, $0
	JEQ  done_xor_sse2

loopback_xor_sse2:
	MOVOU (SI), X0          // in[x]
	MOVOU (DX), X1          // out[x]
	PXOR  X0, X1
	MOVOU X1, (DX)
	ADDQ  $16, SI           // in+=16
	ADDQ  $16, DX           // out+=16
	SUBQ  $1, R9
	JNZ   loopback_xor_sse2

done_xor_sse2:
	RET
