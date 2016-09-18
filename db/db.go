package db

import (
	"bytes"
	"io/ioutil"
	"log"
	"strconv"
	"time"
)

type result struct {
	value []byte
	err   error
}

var (
	resultOk            = result{[]byte("Ok"), nil}
	resultNotFound      = result{nil, ErrNotFound}
	resultInvalidFormat = result{nil, ErrInvalidFormat}
	resultInvalidIndex  = result{nil, ErrInvalidIndex}
	resultInvalidType   = result{nil, ErrInvalidType}
	resultKeyNotFound   = result{nil, ErrKeyNotFound}
)

type task struct {
	cmd Command
	arg []byte
	ret chan result
}

type key string

type value interface {
	get() result
	set(k []byte) result
	getKey(k []byte) result
	setKey(k []byte, nv []byte) result // nv = nil - remove key
	pop() result
	push(k []byte) result
	empty() bool
}

type Database struct {
	queue   chan task
	started bool
	closing bool
	log     *log.Logger
	m       map[key]value
	t       map[key]*time.Timer
	e       EventHandler
}

// New creates new Database instance
func New(cfg Config) (*Database, error) {
	d := &Database{
		queue:   make(chan task, cfg.QueueLength),
		started: false,
		closing: false,
		log:     cfg.Log,
		m:       make(map[key]value, 1024),
		t:       make(map[key]*time.Timer, 1024),
		e:       cfg.Handler,
	}

	if d.log == nil {
		d.log = log.New(ioutil.Discard, "", 0)
	}

	go d.run()

	return d, nil
}

func (d *Database) Close() error {
	d.closing = true

	if d.queue == nil {
		return ErrAlreadyClosed
	}

	close(d.queue)
	d.queue = nil

	return nil
}

func (d *Database) Exec(cmd Command, arg []byte) ([]byte, error) {
	if d.closing || !d.started {
		return nil, ErrAlreadyClosed
	}
	ret := make(chan result)
	d.queue <- task{cmd: cmd, arg: arg, ret: ret}
	result := <-ret
	return result.value, result.err
}

func (d *Database) run() {
	d.started = true
	for {
		t, ok := <-d.queue
		if !ok {
			break
		}
		switch t.cmd {
		case CommandNop:
			t.ret <- resultOk
		case CommandGet:
			t.ret <- d.get(t.arg)
		case CommandSet:
			t.ret <- d.set(t.arg)
		case CommandPush:
			t.ret <- d.push(t.arg)
		case CommandPop:
			t.ret <- d.pop(t.arg)
		case CommandRemove:
			t.ret <- d.remove(t.arg)
		case CommandTtl:
			t.ret <- d.ttl(t.arg)
		case CommandKeys:
			t.ret <- d.keys(t.arg)
		default:
			t.ret <- result{nil, ErrInvalidCommand}
		}
	}

	for _, t := range d.t {
		t.Stop()
	}

	d.started = false
}

func parseArg(arg []byte) [][]byte {
	result := make([][]byte, 0, 3)

	slash := false
	start := 0
	for i, b := range arg {
		switch b {
		case '\\':
			slash = true
		case ',':
			if !slash {
				result = append(result, bytes.TrimSpace(arg[start:i]))
				start = i + 1
			}
			fallthrough
		default:
			slash = false
		}
	}

	result = append(result, bytes.TrimSpace(arg[start:]))

	return result
}

func (d *Database) get(arg []byte) result {
	if len(arg) == 0 {
		return resultInvalidFormat
	}

	args := parseArg(arg)
	switch len(args) {
	case 1:
		if v, ok := d.m[key(args[0])]; ok {
			return v.get()
		} else {
			return resultNotFound
		}
	case 2:
		if v, ok := d.m[key(args[0])]; ok {
			return v.getKey(args[1])
		} else {
			return resultNotFound
		}
	default:
		return result{nil, ErrInvalidFormat}
	}
}

func (d *Database) set(arg []byte) result {
	if len(arg) == 0 {
		return resultInvalidFormat
	}

	args := parseArg(arg)

	switch len(args) {
	case 2:
		if v, ok := d.m[key(args[0])]; ok {
			return v.set(args[1])
		} else {
			d.m[key(args[0])] = str(args[1])
			return resultOk
		}
	case 3:
		if v, ok := d.m[key(args[0])]; ok {
			return v.setKey(args[1], args[2])
		} else {
			var v value = make(dict)
			v.setKey(args[1], args[2])
			d.m[key(args[0])] = v
			return resultOk
		}
	default:
		return resultInvalidFormat
	}
}

func (d *Database) push(arg []byte) result {
	if len(arg) == 0 {
		return resultInvalidFormat
	}

	args := parseArg(arg)
	switch len(args) {
	case 2:
		if v, ok := d.m[key(args[0])]; ok {
			return v.push(args[1])
		} else {
			var v value = new(list)
			v.push(args[1])
			d.m[key(args[0])] = v
			return resultOk
		}
	default:
		return resultInvalidFormat
	}
}

func (d *Database) pop(arg []byte) result {
	if len(arg) == 0 {
		return resultInvalidFormat
	}

	args := parseArg(arg)
	switch len(args) {
	case 1:
		k := key(args[0])
		if v, ok := d.m[k]; ok {
			r := v.pop()
			if v.empty() {
				delete(d.m, k)
				if t, ok := d.t[k]; ok {
					t.Stop()
					delete(d.t, k)
				}
			}
			return r
		} else {
			return resultNotFound
		}
	default:
		return resultInvalidFormat
	}
}

func (d *Database) remove(arg []byte) result {
	if len(arg) == 0 {
		return resultInvalidFormat
	}

	args := parseArg(arg)
	switch len(args) {
	case 1:
		k := key(args[0])
		if _, ok := d.m[k]; ok {
			delete(d.m, k)
			if t, ok := d.t[k]; ok {
				t.Stop()
				delete(d.t, k)
			}
			return resultOk
		} else {
			return resultNotFound
		}
	case 2:
		if v, ok := d.m[key(args[0])]; ok {
			return v.setKey(args[1], nil)
		} else {
			return resultNotFound
		}
	default:
		return resultInvalidFormat
	}
}

func (d *Database) ttl(arg []byte) result {
	if len(arg) == 0 {
		return resultInvalidFormat
	}

	args := parseArg(arg)
	switch len(args) {
	case 2:
		k := key(args[0])
		if timeout, err := strconv.ParseUint(string(args[1]), 10, 64); err != nil {
			return result{nil, err}
		} else if _, ok := d.m[k]; ok {
			duration := time.Millisecond * time.Duration(timeout)
			if t, ok := d.t[k]; ok {
				if !t.Stop() {
					<-t.C
				}
				t.Reset(duration)
			} else {
				t := time.AfterFunc(duration, func() {
					d.Exec(CommandRemove, args[0])
					if d.e != nil {
						d.e(EventExpired, args[0])
					}
				})
				d.t[k] = t
			}
			return resultOk
		} else {
			return resultNotFound
		}
	default:
		return resultInvalidFormat
	}
}

func (d *Database) keys(arg []byte) result {
	if arg == nil || bytes.Equal(arg, []byte("")) {
		var r []byte
		var first = false
		for k, _ := range d.m {
			if first {
				r = append(r, ',')
			} else {
				first = true
			}
			r = append(r, k...)
		}
		return result{r, nil}
	}

	args := parseArg(arg)
	switch len(args) {
	case 1:
		if v, ok := d.m[key(args[0])]; ok {
			if dv, ok := v.(dict); ok {
				var r []byte
				var first = false
				for k, _ := range dv {
					if first {
						r = append(r, ',')
					} else {
						first = true
					}
					r = append(r, k...)
				}
				return result{r, nil}
			}
			return resultNotFound
		} else {
			return resultNotFound
		}
	default:
		return resultInvalidFormat
	}
}
