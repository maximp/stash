package client

import "testing"

func TestEncoding(t *testing.T) {
	var encodingTests = []struct {
		input string
		wants string
	}{
		{"", ""},
		{"\\", "\\"},
		{"\\'", "\\'"},
		{"\n", "\\n"},
		{"\\n", "\\n"},
		{"\r", "\\r"},
		{"\\r", "\\r"},
	}

	for i, test := range encodingTests {
		if result := encode(test.input); result != test.wants {
			t.Errorf("[%d] '%s' = encode('%s') != '%s'", i, result, test.input, test.wants)
		} else if decoded := decode(result); decoded != decode(test.input) {
			t.Errorf("[%d] '%s' = decode(encode('%s')) != '%s'", i, decoded, test.input, test.input)
		}
	}
}

func TestDecoding(t *testing.T) {
	var decodingTests = []struct {
		input string
		wants string
	}{
		{"", ""},
		{"\\", "\\"},
		{"\\r", "\r"},
		{"\\n", "\n"},
		{"\\\\n", "\\\n"},
		{"\\'", "\\'"},
		{"\\\\\\r", "\\\\\r"},
		{"nr", "nr"},
		{"\\\\n\\\\r", "\\\n\\\r"},
	}

	for i, test := range decodingTests {
		if result := decode(test.input); result != test.wants {
			t.Errorf("[%d] test: '%s' = decode('%s') != %s", i, result, test.input, test.wants)
		}
	}
}
