package configs

import "github.com/brickingsoft/brick/pkg/mists"

func init() {
	mists.RegisterCustomMarshaler[Config](configMistMarshaler)
	mists.RegisterCustomMarshaler[*Config](configMistPtrMarshaler)
	mists.RegisterCustomUnmarshaler[Config](configMistUnmarshaler)
}
