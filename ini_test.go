package ini_test

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strings"
	"testing"
	"testing/iotest"
	"time"

	ini "github.com/pierrec/go-ini"
)

func TestNew(t *testing.T) {
	conf, err := ini.New()
	if err != nil {
		t.Fatal(err)
	}

	// Check there is no section and key.
	if n := len(conf.Sections()); n != 0 {
		t.Fatalf("expected no section")
	}

	if n := len(conf.Keys("")); n != 0 {
		t.Fatalf("expected no keys in the global section")
	}

	// An option returning an error should be returned by the constructor.
	_, err = ini.New(
		func(*ini.INI) error {
			return errors.New("this is a test error")
		},
	)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestGlobal(t *testing.T) {
	conf, _ := ini.New()

	// Add a new key.
	conf.Set("", "k1", "v1")

	// The new section should be added.
	if n := len(conf.Sections()); n != 0 {
		t.Fatalf("expected no section")
	}

	// The new key should be added.
	if got, want := conf.Keys(""), []string{"k1"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v; want %v", got, want)
	}

	// The key value should be added.
	if got, want := conf.Get("", "k1"), "v1"; got != want {
		t.Fatalf("got %v; want %v", got, want)
	}

	// Overwrite the key value.
	conf.Set("", "k1", "v1.1")
	if got, want := conf.Get("", "k1"), "v1.1"; got != want {
		t.Fatalf("got %v; want %v", got, want)
	}

	// Missing key.
	if got, want := conf.Get("", "k2"), ""; got != want {
		t.Fatalf("got %v; want %v", got, want)
	}

}

func TestSetGet(t *testing.T) {
	conf, _ := ini.New()

	// Add a new section and key.
	conf.Set("sec1", "k1", "v1")

	// The new section should be added.
	if got, want := conf.Sections(), []string{"sec1"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v; want %v", got, want)
	}

	// The new key should be added.
	if got, want := conf.Keys("sec1"), []string{"k1"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v; want %v", got, want)
	}

	// Add a newline to separate the keys.
	conf.Set("sec1", "", "")
	if got, want := conf.Keys("sec1"), []string{"k1", ""}; !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v; want %v", got, want)
	}

	// Consecutive newlines are not allowed.
	conf.Set("sec1", "", "")
	if got, want := conf.Keys("sec1"), []string{"k1", ""}; !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v; want %v", got, want)
	}

	// The key value should be added.
	if got, want := conf.Get("sec1", "k1"), "v1"; got != want {
		t.Fatalf("got %v; want %v", got, want)
	}

	// Overwrite the key value.
	conf.Set("sec1", "k1", "v1.1")
	if got, want := conf.Get("sec1", "k1"), "v1.1"; got != want {
		t.Fatalf("got %v; want %v", got, want)
	}

	// Missing key.
	if got, want := conf.Get("sec1", "k2"), ""; got != want {
		t.Fatalf("got %v; want %v", got, want)
	}

	// Missing section.
	if got, want := conf.Get("sec2", "k1"), ""; got != want {
		t.Fatalf("got %v; want %v", got, want)
	}

	// Add a new empty section.
	conf.Set("sec2", "", "")

	// The new section should be added.
	if got, want := conf.Sections(), []string{"sec1", "sec2"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v; want %v", got, want)
	}
}

func TestDel(t *testing.T) {
	conf, _ := ini.New()

	conf.Set("", "k1", "v1")
	conf.Set("sec1", "k1", "v1")
	conf.Set("sec1", "", "")
	conf.Set("sec1", "k2", "v2")
	conf.Set("sec2", "k1", "v1")

	// Global section: remove missing and existing keys
	if conf.Del("", "k") {
		t.Error("should not be true")
	}
	if !conf.Del("", "k1") {
		t.Error("should be true")
	}

	// Section: remove missing and existing keys
	if conf.Del("secX", "k") {
		t.Error("should not be true")
	}
	if conf.Del("sec2", "k") {
		t.Error("should not be true")
	}
	if !conf.Del("sec2", "k1") {
		t.Error("should be true")
	}
	if !conf.Del("sec1", "k2") {
		t.Error("should be true")
	}

	// Remove sections.
	if conf.Del("secX", "") {
		t.Error("should not be true")
	}
	if !conf.Del("sec1", "") {
		t.Error("should be true")
	}
	if !conf.Del("sec2", "") {
		t.Error("should be true")
	}
	if !conf.Del("", "") {
		t.Error("should be true")
	}
}

func TestDelWithEmptyBlocks(t *testing.T) {
	conf, _ := ini.New()

	conf.Set("sec", "k1", "v1")
	conf.Set("sec", "", "")
	conf.Set("sec", "k2", "v2")
	conf.Set("sec", "", "")
	conf.Set("sec", "k3", "v3")
	conf.Set("sec", "", "")
	conf.Set("sec", "k4", "v4")

	conf.Del("sec", "k1")
	if got, want := conf.Keys("sec"), []string{"k2", "", "k3", "", "k4"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v; want %v", got, want)
	}

	conf.Del("sec", "k3")
	if got, want := conf.Keys("sec"), []string{"k2", "", "k4"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v; want %v", got, want)
	}
}

func TestOptionCaseSensitive(t *testing.T) {
	conf, _ := ini.New(ini.CaseSensitive())

	conf.Set("sec1", "k1", "v1")

	// Match on same section and key.
	if got, want := conf.Get("sec1", "k1"), "v1"; got != want {
		t.Fatalf("got %v; want %v", got, want)
	}

	// Match on upper case section.
	if got, want := conf.Get("Sec1", "k1"), ""; got != want {
		t.Fatalf("got %v; want %v", got, want)
	}

	// Match on upper case key.
	if got, want := conf.Get("sec1", "K1"), ""; got != want {
		t.Fatalf("got %v; want %v", got, want)
	}
}

// Only a pointer to a structure can be encoded into.
func TestInvalidEncode(t *testing.T) {
	conf, _ := ini.New()

	if err := conf.Encode(1); err == nil {
		t.Fatal("expected error")
	}

	if err := conf.Encode(struct{}{}); err == nil {
		t.Fatal("expected error")
	}

	var i int
	if err := conf.Encode(&i); err == nil {
		t.Fatal("expected error")
	}
}

// Parse errors when parsing the ini source.
func TestInvalidIni(t *testing.T) {
	conf, _ := ini.New()

	buf := bytes.NewBuffer(nil)
	for _, data := range []string{
		"[sectionA",
		"[]",
		"key",
		"key='xyz",
	} {
		buf.Reset()
		buf.WriteString(data)

		if _, err := conf.ReadFrom(buf); err == nil {
			t.Fatalf("expected error when parsing %v", data)
		}
	}
}

func TestInvalidDecode(t *testing.T) {
	conf, _ := ini.New()

	// Only a pointer to a structure can be decoded from.
	if err := conf.Decode(1); err == nil {
		t.Fatal("expected error")
	}

	if err := conf.Decode(struct{}{}); err == nil {
		t.Fatal("expected error")
	}

	var i int
	if err := conf.Decode(&i); err == nil {
		t.Fatal("expected error")
	}

	// Invalid values for the expected types.
	type config struct {
		A int           `ini:"idx"`
		B bool          `ini:"flag"`
		C time.Duration `ini:"dur"`
		D time.Time     `ini:"date"`
		E uint32        `ini:"hash"`
		F float64       `ini:"v"`
		S []int         `ini:"lst"`
	}

	buf := bytes.NewBuffer(nil)
	for _, data := range []string{
		"idx=xyz",
		"flag=xyz",
		"dur=xyz",
		"date=xyz",
		"hash=xyz",
		"v=xyz",
		"lst=a,b",
	} {
		buf.Reset()
		buf.WriteString(data)

		var conf config
		if err := ini.Decode(buf, &conf); err == nil {
			t.Fatalf("expected error when parsing %v", data)
		}
	}
}

func TestEncode(t *testing.T) {
	type config struct {
		dummy int
		io.Reader
		A int           `ini:"idx,sec1"`
		B string        `ini:"str,sec1"`
		C bool          `ini:"flag,sec2"`
		D time.Duration `ini:"dur,sec2"`
		E time.Time     `ini:"date,sec2"`
		F uint32        `ini:"hash,sec3"`
		G float64       `ini:"v,sec3"`
	}

	buf := new(bytes.Buffer)
	date, _ := time.Parse("2006-Jan-02", "2013-Feb-03")
	conf := &config{1, nil, 123, "abc", true, time.Second, date, 0xC4F3, 1.234}

	if err := ini.Encode(buf, conf); err != nil {
		t.Fatal(err)
	}
}

func TestDecode(t *testing.T) {
	type config struct {
		dummy int
		io.Reader
		A  int           `ini:"idx,sec1"`
		B  string        `ini:"str,sec1"`
		B1 string        `ini:"str2,sec1"`
		B2 string        `ini:"str3,sec1"`
		B3 string        `ini:"str4,sec1"`
		C  bool          `ini:"flag,sec2"`
		D  time.Duration `ini:"dur,sec2"`
		E  time.Time     `ini:"date,sec2"`
		F  uint32        `ini:"hash,sec3"`
		G  float64       `ini:"v,sec3"`
		S  []int         `ini:"lst,sec3"`
	}

	data := `
[sec1]
idx=123
str="a\"b\"c"
str2="a\"bc"
str3=
str4="abc"
[sec2]
flag=true
dur=1s
date=2006-01-02T15:04:05Z
[sec3]
hash=0xC4F3
v=1.234
lst=1,2,3
`

	for _, ending := range []string{"\n", "\r\n"} {
		ndata := strings.Replace(data, "\n", ending, -1)
		buf := bytes.NewBufferString(ndata)
		var conf config

		if err := ini.Decode(buf, &conf); err != nil {
			t.Fatal(err)
		}

		date, _ := time.Parse(time.RFC3339, "2006-01-02T15:04:05Z")
		got, want := conf, config{0, nil, 123, "a\"b\"c", "a\"bc", "", "abc", true, time.Second, date, 0xC4F3, 1.234, []int{1, 2, 3}}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("got %v; want %v", got, want)
		}
	}
}

func TestDecodeEncode(t *testing.T) {
	data := `
Gk1 = gv1
Gk2 = gv2

Global3 = gv3

[sectionA]
key1 = abc
k2   = xyz

[sectionA]
key1 = v1.1
k2   = v1.2

[sectionB]
k1 = abc
k1 = v2.1
k2 = xyz

k2 = v2.2
k3 = v2.3
`
	want := `Gk1 = gv1
Gk2 = gv2

Global3 = gv3

[sectionA]
key1 = v1.1
k2   = v1.2

[sectionB]
k1 = v2.1

k2 = v2.2
k3 = v2.3
`

	buf := bytes.NewBufferString(data)

	type config struct {
		Gk1     string
		Gk2     string `ini:",,true"`
		Global3 string

		A1 string `ini:"key1,sectionA"`
		A2 string `ini:"k2,sectionA"`

		B1 string `ini:"k1,sectionB,true"`
		B2 string `ini:"k2,sectionB"`
		B3 string `ini:"k3,sectionB"`
	}

	var conf config

	if err := ini.Decode(buf, &conf); err != nil {
		t.Fatal(err)
	}

	output := bytes.NewBuffer(nil)
	if err := ini.Encode(output, &conf); err != nil {
		t.Fatal(err)
	}

	if got := string(output.Bytes()); !reflect.DeepEqual(got, want) {
		t.Fatalf("got '%v'; want '%v'", got, want)
	}
}

type Tuser struct {
	P password `ini:"pwd"`
}

type password string

func (p *password) MarshalText() ([]byte, error) {
	if *p == "doerror" {
		return nil, errors.New("fake error")
	}
	s := fmt.Sprintf("__%s__", *p)
	return []byte(s), nil
}

func (p *password) UnmarshalText(buf []byte) error {
	n := len(buf)
	if n < 4 || string(buf[:2]) != "__" || string(buf[n-2:]) != "__" {
		return errors.New("invalid input")
	}
	*p = password(buf[2 : n-2])
	return nil
}

func TestTexter(t *testing.T) {
	// The MarshalText interface should be applied.
	// Even to embedded structs.
	type config struct {
		Tuser
	}

	conf := config{Tuser{"secret"}}
	buf := bytes.NewBuffer(nil)

	// The password should be encoded using MarshalText.
	if err := ini.Encode(buf, &conf); err != nil {
		t.Fatal(err)
	}

	want := `pwd = __secret__
`
	if got := string(buf.Bytes()); got != want {
		t.Fatalf("got '%v'; want '%v'", got, want)
	}

	// The password should be decoded using UnmarshalText.
	conf.P = ""
	if err := ini.Decode(buf, &conf); err != nil {
		t.Fatal(err)
	}

	if got, want := string(conf.P), "secret"; got != want {
		t.Fatalf("got '%v'; want '%v'", got, want)
	}

	// Texter error: the encoded password is invalid.
	buf.Reset()
	buf.WriteString("pwd = secret")
	if err := ini.Decode(buf, &conf); err == nil {
		t.Fatal("expected error")
	}

	conf.P = "doerror"
	if err := ini.Encode(buf, &conf); err == nil {
		t.Fatal("expected error")
	}
}

func TestOverwritingSections(t *testing.T) {
	data := `
[sectionA]
key1 = abc
key2 = xyz

[sectionA]
key1 = v1.1
k2   = v1.2
`
	want := `[sectionA]
key1 = v1.1
k2   = v1.2
`

	buf := bytes.NewBufferString(data)
	output := bytes.NewBuffer(nil)
	conf, _ := ini.New()

	if _, err := conf.ReadFrom(buf); err != nil {
		t.Fatal(err)
	}

	if _, err := conf.WriteTo(output); err != nil {
		t.Fatal(err)
	}

	if got := string(output.Bytes()); !reflect.DeepEqual(got, want) {
		t.Fatalf("got '%v'; want '%v'", got, want)
	}
}

func TestMergingSections(t *testing.T) {
	data := `
[sectionA]
key1 = abc
key2 = xyz

[sectionA]
key1 = v1.1
k2   = v1.2
`
	want := `[sectionA]
key2 = xyz

key1 = v1.1
k2   = v1.2
`

	buf := bytes.NewBufferString(data)
	output := bytes.NewBuffer(nil)
	conf, _ := ini.New(ini.MergeSections())

	if _, err := conf.ReadFrom(buf); err != nil {
		t.Fatal(err)
	}

	if _, err := conf.WriteTo(output); err != nil {
		t.Fatal(err)
	}

	if got := string(output.Bytes()); !reflect.DeepEqual(got, want) {
		t.Fatalf("got '%v'; want '%v'", got, want)
	}
}

func TestDefaultOptions(t *testing.T) {
	type config struct {
		AS int   `ini:"a,S"`
		BS int   `ini:"b,S"`
		As int   `ini:"A,s"`
		Bs int   `ini:"B,s"`
		S  []int `ini:"lst"`
	}

	data := `lst = 1_2_3

# section S
[S]
# key A
a = 1

# section s
[s]
# key B
B = 2

[S]
b = 2

[s]
A = 1
`

	want := `lst = 1_2_3

[S]
a = 1
b = 2

[s]
A = 1
B = 2
`

	var conf config

	ini.DefaultOptions = []ini.Option{
		ini.CaseSensitive(),
		ini.Comment('#'),
		ini.MergeSections(),
		ini.SliceSeparator("_"),
	}
	defer func() { ini.DefaultOptions = nil }()

	buf := bytes.NewBufferString(data)
	if err := ini.Decode(buf, &conf); err != nil {
		t.Error(err)
	}

	output := bytes.NewBuffer(nil)
	if err := ini.Encode(output, &conf); err != nil {
		t.Error(err)
	}

	if got := string(output.Bytes()); !reflect.DeepEqual(got, want) {
		t.Errorf("got '%v'; want '%v'", got, want)
	}

	// Error option handling.
	ini.DefaultOptions = []ini.Option{
		func(*ini.INI) error { return errors.New("option error") },
	}

	buf.WriteString(data)
	if err := ini.Decode(buf, &conf); err == nil {
		t.Error("expected error")
	}

	output.Reset()
	if err := ini.Encode(output, &conf); err == nil {
		t.Error("expected error")
	}
}

func TestFormatting(t *testing.T) {
	data := `
; Global section comment1
; Global section comment2

Gk1 = gv1

; sectionA comment1
; sectionA comment2
[sectionA]
; A.k1 comment
k1   = xyz



[sectionAA]
  myKey   =  myValue
myKeyBis   =  myValueBis


  mySecondKey   =  myValue
mySecondKeyBis   =  myValueBis


; sectionB comment1
; sectionB comment2
[sectionB]
; B.k1 comment
k1 = abc
`

	want := `; Global section comment1
; Global section comment2
Gk1 = gv1

; sectionA comment1
; sectionA comment2
[sectionA]
; A.k1 comment
k1 = xyz

[sectionAA]
myKey    = myValue
myKeyBis = myValueBis

mySecondKey    = myValue
mySecondKeyBis = myValueBis

; sectionB comment1
; sectionB comment2
[sectionB]
; B.k1 comment
k1 = abc
`

	buf := bytes.NewBufferString(data)
	output := bytes.NewBuffer(nil)
	conf, _ := ini.New()

	if _, err := conf.ReadFrom(buf); err != nil {
		t.Fatal(err)
	}

	if _, err := conf.WriteTo(output); err != nil {
		t.Fatal(err)
	}

	if got := string(output.Bytes()); !reflect.DeepEqual(got, want) {
		t.Fatalf("got '%v'; want '%v'", got, want)
	}
}

func TestSetComments(t *testing.T) {
	data := `
[sectionA]
k1 = xyz
`

	want := `#Global section comment

#sectionA comment
[sectionA]
#A.k1 comment
k1 = xyz
`

	buf := bytes.NewBufferString(data)
	output := bytes.NewBuffer(nil)
	conf, _ := ini.New(ini.Comment('#'))

	if _, err := conf.ReadFrom(buf); err != nil {
		t.Fatal(err)
	}

	conf.SetComments("", "", "Global section comment")
	conf.SetComments("sectionA", "", "sectionA comment")
	conf.SetComments("sectionA", "k1", "A.k1 comment")
	conf.SetComments("sectionA", "k", "missing key comment")

	if _, err := conf.WriteTo(output); err != nil {
		t.Fatal(err)
	}

	if got := string(output.Bytes()); !reflect.DeepEqual(got, want) {
		t.Fatalf("got '%v'; want '%v'", got, want)
	}

	if got, want := conf.GetComments("", ""), []string{"Global section comment"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("got '%v'; want '%v'", got, want)
	}

	if got, want := conf.GetComments("sectionA", ""), []string{"sectionA comment"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("got '%v'; want '%v'", got, want)
	}

	if got, want := conf.GetComments("sectionA", "k1"), []string{"A.k1 comment"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("got '%v'; want '%v'", got, want)
	}

	if got, want := conf.GetComments("sectionA", "k"), []string(nil); !reflect.DeepEqual(got, want) {
		t.Fatalf("got '%v'; want '%v'", got, want)
	}

	if got, want := conf.GetComments("sectionX", "k"), []string(nil); !reflect.DeepEqual(got, want) {
		t.Fatalf("got '%v'; want '%v'", got, want)
	}
}

func TestFaultyReader(t *testing.T) {
	r := iotest.TimeoutReader(bytes.NewBufferString("a=1"))

	if err := ini.Decode(r, &struct{}{}); err == nil {
		t.Fatal("expected error")
	}
}

var faultyWriterError = errors.New("faulty")

type faultyWriter struct {
	n int64
}

func (w *faultyWriter) Write(buf []byte) (int, error) {
	if w.n == 0 {
		return 0, faultyWriterError
	}
	n := int64(len(buf))
	if n < w.n {
		w.n -= n
		return int(n), nil
	}
	n, w.n = w.n, 0
	return int(n), faultyWriterError
}

// Makes tests coverage 100%...
// ...but this is painful to maintain.
func TestFaultyWriter(t *testing.T) {
	conf, _ := ini.New()

	data := `;

[s]
;
k = v

kk = vv
`
	_ = data
	conf.SetComments("", "", "")
	conf.Set("s", "k", "v")
	conf.Set("s", "", "")
	conf.SetComments("s", "k", "")
	conf.Set("s", "kk", "vv")

	for _, size := range []int64{
		// Global section comment fails.
		1,
		// Newline between sections fails.
		3,
		// Section name fails.
		4,
		// Key comment fails.
		8,
		// Key name fails.
		10,
		// Newline between keys fails.
		16,
	} {
		w := &faultyWriter{size}

		n, err := conf.WriteTo(w)
		if err == nil {
			t.Fatal("expected error")
		}
		if got, want := n, size; got != want {
			t.Fatalf("got '%v'; want '%v'", got, want)
		}
	}
}

func TestReset(t *testing.T) {
	conf, _ := ini.New(ini.CaseSensitive(), ini.Comment('#'))

	conf.Set("", "key1", "value1")
	conf.Set("sectionA", "key1", "value1")
	conf.Reset()

	if n := len(conf.Sections()); n != 0 {
		t.Fatalf("expected no section")
	}

	if n := len(conf.Keys("")); n != 0 {
		t.Fatalf("expected no keys in the global section")
	}
}
