package server

import "testing"

func TestParseCommand(t *testing.T) {
	var parseCmdTests = []struct {
		input    string
		want_cmd string
		want_arg string
	}{
		{"", "", ""},
		{" cmd ", "cmd", ""},
		{"cmd  ", "cmd", ""},
		{"cmd  arg ", "cmd", "arg"},
		{"cmd  arg1 arg2  arg3 ", "cmd", "arg1 arg2  arg3"},
	}

	for i, test := range parseCmdTests {
		if cmd, arg := parseCommand(test.input); string(cmd) != test.want_cmd || string(arg) != test.want_arg {
			t.Errorf("[%d] (parseCommand('%s') == ('%s', '%s')) != ('%s', '%s')", i, test.input,
				string(cmd), string(arg), test.want_cmd, test.want_arg)
		}
	}
}
