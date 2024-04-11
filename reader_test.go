package jsonparser

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"
)

var inputs = []struct {
	path       string
	tokens     int
	allTokens  int
	whitespace int
}{
	{"canada", 223236, 334373, 33},
	{"citm_catalog", 85035, 135990, 1227563},
	{"twitter", 29573, 55263, 167931},
	{"code", 217707, 396293, 3},
	{"example", 710, 1297, 4246},
	{"sample", 5276, 8677, 518549},
}

func fixture(tb testing.TB, path string) *bytes.Reader {
	f, err := os.Open(filepath.Join("./testdata", path+".json.gz"))
	check(tb, err)
	defer f.Close()
	gz, err := gzip.NewReader(f)
	check(tb, err)
	buf, err := io.ReadAll(gz)
	check(tb, err)
	return bytes.NewReader(buf)
}

func check(tb testing.TB, err error) {
	if err != nil {
		tb.Helper()
		tb.Fatal(err)
	}
}

func BenchmarkCountWhitespace(b *testing.B) {
	var buf [8 << 10]byte
	for _, tc := range inputs {
		r := fixture(b, tc.path)
		b.Run(tc.path, func(b *testing.B) {
			b.ReportAllocs()
			b.SetBytes(r.Size())
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				r.Seek(0, 0)
				br := byteReader{
					data: buf[:0],
					r:    r,
				}
				got := countWhitespace(&br)
				if got != tc.whitespace {
					b.Fatalf("expected: %v, got: %v", tc.whitespace, got)
				}
			}
		})
	}
}

func TestFuck(t *testing.T) {
	var buf [8 << 10]byte
	for _, tc := range inputs {
		r := fixture(t, tc.path)
		t.Run(tc.path, func(b *testing.T) {
			br := byteReader{
				data: buf[:0],
				r:    r,
			}
			got := countWhitespace(&br)
			if got != tc.whitespace {
				b.Fatalf("expected: %v, got: %v", tc.whitespace, got)
			}
		})
	}
}

func countWhitespace(br *byteReader) int {
	n := 0
	w := br.window()
	for {
		for _, c := range w {
			if whitespace[c] {
				n++
			}
		}
		br.release(len(w))
		if br.extend() == 0 {
			return n
		}
		w = br.window()
		fmt.Println(w)
	}
}
