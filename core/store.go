package core

import "time"

var store map[string]*Obj

type Obj struct {
	Value  interface{}
	Expiry int64
}

func init() {
	store = make(map[string]*Obj)
}

func NewObj(value interface{}, duration int64) *Obj {
	var expiresAt int64 = -1

	if duration > 0 {
		expiresAt = time.Now().UnixMilli() + duration
	}

	return &Obj{
		Value:  value,
		Expiry: expiresAt,
	}
}

func Put(key string, newObject *Obj) {
	store[key] = newObject
}

func Get(k string) *Obj {

	v := store[k]

	if v != nil {
		if v.Expiry <= time.Now().UnixMilli() {
			delete(store, k)
			return nil
		}
	}

	return v
}

func DeleteKey(k string) bool {
	_, ok := store[k]
	if ok {
		delete(store, k)
		return true
	}
	return false
}
