package db

import "strconv"

type dict map[key][]byte

func (v dict) get() result {
	return result{[]byte(strconv.Itoa(len(v))), nil}
}

func (v dict) set(k []byte) result {
	return resultInvalidType
}

func (v dict) getKey(k []byte) result {
	if val, ok := v[key(k)]; ok {
		return result{val, nil}
	}

	return resultKeyNotFound
}

func (v dict) setKey(k []byte, nv []byte) result {
	if nv == nil {
		delete(v, key(k))
	} else {
		v[key(k)] = nv
	}
	return resultOk
}

func (v dict) empty() bool {
	return len(v) == 0
}

func (v dict) pop() result {
	return resultInvalidType
}

func (v dict) push(k []byte) result {
	return resultInvalidType
}
