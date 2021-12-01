package radix

import (
	"bytes"
	crand "crypto/rand"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"testing"
)

func TestRadix(t *testing.T) {
	var min, max []byte

	inp := make(map[string]interface{})
	for i := 0; i < 1000; i++ {
		gen := generateUUID()
		inp[gen] = i
		bgen := []byte(gen)
		if bytes.Compare(bgen, min) < 0 || i == 0 {
			min = bgen
		}
		if bytes.Compare(bgen, max) > 0 || i == 0 {
			max = bgen
		}
	}

	r := NewFromMap(inp)
	if r.Len() != len(inp) {
		t.Fatalf("bad length: %v %v", r.Len(), len(inp))
	}

	var count int = 0
	r.Walk(func(k []byte, v interface{}) bool {
		//println(k)
		count++
		return false
	})

	if count != 1000 {
		t.Fatalf("wrong count of keys: %d != %d", count, 1000)
	}

	for k, v := range inp {
		out, ok := r.Get([]byte(k))
		if !ok {
			t.Fatalf("missing key: %v", k)
		}
		if out != v {
			t.Fatalf("value mis-match: %v %v", out, v)
		}
	}

	// Check min and max
	outMin, _, _ := r.Minimum()
	if bytes.Compare(outMin, min) != 0 {
		t.Fatalf("bad minimum: %v %v", outMin, min)
	}

	outMax, _, _ := r.Maximum()
	if bytes.Compare(outMax, max) != 0 {
		t.Fatalf("bad maximum: %v %v", outMax, max)
	}

	for k, v := range inp {
		bk := []byte(k)
		out, ok := r.Delete(bk)
		if !ok {
			t.Fatalf("missing key: %v", k)
		}
		if out != v {
			t.Fatalf("value mis-match: %v %v", out, v)
		}
	}
	if r.Len() != 0 {
		t.Fatalf("bad length: %v", r.Len())
	}
}

func TestRoot(t *testing.T) {
	r := New()
	_, ok := r.Delete([]byte{})
	if ok {
		t.Fatalf("bad")
	}

	_, ok = r.Insert([]byte{}, true)
	if ok {
		t.Fatalf("bad")
	}
	val, ok := r.Get([]byte{})
	if !ok || val != true {
		t.Fatalf("bad: %v", val)
	}

	val, ok = r.Delete([]byte{})
	if !ok || val != true {
		t.Fatalf("bad: %v", val)
	}
}

func TestDelete(t *testing.T) {

	r := New()

	s := []string{"", "A", "AB"}

	for _, ss := range s {
		bss := []byte(ss)
		r.Insert(bss, true)
	}

	for _, ss := range s {
		bss := []byte(ss)
		_, ok := r.Delete(bss)
		if !ok {
			t.Fatalf("bad %q", ss)
		}
	}
}

func TestDeletePrefix(t *testing.T) {
	type exp struct {
		inp        []string
		prefix     string
		out        []string
		numDeleted int
	}

	cases := []exp{
		{[]string{"", "A", "AB", "ABC", "R", "S"}, "A", []string{"", "R", "S"}, 3},
		{[]string{"", "A", "AB", "ABC", "R", "S"}, "ABC", []string{"", "A", "AB", "R", "S"}, 1},
		{[]string{"", "A", "AB", "ABC", "R", "S"}, "", []string{}, 6},
		{[]string{"", "A", "AB", "ABC", "R", "S"}, "S", []string{"", "A", "AB", "ABC", "R"}, 1},
		{[]string{"", "A", "AB", "ABC", "R", "S"}, "SS", []string{"", "A", "AB", "ABC", "R", "S"}, 0},
	}

	for _, test := range cases {
		r := New()
		for _, ss := range test.inp {
			bss := []byte(ss)
			r.Insert(bss, true)
		}

		deleted := r.DeletePrefix([]byte(test.prefix))
		if deleted != test.numDeleted {
			t.Fatalf("Bad delete, expected %v to be deleted but got %v", test.numDeleted, deleted)
		}

		out := []byte{}
		fn := func(s []byte, v interface{}) bool {
			out = append(out, s...)
			return false
		}
		r.Walk(fn)

		cv := strings.Join(test.out, "")
		if bytes.Compare(out, []byte(cv)) != 0 {
			t.Fatalf("mis-match: %v <> %v", out, test.out)
		}
	}
}

func TestLongestPrefix(t *testing.T) {
	r := New()

	keys := []string{
		"",
		"foo",
		"foobar",
		"foobarbaz",
		"foobarbazzip",
		"foozip",
	}
	for _, k := range keys {
		bk := []byte(k)
		r.Insert(bk, nil)
	}
	if r.Len() != len(keys) {
		t.Fatalf("bad len: %v %v", r.Len(), len(keys))
	}

	type exp struct {
		inp string
		out string
	}
	cases := []exp{
		{"a", ""},
		{"abc", ""},
		{"fo", ""},
		{"foo", "foo"},
		{"foob", "foo"},
		{"foobar", "foobar"},
		{"foobarba", "foobar"},
		{"foobarbaz", "foobarbaz"},
		{"foobarbazzi", "foobarbaz"},
		{"foobarbazzip", "foobarbazzip"},
		{"foozi", "foo"},
		{"foozip", "foozip"},
		{"foozipzap", "foozip"},
	}
	for _, test := range cases {
		btestinp := []byte(test.inp)
		m, _, ok := r.LongestPrefix(btestinp)
		if !ok {
			t.Fatalf("no match: %v", test)
		}
		if string(m) != test.out {
			t.Fatalf("mis-match: %v %v", m, test)
		}
	}
}

func TestWalkPrefix(t *testing.T) {
	r := New()

	keys := []string{
		"foobar",
		"foo/bar/baz",
		"foo/baz/bar",
		"foo/zip/zap",
		"zipzap",
	}
	for _, k := range keys {
		bk := []byte(k)
		r.Insert(bk, nil)
	}
	if r.Len() != len(keys) {
		t.Fatalf("bad len: %v %v", r.Len(), len(keys))
	}

	type exp struct {
		inp string
		out []string
	}
	cases := []exp{
		{
			"f",
			[]string{"foobar", "foo/bar/baz", "foo/baz/bar", "foo/zip/zap"},
		},
		{
			"foo",
			[]string{"foobar", "foo/bar/baz", "foo/baz/bar", "foo/zip/zap"},
		},
		{
			"foob",
			[]string{"foobar"},
		},
		{
			"foo/",
			[]string{"foo/bar/baz", "foo/baz/bar", "foo/zip/zap"},
		},
		{
			"foo/b",
			[]string{"foo/bar/baz", "foo/baz/bar"},
		},
		{
			"foo/ba",
			[]string{"foo/bar/baz", "foo/baz/bar"},
		},
		{
			"foo/bar",
			[]string{"foo/bar/baz"},
		},
		{
			"foo/bar/baz",
			[]string{"foo/bar/baz"},
		},
		{
			"foo/bar/bazoo",
			[]string{},
		},
		{
			"z",
			[]string{"zipzap"},
		},
	}

	for _, test := range cases {
		out := []string{}
		fn := func(s []byte, v interface{}) bool {
			out = append(out, string(s))
			return false
		}

		btestinp := []byte(test.inp)
		r.WalkPrefix(btestinp, fn)

		sort.Strings(out)
		sort.Strings(test.out)

		if !reflect.DeepEqual(out, test.out) {
			t.Fatalf("mis-match: %v %v", out, test.out)
		}
	}
}

func TestWalkPath(t *testing.T) {
	r := New()

	keys := []string{
		"foo",
		"foo/bar",
		"foo/bar/baz",
		"foo/baz/bar",
		"foo/zip/zap",
		"zipzap",
	}
	for _, k := range keys {
		bk := []byte(k)
		r.Insert(bk, nil)
	}

	if r.Len() != len(keys) {
		t.Fatalf("bad len: %v %v", r.Len(), len(keys))
	}

	type exp struct {
		inp string
		out []string
	}
	cases := []exp{
		{
			"f",
			[]string{},
		},
		{
			"foo",
			[]string{"foo"},
		},
		{
			"foo/",
			[]string{"foo"},
		},
		{
			"foo/ba",
			[]string{"foo"},
		},
		{
			"foo/bar",
			[]string{"foo", "foo/bar"},
		},
		{
			"foo/bar/baz",
			[]string{"foo", "foo/bar", "foo/bar/baz"},
		},
		{
			"foo/bar/bazoo",
			[]string{"foo", "foo/bar", "foo/bar/baz"},
		},
		{
			"z",
			[]string{},
		},
	}

	for _, test := range cases {
		out := []string{}
		fn := func(s []byte, v interface{}) bool {
			out = append(out, string(s))
			return false
		}

		btestinp := []byte(test.inp)
		r.WalkPath(btestinp, fn)

		sort.Strings(out)
		sort.Strings(test.out)
		if !reflect.DeepEqual(out, test.out) {
			t.Fatalf("mis-match: %v %v", out, test.out)
		}
	}
}

func BenchmarkInsert(b *testing.B) {
	r := New()

	for i := 0; i < 10000; i++ {
		s := fmt.Sprintf("init%d", i)
		bs := []byte(s)
		r.Insert(bs, true)
	}

	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		s := strconv.Itoa(n)
		bs := []byte(s)
		_, updated := r.Insert(bs, true)
		if updated {
			b.Fatal("bad")
		}
	}
}

// generateUUID is used to generate a random UUID
func generateUUID() string {
	buf := make([]byte, 16)
	if _, err := crand.Read(buf); err != nil {
		panic(fmt.Errorf("failed to read random bytes: %v", err))
	}

	return fmt.Sprintf("%08x-%04x-%04x-%04x-%12x",
		buf[0:4],
		buf[4:6],
		buf[6:8],
		buf[8:10],
		buf[10:16])
}
