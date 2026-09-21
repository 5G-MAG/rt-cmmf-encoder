package simd_xor

import (
	"bytes"
	"math/rand"
	"testing"
)

// refXOR is the obvious byte-at-a-time XOR the SIMD path must agree with.
func refXOR(dst, src []byte) {
	size := len(src)
	if len(dst) < size {
		size = len(dst)
	}
	for i := 0; i < size; i++ {
		dst[i] ^= src[i]
	}
}

func TestXORMatchesReference(t *testing.T) {
	rng := rand.New(rand.NewSource(1))
	// sizes around the 16-byte SIMD chunk boundary, plus larger ones
	sizes := []int{0, 1, 2, 7, 15, 16, 17, 31, 32, 33, 63, 64, 65, 127, 1024, 4096, 16383, 16384}
	for _, n := range sizes {
		for _, m := range sizes {
			src := make([]byte, n)
			dstA := make([]byte, m)
			rng.Read(src)
			rng.Read(dstA)
			dstB := append([]byte(nil), dstA...)

			XOR(dstA, src)
			refXOR(dstB, src)
			if !bytes.Equal(dstA, dstB) {
				t.Fatalf("XOR mismatch for len(src)=%d len(dst)=%d", n, m)
			}
		}
	}
}

// TestXORUnalignedOffsets exercises non-16-byte-aligned slice starts, which the
// MOVOU (unaligned load) path must handle.
func TestXORUnalignedOffsets(t *testing.T) {
	rng := rand.New(rand.NewSource(2))
	backingSrc := make([]byte, 4096)
	backingDst := make([]byte, 4096)
	rng.Read(backingSrc)
	rng.Read(backingDst)

	for off := 0; off < 32; off++ {
		for _, n := range []int{16, 17, 48, 100} {
			src := backingSrc[off : off+n]
			dstA := append([]byte(nil), backingDst[off:off+n]...)
			dstB := append([]byte(nil), dstA...)

			XOR(dstA, src)
			refXOR(dstB, src)
			if !bytes.Equal(dstA, dstB) {
				t.Fatalf("XOR mismatch at offset %d, len %d", off, n)
			}
		}
	}
}

// TestXORDoesNotWriteBeyondSize guards the asm loop against running past the
// smaller of the two slices.
func TestXORDoesNotWriteBeyondSize(t *testing.T) {
	rng := rand.New(rand.NewSource(3))
	for _, n := range []int{1, 15, 16, 17, 33, 64, 65} {
		const guard = 64
		full := make([]byte, n+guard)
		rng.Read(full)
		tail := append([]byte(nil), full[n:]...)

		src := make([]byte, n)
		rng.Read(src)

		XOR(full[:n], src)

		if !bytes.Equal(full[n:], tail) {
			t.Fatalf("XOR wrote past size %d into the guard region", n)
		}
	}
}

func FuzzXOR(f *testing.F) {
	f.Add([]byte("hello"), []byte("world"))
	f.Add(bytes.Repeat([]byte{0xff}, 33), bytes.Repeat([]byte{0x0f}, 17))
	f.Fuzz(func(t *testing.T, dst, src []byte) {
		a := append([]byte(nil), dst...)
		b := append([]byte(nil), dst...)
		XOR(a, src)
		refXOR(b, src)
		if !bytes.Equal(a, b) {
			t.Fatalf("XOR mismatch: len(dst)=%d len(src)=%d", len(dst), len(src))
		}
	})
}
