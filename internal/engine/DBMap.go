package engine

import "errors"

const ReservedKeyPrefix = '\x00'

type DBMap struct {
	dbMap  map[string][]byte
	closed bool
}

func NewDBMap() *DBMap {
	return &DBMap{map[string][]byte{}, false}
}

func (db *DBMap) Put(key string, value []byte) error {
	if db.closed {
		return errors.New("db is closed")
	}
	if key == "" {
		return errors.New("key is empty string")
	}
	if isReservedKey(key) {
		return errors.New("invalid key, contains reserved prefix '\x00'")
	}
	db.dbMap[key] = value
	return nil
}

func (db *DBMap) Get(key string) ([]byte, bool, error) {
	if db.closed {
		return nil, false, errors.New("db is closed")
	}
	if key == "" {
		return nil, false, errors.New("key is empty string")
	}

	value, ok := db.dbMap[key]

	if !ok {
		return nil, false, nil
	}

	return value, true, nil
}

func (db *DBMap) Delete(key string) error {
	if db.closed {
		return errors.New("db is closed")
	}
	if key == "" {
		return errors.New("key is empty string")
	}
	if isReservedKey(key) {
		return errors.New("invalid key, contains reserved prefix '\x00'")
	}

	delete(db.dbMap, key)
	return nil
}

func (db *DBMap) Close() error {
	if db.closed {
		return errors.New("db is closed")
	}
	db.closed = true
	return nil
}

func isReservedKey(key string) bool {
	if key[0] == ReservedKeyPrefix {
		return true
	}
	return false
}

var _ DB = (*DBMap)(nil)
