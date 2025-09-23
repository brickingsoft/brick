package bpack

import (
	"unsafe"
)

type HeaderField struct {
	Name  string
	Value []byte
}

type HeaderFields struct {
	fields []HeaderField
}

func (h *HeaderFields) Set(name []byte, value []byte) (err error) {
	h.fields = append(h.fields, HeaderField{string(name), value})
	return
}

type HeadValueIndex struct {
	Index  int
	Values map[string]int
}

type Dictionary struct {
	fields  []HeaderField
	indexes map[string]*HeadValueIndex
}

// Index
// get header field index, return i and j.
//
// when i < 0, means not found.
//
// when i > -1 and j > -1, means found field, the j is the field index.
//
// when i > -1 and j < 0, means found field name, the i is the name index.
func (dict *Dictionary) Index(name []byte, value []byte) (int, int) {
	if len(name) == 0 {
		return -1, -1
	}
	ns := unsafe.String(unsafe.SliceData(name), len(name))
	if ni := dict.indexes[ns]; ni != nil {
		if len(value) == 0 {
			return ni.Index, -1
		}
		vs := unsafe.String(unsafe.SliceData(value), len(value))
		if vi, has := ni.Values[vs]; has {
			return ni.Index, vi
		}
		return ni.Index, -1
	}
	return -1, -1
}

// Get field by index.
//
// when length of name is 0, means not exist.
//
// when length if value is 0, means name existed but value not exist.
func (dict *Dictionary) Get(i int) (name []byte, value []byte) {
	if i < 0 || i >= len(dict.fields) {
		return
	}
	field := dict.fields[i]
	name = unsafe.Slice(unsafe.StringData(field.Name), len(field.Name))
	value = field.Value
	return
}

func (dict *Dictionary) Load(fields []HeaderField) {
	if len(fields) == 0 {
		return
	}
	if len(dict.fields) > 0 {
		dict.Reset()
	}
	dict.fields = fields
	if dict.indexes == nil {
		dict.indexes = make(map[string]*HeadValueIndex)
	}
	for i, field := range dict.fields {
		ni := dict.indexes[field.Name]
		if ni == nil {
			ni = &HeadValueIndex{Index: -1, Values: make(map[string]int)}
			dict.indexes[field.Name] = ni
		}
		if len(field.Value) == 0 {
			ni.Index = i
			continue
		}
		vs := unsafe.String(unsafe.SliceData(field.Value), len(field.Value))
		ni.Values[vs] = i
	}
	return
}

func (dict *Dictionary) Range(f func(i int, name []byte, value []byte) bool) {
	for i, field := range dict.fields {
		name := unsafe.Slice(unsafe.StringData(field.Name), len(field.Name))
		if !f(i, name, field.Value) {
			return
		}
	}
}

func (dict *Dictionary) Reset() {
	dict.fields = dict.fields[:0]
	clear(dict.indexes)
}
