package db

type str []byte

func (v str) get() result {
	return result{v, nil}
}

func (v str) set(k []byte) result {
	copy(v, k)
	return resultOk
}

func (v str) getKey(k []byte) result {
	return resultInvalidType
}

func (v str) setKey(k []byte, nv []byte) result {
	return resultInvalidType
}

func (v str) empty() bool {
	return false
}

func (v str) pop() result {
	return resultInvalidType
}

func (v str) push(k []byte) result {
	return resultInvalidType
}
