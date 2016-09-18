package db

import (
	"bytes"
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"
)

var _ fmt.Stringer = CommandGet

type testLog struct {
	*testing.T
}

func (t *testLog) Write(p []byte) (n int, err error) {
	t.Log(string(bytes.TrimSpace(p)))
	return len(p), nil
}

func create_db(t *testing.T) *Database {
	d, err := New(Config{
		QueueLength: 10,
		Log:         log.New(&testLog{t}, "", log.LstdFlags|log.Lmicroseconds),
	})
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(time.Millisecond)
	return d
}

func TestDatabaseCreateClose(t *testing.T) {
	dd, err := New(Config{
		QueueLength: 10,
	})

	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(time.Millisecond)

	if !dd.started {
		t.Fatal("database is not started")
	}

	if _, err := dd.Exec(CommandNop, nil); err != nil {
		t.Errorf("failed execute command 'nop', err = %v", err)
	}

	// test TTL timers stopping
	if _, err := dd.Exec(CommandSet, []byte("a,b")); err != nil {
		t.Errorf("failed execute command 'set a,b', err = %v", err)
	}
	if _, err := dd.Exec(CommandTtl, []byte("a,1000000")); err != nil {
		t.Errorf("failed execute command 'ttl a,N', err = %v", err)
	}

	dd.Close()
	if !dd.started {
		t.Errorf("database is started")
	}

	if err := dd.Close(); err != ErrAlreadyClosed {
		t.Errorf("second db close must return ErrAlreadyClosed, returned = %v", err)
	}

	if _, err := dd.Exec(CommandNop, nil); err != ErrAlreadyClosed {
		t.Errorf("command execution must failed with ErrAlreadyClosed after db close, err = %v", err)
	}
}

func TestDatabaseInvalidCmd(t *testing.T) {
	dd := create_db(t)
	defer dd.Close()

	if _, err := dd.Exec(Command(987654321), []byte("")); err != ErrInvalidCommand {
		t.Error("err is %v", err)
	}
}

func TestDatabaseNop(t *testing.T) {
	dd := create_db(t)
	defer dd.Close()

	if r, err := dd.Exec(CommandNop, nil); err != nil || !bytes.Equal(r, resultOk.value) {
		t.Errorf("nop failed with '%s', %v", r, err)
	}

	if r, err := dd.Exec(CommandNop, []byte("arg1,arg2, arg3")); err != nil || !bytes.Equal(r, resultOk.value) {
		t.Errorf("nop, args failed with '%s', %v", r, err)
	}
}

func TestDatabaseNotFound(t *testing.T) {
	dd := create_db(t)
	defer dd.Close()

	dd.Exec(CommandSet, []byte("str,value"))
	dd.Exec(CommandPush, []byte("list,1"))

	var tests = []struct {
		cmd Command
		arg string
	}{
		{CommandGet, "name"},
		{CommandGet, "name,key"},
		{CommandPop, "name"},
		{CommandRemove, "name"},
		{CommandRemove, "name,key"},
		{CommandTtl, "name,12345"},
		{CommandKeys, "name"},
		{CommandKeys, "str"},
		{CommandKeys, "list"},
	}

	for i, test := range tests {
		r, err := dd.Exec(test.cmd, []byte(test.arg))
		if !bytes.Equal(r, []byte("")) || err != ErrNotFound {
			t.Errorf("[%d] - '%s %s' failed with '%s' (%v), expected: ''",
				i, test.cmd, test.arg, r, err)
		}
	}
}

func TestDatabaseInvalidFormat(t *testing.T) {
	dd := create_db(t)
	defer dd.Close()

	var tests = []struct {
		cmd Command
		arg string
	}{
		{CommandGet, ""},
		{CommandGet, "1,2,3"},
		{CommandSet, ""},
		{CommandSet, "1"},
		{CommandSet, "1,2,3,4"},
		{CommandPush, ""},
		{CommandPush, "1"},
		{CommandPush, "1,2,3"},
		{CommandPop, ""},
		{CommandPop, "1,2"},
		{CommandRemove, ""},
		{CommandRemove, "a,b,c"},
		{CommandTtl, ""},
		{CommandTtl, "a"},
		{CommandTtl, "a,b,c"},
		{CommandKeys, "x,y"},
	}

	for i, test := range tests {
		r, err := dd.Exec(test.cmd, []byte(test.arg))
		if !bytes.Equal(r, []byte("")) || err != ErrInvalidFormat {
			t.Errorf("[%d] - '%s %s' failed with '%s' (%v), expected: ''",
				i, test.cmd, test.arg, r, err)
		}
	}
}

func TestDatabaseInvalidType(t *testing.T) {
	dd := create_db(t)
	defer dd.Close()

	if _, err := dd.Exec(CommandSet, []byte("str, value")); err != nil {
		t.Fatal("failed set str value")
	}
	if _, err := dd.Exec(CommandSet, []byte("dict,key,value")); err != nil {
		t.Fatal("failed set dict value")
	}

	var tests = []struct {
		cmd Command
		arg string
	}{
		{CommandGet, "str,key"},
		{CommandSet, "str,key,value"},
		{CommandSet, "dict,key"},
		{CommandPush, "dict,value"},
		{CommandPop, "dict"},
		{CommandPush, "str,value"},
		{CommandPop, "str"},
	}

	for i, test := range tests {
		r, err := dd.Exec(test.cmd, []byte(test.arg))
		if !bytes.Equal(r, []byte("")) || err != ErrInvalidType {
			t.Errorf("[%d] - '%s %s' failed with '%s' (%v), expected: '' (%v)",
				i, test.cmd, test.arg, r, err, ErrInvalidType)
		}
	}
}

func TestDatabaseInvalidIndex(t *testing.T) {
	dd := create_db(t)
	defer dd.Close()

	if _, err := dd.Exec(CommandPush, []byte("list,value")); err != nil {
		t.Fatal("failed push list value")
	}

	var tests = []struct {
		cmd Command
		arg string
	}{
		{CommandGet, "list,1"},
		{CommandSet, "list,1,new-value"},
	}

	for i, test := range tests {
		r, err := dd.Exec(test.cmd, []byte(test.arg))
		if !bytes.Equal(r, []byte("")) || err != ErrInvalidIndex {
			t.Errorf("[%d] - '%s %s' failed with '%s' (%v), expected: '' (%v)",
				i, test.cmd, test.arg, r, err, ErrInvalidIndex)
		}
	}
}

func TestDatabaseNumError(t *testing.T) {
	dd := create_db(t)
	defer dd.Close()

	if _, err := dd.Exec(CommandPush, []byte("list,value")); err != nil {
		t.Fatal("failed push list value")
	}

	var tests = []struct {
		cmd Command
		arg string
	}{
		{CommandGet, "list,xxx"},
		{CommandSet, "list,xxx"},
		{CommandSet, "list,xxx,new-value"},
		{CommandTtl, "list,new-value"},
	}

	for i, test := range tests {
		r, err := dd.Exec(test.cmd, []byte(test.arg))
		e, ok := err.(*strconv.NumError)
		if !bytes.Equal(r, []byte("")) || !ok || (err != nil && e == nil) {
			t.Errorf("[%d] - '%s %s' failed with '%s' (%v), expected: '' (%v)",
				i, test.cmd, test.arg, r, err, ErrInvalidIndex)
		}
	}
}

func TestDatabaseString(t *testing.T) {
	dd := create_db(t)
	defer dd.Close()

	var tests = []struct {
		cmd    Command
		arg    string
		result string
		err    error
	}{
		{CommandSet, "str,1", "Ok", nil},                // create string key 'str'->'1'
		{CommandGet, "str", "1", nil},                   // get string key '1'
		{CommandSet, "str,2", "Ok", nil},                // modify string key 'str'->'2'
		{CommandGet, "str", "2", nil},                   // get string key '2'
		{CommandRemove, "str", "Ok", nil},               // remove string key
		{CommandGet, "str", "", ErrNotFound},            // get removed string key
		{CommandSet, "str a\\,bc,\\,cde\\,", "Ok", nil}, // set string value with commas
		{CommandGet, "str a\\,bc", "\\,cde\\,", nil},    // get string value with commas
	}

	for i, test := range tests {
		r, err := dd.Exec(test.cmd, []byte(test.arg))
		if !bytes.Equal(r, []byte(test.result)) || err != test.err {
			t.Errorf("[%d] - '%s %s' failed with '%s' (%v), expected: '%s' (%v)",
				i, test.cmd, test.arg, r, err, test.result, test.err)
		}
	}
}

func TestDatabaseList(t *testing.T) {
	dd := create_db(t)
	defer dd.Close()

	var tests = []struct {
		cmd    Command
		arg    string
		result string
		err    error
	}{
		{CommandPush, "list,1", "Ok", nil},  // push to list, create list value '1'
		{CommandPush, "list,2", "Ok", nil},  // push to list, create list value '2'
		{CommandGet, "list,0", "1", nil},    // get from list by index
		{CommandGet, "list,1", "2", nil},    // get from list by index
		{CommandSet, "list,0,0", "Ok", nil}, // change list by index
		{CommandSet, "list,1,1", "Ok", nil}, // change list by index

		{CommandGet, "list,0", "0", nil}, // get from list by index
		{CommandGet, "list,1", "1", nil}, // get from list by index

		{CommandSet, "list,1", "Ok", nil},   // change list size to 1
		{CommandGet, "list", "1", nil},      // check list size
		{CommandSet, "list,2", "Ok", nil},   // change list size to 2
		{CommandGet, "list", "2", nil},      // check list size
		{CommandGet, "list,1", "", nil},     // check list[1] == ''
		{CommandSet, "list,1,1", "Ok", nil}, // set list[1] == 2

		{CommandGet, "list", "2", nil}, // get list size
		{CommandPop, "list", "1", nil}, // pop from list
		{CommandGet, "list", "1", nil}, // get list size
		{CommandPop, "list", "0", nil}, // pop from list (autodelete empty)
	}

	for i, test := range tests {
		r, err := dd.Exec(test.cmd, []byte(test.arg))
		if !bytes.Equal(r, []byte(test.result)) || err != test.err {
			t.Errorf("[%d] - '%s %s' failed with '%s' (%v), expected: '%s' (%v)",
				i, test.cmd, test.arg, r, err, test.result, test.err)
		}
	}
}

func TestDatabaseDict(t *testing.T) {
	dd := create_db(t)
	defer dd.Close()

	var tests = []struct {
		cmd    Command
		arg    string
		result string
		err    error
	}{
		{CommandSet, "dict,key1,value1", "Ok", nil}, // create dict value 1
		{CommandGet, "dict", "1", nil},              // get dict keys count
		{CommandKeys, "dict", "key1", nil},          // get dict keys
		{CommandSet, "dict,key2,value2", "Ok", nil}, // create dict value '2'

		{CommandGet, "dict,key1", "value1", nil}, // get from dict key1
		{CommandGet, "dict,key2", "value2", nil}, // get from dict key2

		{CommandSet, "dict,key1,nvalue1", "Ok", nil}, // change dict key1
		{CommandSet, "dict,key2,nvalue2", "Ok", nil}, // change dict key2

		{CommandGet, "dict,key1", "nvalue1", nil}, // get from dict key1
		{CommandGet, "dict,key2", "nvalue2", nil}, // get from dict key2

		{CommandRemove, "dict,key2", "Ok", nil},       // remove from dict key2
		{CommandKeys, "dict", "key1", nil},            // get dict keys
		{CommandGet, "dict,key1", "nvalue1", nil},     // get from dict key1
		{CommandGet, "dict,key2", "", ErrKeyNotFound}, // get from dict key2

		{CommandSet, "dict,key2,nvalue2", "Ok", nil}, // set dict key2 to test get keys
	}

	for i, test := range tests {
		r, err := dd.Exec(test.cmd, []byte(test.arg))

		if !bytes.Equal(r, []byte(test.result)) || err != test.err {
			t.Errorf("[%d] - '%s %s' failed with '%s' (%v), expected: '%s' (%v)",
				i, test.cmd, test.arg, r, err, test.result, test.err)
		}
	}

	r, err := dd.Exec(CommandKeys, []byte("dict"))
	if err != nil {
		t.Fatalf("get dict failed with %v", err)
	}

	keys := strings.Split(string(r), ",")
	sort.Strings(keys)
	if str := strings.Join(keys, ","); str != "key1,key2" {
		t.Fatalf("get dict failed with value %v", str)
	}
}

func TestDatabaseTtl(t *testing.T) {
	dd := create_db(t)
	defer dd.Close()

	var (
		evt_received bool = false
		evt          Event
		evt_name     string
	)
	dd.e = func(e Event, name []byte) {
		evt_received = true
		evt = e
		evt_name = string(name)
	}

	if _, err := dd.Exec(CommandSet, []byte("str, value")); err != nil {
		t.Fatal("failed set str value")
	}

	if _, err := dd.Exec(CommandTtl, []byte("str, 1")); err != nil {
		t.Errorf("failed set ttl str value to 1 ms: %v", err)
	}

	time.Sleep(10 * time.Millisecond)

	if _, err := dd.Exec(CommandGet, []byte("str")); err != ErrNotFound {
		t.Error("found str value after expired ttl")
	}

	if !evt_received || evt != EventExpired || evt_name != "str" {
		t.Errorf("event problem: rcvd = %v, evt = %v, name = %s",
			evt_received, evt, evt_name)
	}

	evt_received = false
	evt_name = ""

	// create list var
	if _, err := dd.Exec(CommandPush, []byte("list, value")); err != nil {
		t.Fatal("failed push list value")
	}

	// set huge TTL
	if _, err := dd.Exec(CommandTtl, []byte("list, 1000000")); err != nil {
		t.Errorf("failed set ttl list value to 1000000 ms: %v", err)
	}

	time.Sleep(10 * time.Millisecond)

	// change TTL
	if _, err := dd.Exec(CommandTtl, []byte("list, 100000")); err != nil {
		t.Errorf("failed set ttl list value to 100000 ms: %v", err)
	}

	// pop list, autoremove
	if _, err := dd.Exec(CommandPop, []byte("list")); err != nil {
		t.Errorf("failed set pop list value: %v", err)
	}

	if evt_received || evt_name == "list" {
		t.Errorf("event problem after list pop: rcvd = %v, evt = %v, name = %s",
			evt_received, evt, evt_name)
	}
}

func TestDatabaseTtlExpireAfterChange(t *testing.T) {
	dd := create_db(t)
	defer dd.Close()

	var (
		evt_received bool = false
		evt          Event
		evt_name     string
	)
	dd.e = func(e Event, name []byte) {
		evt_received = true
		evt = e
		evt_name = string(name)
	}

	if _, err := dd.Exec(CommandSet, []byte("str, value")); err != nil {
		t.Fatal("failed set str value")
	}

	if _, err := dd.Exec(CommandTtl, []byte("str, 1000000")); err != nil {
		t.Errorf("failed set ttl str value to 1000000 ms: %v", err)
	}

	time.Sleep(10 * time.Millisecond)

	// change TTL
	if _, err := dd.Exec(CommandTtl, []byte("str, 10")); err != nil {
		t.Errorf("failed set ttl list value to 100000 ms: %v", err)
	}

	time.Sleep(20 * time.Millisecond)

	if _, err := dd.Exec(CommandGet, []byte("str")); err != ErrNotFound {
		t.Error("found str value after expired ttl")
	}

	if !evt_received || evt != EventExpired || evt_name != "str" {
		t.Errorf("event problem: rcvd = %v, evt = %v, name = %s",
			evt_received, evt, evt_name)
	}
}

func TestDatabaseKeys(t *testing.T) {
	dd := create_db(t)
	defer dd.Close()

	if _, err := dd.Exec(CommandSet, []byte("str, value")); err != nil {
		t.Errorf("failed set str value: %v", err)
	}
	if _, err := dd.Exec(CommandPush, []byte("list, value")); err != nil {
		t.Errorf("failed set list value: %v", err)
	}
	if _, err := dd.Exec(CommandSet, []byte("dict, name, value")); err != nil {
		t.Errorf("failed set dict value: %v", err)
	}

	r, err := dd.Exec(CommandKeys, nil)
	if err != nil {
		t.Fatalf("failed get db keys: %v", err)
	}

	keys := strings.Split(string(r), ",")
	sort.Strings(keys)
	if str := strings.Join(keys, ","); str != "dict,list,str" {
		t.Fatalf("keys failed with value %v", str)
	}
}
