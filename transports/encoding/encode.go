package encoding

import (
	"fmt"

	"github.com/brickingsoft/brick/pkg/avro"
)

type Encoder interface {
	Marshal(v any) (b []byte, err error)
	Unmarshal(b []byte, v any) (err error)
}

type AvroEncoder struct{}

func (encoder *AvroEncoder) Marshal(v any) ([]byte, error) {
	return avro.Marshal(v)
}

func (encoder *AvroEncoder) Unmarshal(b []byte, v any) error {
	return avro.Unmarshal(b, v)
}

type JsonEncoder struct{}

func (encoder *JsonEncoder) Marshal(v any) ([]byte, error) {
	return avro.Marshal(v)
}

func (encoder *JsonEncoder) Unmarshal(b []byte, v any) error {
	return avro.Unmarshal(b, v)
}

type InvalidEncoder struct {
	name string
}

func (encoder InvalidEncoder) Marshal(_ any) (b []byte, err error) {
	err = fmt.Errorf("encoder not found for name %s", encoder.name)
	return
}

func (encoder InvalidEncoder) Unmarshal(_ []byte, _ any) (err error) {
	err = fmt.Errorf("encoder not found for name %s", encoder.name)
	return
}

const (
	JsonEncoderType = "json"
	AvroEncoderType = "avro"
)

var (
	encoders            = make(map[string]Encoder)
	avroEncoder Encoder = new(AvroEncoder)
	jsonEncoder Encoder = new(JsonEncoder)
)

func Register(name string, encoder Encoder) {
	switch name {
	case AvroEncoderType:
		avroEncoder = encoder
	case JsonEncoderType:
		jsonEncoder = encoder
		break
	default:
		encoders[name] = encoder
		break
	}
}

func Retrieve(name string) Encoder {
	switch name {
	case AvroEncoderType:
		return avroEncoder
	case JsonEncoderType:
		return jsonEncoder
	default:
		if encoder, ok := encoders[name]; ok {
			return encoder
		}
		return &InvalidEncoder{name: name}
	}
}
