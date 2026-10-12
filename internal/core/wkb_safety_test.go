package core

import (
	"encoding/binary"
	"testing"
)

func nestedWKB(depth int) []byte {
	b := wkbPoint(1, 1, 2)
	for i := 0; i < depth; i++ {
		header := []byte{1, 7, 0, 0, 0, 1, 0, 0, 0}
		b = append(header, b...)
	}
	return b
}

func TestWKBRejectsUnboundedCountsAndNesting(t *testing.T) {
	for _, kind := range []uint32{2, 3, 4, 5, 6, 7} {
		b := make([]byte, 9)
		b[0] = 1
		binary.LittleEndian.PutUint32(b[1:5], kind)
		binary.LittleEndian.PutUint32(b[5:9], ^uint32(0))
		if _, err := parseWKB(b); err == nil {
			t.Fatalf("accepted impossible count for kind %d", kind)
		}
	}
	if _, err := parseWKB(nestedWKB(maxWKBDepth)); err == nil {
		t.Fatal("accepted unbounded nesting")
	}
	if _, err := parseWKB(nestedWKB(8)); err != nil {
		t.Fatalf("valid nesting rejected: %v", err)
	}
	// MultiPoint must not quietly discard a non-point child.
	if _, err := parseWKB(append([]byte{1, 4, 0, 0, 0, 1, 0, 0, 0}, nestedWKB(1)...)); err == nil {
		t.Fatal("accepted wrong multi-point member")
	}
}

func TestGeoPackageRejectsReservedEnvelope(t *testing.T) {
	for _, envelope := range []byte{5, 6, 7} {
		blob := append([]byte{'G', 'P', 0, 1 | envelope<<1, 0, 0, 0, 0}, wkbPoint(1, 1, 2)...)
		if _, err := parseGeoPackageBinary(blob); err == nil {
			t.Fatalf("accepted reserved envelope %d", envelope)
		}
	}
}

func FuzzWKB(f *testing.F) {
	f.Add(wkbPoint(1, 1, 2))
	f.Add(wkbLine(2, 2, 0, 0, 1, 1))
	f.Add(nestedWKB(4))
	f.Add([]byte{1, 2, 0, 0, 0, 255, 255, 255, 255})
	f.Fuzz(func(t *testing.T, b []byte) { _, _ = parseWKB(b) })
}

func FuzzGeoPackageBinary(f *testing.F) {
	f.Add(append([]byte{'G', 'P', 0, 1, 0, 0, 0, 0}, wkbPoint(1, 1, 2)...))
	f.Add([]byte("GP"))
	f.Fuzz(func(t *testing.T, b []byte) { _, _ = parseGeoPackageBinary(b) })
}
