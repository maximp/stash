package db

import "strconv"

type list struct {
	v [][]byte
}

func (v *list) get() result {
	return result{strconv.AppendInt(nil, int64(len(v.v)), 10), nil}
}

func (v *list) set(k []byte) result {
	i, err := strconv.ParseUint(string(k), 10, 64)
	if err != nil {
		return result{nil, err}
	}

	if i < uint64(len(v.v)) {
		v.v = v.v[:i]
	} else {
		var n = make([][]byte, i)
		copy(n, v.v)
		v.v = n
	}

	return resultOk
}

func (v *list) getKey(k []byte) result {
	if i, err := strconv.ParseUint(string(k), 10, 64); err != nil {
		return result{nil, err}
	} else if int(i) >= len(v.v) {
		return resultInvalidIndex
	} else {
		return result{v.v[i], nil}
	}
}

func (v *list) setKey(k []byte, nv []byte) result {
	if i, err := strconv.ParseUint(string(k), 10, 64); err != nil {
		return result{nil, err}
	} else if int(i) >= len(v.v) {
		return resultInvalidIndex
	} else {
		v.v[i] = nv
		return resultOk
	}
}

func (v *list) empty() bool {
	return len(v.v) == 0
}

func (v *list) pop() result {
	r := result{v.v[len(v.v)-1], nil}
	v.v = v.v[:len(v.v)-1]
	return r
}

func (v *list) push(k []byte) result {
	v.v = append(v.v, k)
	return resultOk
}
