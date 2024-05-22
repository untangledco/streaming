package m3u8

import (
	"os"
	"path"
	"testing"
)

func TestLex(t *testing.T) {
	files := []string{"testdata/master.m3u8"}
	for _, name := range files {
		fname := path.Base(name)
		t.Run(fname, func(t *testing.T) {
			f, err := os.Open(name)
			if err != nil {
				t.Fatal(err)
			}
			defer f.Close()
			lexer := newLexer(f)
			go lexer.run()
			for it := range lexer.items {
				t.Log(it)
				if it.typ == itemError {
					t.Error(it.val)
				}
			}
		})
	}
}
