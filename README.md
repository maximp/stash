# juno.test
Test task for Juno company

# Protocol
1. set name, value - set value for string type
	1. str type - set value
	1. list type (resize list to uint(value))

1. set name, key, value - set value to:
	1. str type (ErrInvalidType)
	1. dict type (name[key] = value)
	1. list type (name[int(key)] = value)

1. get name - get key value for any type
	1. str type - get value
	1. dict type - dict size
	1. list type - list size

1. push name, value - push key value to list type

1. pop name - pop key value from list type

1. keys - list of all keys
	1 keys name - keys of dict 'name'

1. ttl name, milliseconds - set TTL value for key from now

1. remove name - remove name from cache

1. remove name, key - remove key from name
	1. dict type - remove name[key]
	1. list type - name[int(key)] = ''
