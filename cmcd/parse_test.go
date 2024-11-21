package cmcd

import (
	"bufio"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"
)

var tests = map[string]Info{
	"testdata/simple": {
		Session: Session{ID: "6e2fb550-c457-11e9-bb97-0800200c9a66", PlayRate: RealTime},
	},
	"testdata/all_four": {
		Request: Request{Throughput: 25400},
		Object: Object{
			Bitrate:    3200,
			Duration:   4004 * time.Millisecond,
			Type:       ObjTypeVideo,
			TopBitrate: 6000,
		},
		Status:  Status{true, 15000},
		Session: Session{ID: "6e2fb550-c457-11e9-bb97-0800200c9a66", PlayRate: RealTime},
	},
	"testdata/booleans": {
		Status:  Status{true, 0},
		Request: Request{Startup: true},
		Session: Session{PlayRate: RealTime},
	},
	"testdata/range": {
		Request: Request{NextRange: [2]int{12323, 48763}},
		Object:  Object{Duration: 4004 * time.Millisecond},
		Session: Session{PlayRate: RealTime},
	},
	"testdata/custom": {
		Object:  Object{Duration: 4004 * time.Millisecond},
		Session: Session{PlayRate: RealTime},
		Custom: map[string]any{
			"com.example.javasucks.int": 500,
			"stringy":                   "yamum",
			"aBool":                     true,
		},
	},
}

func TestParse(t *testing.T) {
	for name, want := range tests {
		t.Run(path.Base(name), func(t *testing.T) {
			pt, err := readParseTest(name)
			if err != nil {
				t.Fatal(err)
			}
			info, err := ParseInfo(pt.query)
			if err != nil {
				t.Errorf("info from query: %v", err)
			}
			if !reflect.DeepEqual(want, info) {
				t.Errorf("info from query: want %+v, got %+v", want, info)
				t.Log(want.Encode())
				t.Log(info.Encode())
			}

			// now try re-encoding to see if we get the same back again.
			// trim stray commas used for testing parser.
			swant := strings.Split(strings.Trim(pt.query, ","), ",")
			sgot := strings.Split(info.Encode(), ",")
			sort.Strings(swant)
			sort.Strings(sgot)
			if !reflect.DeepEqual(sgot, swant) {
				t.Errorf("re-encode: got %v, want %v", sgot, swant)
			}
		})
	}
}

type parseTest struct {
	header http.Header
	query  string
	json   []byte
}

func readParseTest(name string) (parseTest, error) {
	f, err := os.Open(name)
	if err != nil {
		return parseTest{}, err
	}
	defer f.Close()
	var pt parseTest
	pt.header = make(http.Header)
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		if sc.Text() == "" {
			continue
		} else if strings.HasPrefix(sc.Text(), "#") {
			continue // skip comments
		}
		if strings.HasPrefix(sc.Text(), "?CMCD=") {
			raw := strings.TrimPrefix(sc.Text(), "?CMCD=")
			q, err := url.QueryUnescape(raw)
			if err != nil {
				return pt, fmt.Errorf("parse cmcd query: %w", err)
			}
			pt.query = q
			continue
		}
		if strings.HasPrefix(sc.Text(), "{") {
			pt.json = sc.Bytes()
		}
		before, after, found := strings.Cut(sc.Text(), ":")
		if !found {
			return pt, fmt.Errorf("invalid case: %s", sc.Text())
		}
		pt.header.Set(before, strings.TrimSpace(after))
	}
	return pt, sc.Err()
}

/*
			custom := make(map[string]any) // TODO
			if !reflect.DeepEqual(tt.want.Custom, custom) {
				t.Errorf("custom attributes: want %+v, got %+v", tt.want.Custom, custom)
			}
		})
	}
}
*/
