package avro

func Marshal(v any) (b []byte, err error) {

	return
}

func Unmarshal(b []byte, v any) (err error) {

	return
}

type Marshaler interface {
	MarshalAVRO(v any) (b []byte, err error)
}

type Unmarshaler interface {
	UnmarshalAVRO(b []byte, v any) (err error)
}
