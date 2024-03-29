package stringutil

import (
	"strings"
	"testing"

	assert1 "github.com/stretchr/testify/assert"
)

func TestLastRune(t *testing.T) {
	var actual, def rune
	if actual = LastRune("s"); actual != 's' {
		t.Errorf("%c != 's'", actual)
	}
	if actual = LastRune(""); actual != def {
		t.Errorf("%c != '%c'", actual, def)
	}
	if actual = LastRune("test"); actual != 't' {
		t.Errorf("%c != '%c'", actual, def)
	}
	if actual = LastRune("Лайка"); actual != 'а' {
		t.Errorf("%c != 'а'", actual)
	}
}

func TestRuneAtIndex(t *testing.T) {
	var actual, def rune
	if actual = RuneAtIndex("", 10); actual != def {
		t.Errorf("Default rune should be returned when index exceeds length.")
	}
	if actual = RuneAtIndex("asdf", 10); actual != def {
		t.Errorf("Default rune should be returned when index exceeds length.")
	}
	if actual = RuneAtIndex("asdf", -10); actual != def {
		t.Errorf("Default rune should be returned when negative index exceeds length.")
	}
	if actual = RuneAtIndex("s", 0); actual != 's' {
		t.Errorf("%c != 's'", actual)
	}
	if actual = RuneAtIndex("foobar", 4); actual != 'a' {
		t.Errorf("%c != 'a'", actual)
	}
	if actual = RuneAtIndex("Лайка", 2); actual != 'й' {
		t.Errorf("%c != 'й'", actual)
	}
	if actual = RuneAtIndex("Лайка", -2); actual != 'к' {
		t.Errorf("%c != 'к'", actual)
	}
}

func TestSafeSubstring(t *testing.T) {
	var expected string
	if expected = SafeSubstring("?OUTPUT,8,1,76.16", 1, -6); expected != "OUTPUT,8,1" {
		t.Errorf("%s != 'OUTPUT,8,1'", expected)
	}
	if expected = SafeSubstring("", 0, 10); expected != "" {
		t.Errorf("%s != ''", expected)
	}
	if expected = SafeSubstring("short.string", 6, 100); expected != "string" {
		t.Errorf("%s != 'string'", expected)
	}
	if expected = SafeSubstring("Лайка", 0, 3); expected != "Лай" {
		t.Errorf("%s != 'Лай'", expected)
	}
}

func TestReverse(t *testing.T) {
	var actual string
	if actual = Reverse(""); actual != "" {
		t.Errorf("%s != ''", actual)
	}
	if actual = Reverse("asdf"); actual != "fdsa" {
		t.Errorf("%s != 'fdsa'", actual)
	}
}

func TestIsEmpty(t *testing.T) {
	if IsEmpty("s") {
		t.Errorf("Should not be empty.")
	}

	if !IsEmpty("") {
		t.Errorf("Should be empty.")
	}
}

func TestIsWhiteSpace(t *testing.T) {
	if !IsWhiteSpace("") {
		t.Errorf("Should be whitespace")
	}

	if !IsWhiteSpace("   ") {
		t.Errorf("Should be whitespace")
	}

	if !IsWhiteSpace(" \n \t ") {
		t.Errorf("Should be whitespace")
	}

	if IsWhiteSpace("asdf") {
		t.Errorf("Should not be whitespace")
	}
}

var nullTerminatedStringTests = []struct {
	expected string
	bytes    []byte
}{
	{"", []byte{0, 0, 0, 0}},
	{"A", []byte{65, 0, 0, 0, 0}},
	{"AB", []byte{65, 66, 0, 0, 0}},
	{"AB", []byte{65, 66, 0, 75, 0}},
	{"", []byte{0, 66, 0, 75, 0}},
}

func TestNullTeminatedString(t *testing.T) {
	for _, te := range nullTerminatedStringTests {
		actual := NullTerminatedString(te.bytes)
		if actual != te.expected {
			t.Errorf("NullTerminatedString(%v) = '%s', Expected '%s'",
				te.bytes, actual, te.expected)
		}
	}
}

func TestClean(t *testing.T) {
	input := []string{"", " ", "foo"}
	expected := input[1:]

	assert1.Equal(t, expected, Clean(input))
}

var appendIfMissingData = []struct {
	initial  []string
	toAdd    string
	expected []string
}{
	{[]string{}, "A", []string{"A"}},
	{[]string{"A"}, "A", []string{"A"}},
	{[]string{"B"}, "A", []string{"B", "A"}},
}

func TestAppendIfMissing(t *testing.T) {
	assert := assert1.New(t)
	for _, te := range appendIfMissingData {
		actual := AppendIfMissing(te.initial, te.toAdd)
		assert.Equal(te.expected, actual)
	}
}

func TestSprintHex(t *testing.T) {
	assert := assert1.New(t)

	sprintHexStringTests := []struct {
		expected string
		bytes    []byte
	}{
		{"00 0F 02", []byte{0, 15, 2}},
		{"FF", []byte{255}},
		{"", []byte{}},
		{"54 65 73 74", []byte("Test")},
		{"BD B2 3D BC 20 E2 8C 98", []byte("\xbd\xb2\x3d\xbc\x20\xe2\x8c\x98")},
	}

	for _, t := range sprintHexStringTests {
		hexString := SprintHex(t.bytes)
		assert.Equal(t.expected, hexString)
	}
}

func TestByteToASCIIHexValue(t *testing.T) {
	assert := assert1.New(t)

	byteToASCIIHexTests := []struct {
		expected []byte
		bytes    []byte
	}{
		{[]byte("0123456789ABCDEF"), []byte{0x01, 0x23, 0x45, 0x67, 0x89, 0xAB, 0xCD, 0xEF}},
		// Next test looks at conversion of a Serial Command in ASCII characters to ASCII HEX encoding in byte array
		// to ASCII HEX representation for each HEX char in encoded byte array. Input of byte array of "VOLUME +"
		// is equivalent to ASCII HEX encoding of []byte{0x56, 0x4F, 0x4C, 0x55, 0x4D, 0x45, 0x20, 0x2B}.
		{[]byte("564F4C554D45202B"), []byte("VOLUME +")},
		{},
	}

	for _, t := range byteToASCIIHexTests {
		hexString := ByteToHexASCII(t.bytes)
		assert.Equal(t.expected, hexString)
	}
}

func TestMakeASCIIZeros(t *testing.T) {
	assert := assert1.New(t)

	makeASCIIZerosTests := []struct {
		expected []byte
		size     uint
	}{
		{[]byte("00"), 2},
		{[]byte(""), 0},
	}

	for _, t := range makeASCIIZerosTests {
		zerosASCIIByteArray := MakeASCIIZeros(t.size)
		assert.Equal(t.expected, zerosASCIIByteArray)
	}
}

func TestConstantTimeComparison(t *testing.T) {
	assert := assert1.New(t)
	var testData = []struct {
		Expected string
		Actual   string
	}{
		{"", "a"},
		{"a", "a"},
		{"a", "b"},
		{"a", "ba"},
		{"abc", "ba"},
	}
	for _, data := range testData {
		assert.Equal(data.Expected == data.Actual, ConstantTimeComparison(data.Expected, data.Actual))
	}
}
func TestSafeSubstringIndexing(t *testing.T) {
	type args struct {
		s     string
		start int
		end   int
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "safe",
			args: args{
				s:     "Лайка",
				start: 0,
				end:   2,
			},
			want: "Ла",
		},
		{
			name: "simple",
			args: args{
				s:     "asdf",
				start: 0,
				end:   2,
			},
			want: "as",
		},
		{
			name: "negative end",
			args: args{
				s:     "asdf",
				start: 0,
				end:   -1,
			},
			want: "asd",
		},
		{
			name: "negative start",
			args: args{
				s:     "asdf",
				start: -3,
				end:   3,
			},
			want: "sd",
		},
		{
			name: "start negative end negative ",
			args: args{
				s:     "asdf",
				start: -3,
				end:   -2,
			},
			want: "s",
		},
		{
			name: "start negative end negative but weird",
			args: args{
				s:     "asdf",
				start: -2,
				end:   -3,
			},
			want: "s",
		},
		{
			name: "zero end",
			args: args{
				s:     "asdf",
				start: 1,
				end:   0,
			},
			want: "sdf",
		},
		{
			name: "silly",
			args: args{
				s:     "asdf",
				start: -50,
				end:   50,
			},
			want: "asdf",
		},
		{
			name: "sillier",
			args: args{
				s:     "asdf",
				start: 50,
				end:   -50,
			},
			want: "asdf",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SafeSubstring(tt.args.s, tt.args.start, tt.args.end); got != tt.want {
				t.Errorf("SafeSubstring(\"%s\", %d, %d) = \"%v\", want \"%v\"",
					tt.args.s, tt.args.start, tt.args.end, got, tt.want)
			}
		})
	}
}

func TestPSPrint(t *testing.T) {
	type args struct {
		prefix string
		m      map[string]string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "empty",
			args: args{
				prefix: "foo",
				m:      make(map[string]string),
			},
			want: []string{""},
		},
		{
			name: "basic",
			args: args{
				prefix: "foo",
				m:      map[string]string{"🤡": "bar", "ipsum": "lorem"},
			},
			want: []string{"foo'🤡': 'bar'", "foo'ipsum': 'lorem'"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := strings.Split(PSPrint(tt.args.prefix, tt.args.m), "\n")
			assert1.ElementsMatch(t, tt.want, got, "want: %#v, got %#v", tt.want, got)
		})
	}
}

func TestPointer(t *testing.T) {
	type args struct {
		str string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "simple",
			args: args{str: "bar🤡"},
			want: "bar🤡",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Pointer(tt.args.str); *got != tt.want {
				t.Errorf("Pointer() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNumericOnly(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want string
	}{
		{
			"no numeric",
			"the quick brown fox jumped over the lazy dog",
			"",
		},
		{
			"all numeric",
			"0987654321234567890",
			"0987654321234567890",
		},
		{
			"mixed",
			"0a9b8c7d6e5f4g3h21sdfg2rt3er 4t5*(&^@6)(*&+_7asd8eqr9q0/';",
			"0987654321234567890",
		},
		{
			"empty",
			"",
			"",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NumericOnly(tt.s); got != tt.want {
				t.Errorf("NumericOnly(\"%s\") = \"%v\", want \"%v\"", tt.s, got, tt.want)
			}
		})
	}
}

func TestObfuscate(t *testing.T) {
	assert := assert1.New(t)
	type testCase struct {
		Name string

		str        string
		length     int
		direction  Direction
		obfuscator string

		Expected string
	}

	validate := func(t *testing.T, tc *testCase) {
		t.Run(tc.Name, func(t *testing.T) {
			actual := Obfuscate(tc.str, tc.length, tc.direction, tc.obfuscator)

			assert.Equal(tc.Expected, actual, tc.Name)
		})
	}

	validate(t, &testCase{
		Name: "Empty",

		str:        "",
		length:     0,
		direction:  Left,
		obfuscator: "*",

		Expected: "",
	})
	validate(t, &testCase{
		Name: "Obfuscate none",

		str:        "foo bar",
		length:     0,
		direction:  Left,
		obfuscator: "*",

		Expected: "foo bar",
	})
	validate(t, &testCase{
		Name: "Invalid direction",

		str:        "foo bar",
		length:     5,
		direction:  -1,
		obfuscator: "*",

		Expected: "foo bar",
	})
	validate(t, &testCase{
		Name: "Empty obfuscator",

		str:        "foo bar",
		length:     5,
		direction:  Left,
		obfuscator: "",

		Expected: "ar",
	})
	validate(t, &testCase{
		Name: "Long obfuscator",

		str:        "foober",
		length:     3,
		direction:  Left,
		obfuscator: "bar",

		Expected: "barbarbarber",
	})
	validate(t, &testCase{
		Name: "Obfuscation longer than input string",

		str:        "foo",
		length:     5,
		direction:  Left,
		obfuscator: "X",

		Expected: "XXXXX",
	})
	validate(t, &testCase{
		Name: "Obfuscate left",

		str:        "foo bar foo",
		length:     6,
		direction:  Left,
		obfuscator: "*",

		Expected: "******r foo",
	})
	validate(t, &testCase{
		Name: "Obfuscate right",

		str:        "this is redacted",
		length:     8,
		direction:  Right,
		obfuscator: "█",

		Expected: "this is ████████",
	})
	validate(t, &testCase{
		Name: "Obfuscate negative repeat",

		str:        "this isnt redacted",
		length:     -8,
		direction:  Right,
		obfuscator: "█",

		Expected: "this isnt redacted",
	})
}

func TestObfuscateLeft(t *testing.T) {
	assert := assert1.New(t)
	type testCase struct {
		Name string

		str        string
		length     int
		obfuscator string

		Expected string
	}

	validate := func(t *testing.T, tc *testCase) {
		t.Run(tc.Name, func(t *testing.T) {
			actual := ObfuscateLeft(tc.str, tc.length, tc.obfuscator)

			assert.Equal(tc.Expected, actual, tc.Name)
		})
	}

	validate(t, &testCase{
		Name: "Empty",

		str:        "",
		length:     0,
		obfuscator: "*",

		Expected: "",
	})
	validate(t, &testCase{
		Name: "Obfuscation longer than input string",

		str:        "foo",
		length:     5,
		obfuscator: "X",

		Expected: "XXXXX",
	})
	validate(t, &testCase{
		Name: "Obfuscate left",

		str:        "foo bar foo",
		length:     6,
		obfuscator: "*",

		Expected: "******r foo",
	})
	validate(t, &testCase{
		Name: "Obfuscate left v2",

		str:        "this is redacted",
		length:     8,
		obfuscator: "█",

		Expected: "████████redacted",
	})
}

func TestObfuscateRight(t *testing.T) {
	assert := assert1.New(t)
	type testCase struct {
		Name string

		str        string
		length     int
		obfuscator string

		Expected string
	}

	validate := func(t *testing.T, tc *testCase) {
		t.Run(tc.Name, func(t *testing.T) {
			actual := ObfuscateRight(tc.str, tc.length, tc.obfuscator)

			assert.Equal(tc.Expected, actual, tc.Name)
		})
	}

	validate(t, &testCase{
		Name: "Empty",

		str:        "",
		length:     0,
		obfuscator: "*",

		Expected: "",
	})
	validate(t, &testCase{
		Name: "Obfuscation longer than input string",

		str:        "foo",
		length:     5,
		obfuscator: "X",

		Expected: "XXXXX",
	})
	validate(t, &testCase{
		Name: "Obfuscate right",

		str:        "foo bar foo",
		length:     6,
		obfuscator: "*",

		Expected: "foo b******",
	})
	validate(t, &testCase{
		Name: "Obfuscate right v2",

		str:        "this is redacted",
		length:     8,
		obfuscator: "█",

		Expected: "this is ████████",
	})
}

func TestObfuscateRightPercent(t *testing.T) {
	tcs := []struct {
		name       string
		input      string
		percent    int
		expected   string
		obfuscator string
	}{
		{
			name:       "simple 90",
			input:      "simple",
			percent:    90,
			expected:   "s*****",
			obfuscator: "*",
		},
		{
			name:       "simple 50",
			input:      "simple",
			percent:    50,
			expected:   "sim***",
			obfuscator: "*",
		},
		{
			name:       "1 char 90",
			input:      "s",
			percent:    90,
			expected:   "s",
			obfuscator: "*",
		},
		{
			name:       "9 chars 90",
			input:      "123456789",
			percent:    90,
			expected:   "1********",
			obfuscator: "*",
		},
		{
			name:       "zero",
			input:      "123456789",
			percent:    0,
			expected:   "123456789",
			obfuscator: "*",
		},
		{
			name:       "all",
			input:      "123456789",
			percent:    100,
			expected:   "1********",
			obfuscator: "*",
		},
		{
			name:       "long",
			input:      "12345678901234567890",
			percent:    90,
			expected:   "12------------------",
			obfuscator: "-",
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			assert1.Equal(t, tc.expected, ObfuscateRightPercent(tc.input, tc.percent, tc.obfuscator))
		})
	}
}

func TestObfuscateLeftPercent(t *testing.T) {
	tcs := []struct {
		name       string
		input      string
		percent    int
		expected   string
		obfuscator string
	}{
		{
			name:       "simple 90",
			input:      "simple",
			percent:    90,
			expected:   "*****e",
			obfuscator: "*",
		},
		{
			name:       "1 char 90",
			input:      "s",
			percent:    90,
			expected:   "s",
			obfuscator: "*",
		},
		{
			name:       "9 chars 90",
			input:      "123456789",
			percent:    90,
			expected:   "********9",
			obfuscator: "*",
		},
		{
			name:       "zero",
			input:      "123456789",
			percent:    0,
			expected:   "123456789",
			obfuscator: "*",
		},
		{
			name:       "all",
			input:      "123456789",
			percent:    100,
			expected:   "********9",
			obfuscator: "*",
		},
		{
			name:       "long",
			input:      "12345678901234567890",
			percent:    90,
			expected:   "------------------90",
			obfuscator: "-",
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			assert1.Equal(t, tc.expected, ObfuscateLeftPercent(tc.input, tc.percent, tc.obfuscator))
		})
	}
}
