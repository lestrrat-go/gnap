package json

import (
	"bytes"
	"sync"

	"github.com/pkg/errors"
)

var muGlobalConfig sync.RWMutex
var useNumber bool

// Sets the global configuration for json decoding
func DecoderSettings(inUseNumber bool) {
	muGlobalConfig.Lock()
	useNumber = inUseNumber
	muGlobalConfig.Unlock()
}

// Unmarshal respects the values specified in DecoderSettings,
// and uses a Decoder that has certain features turned on/off
func Unmarshal(b []byte, v interface{}) error {
	dec := NewDecoder(bytes.NewReader(b))
	return dec.Decode(v)
}

func DecodeInto(dec *Decoder, dst interface{}) error {
	if err := dec.Decode(dst); err != nil {
		return errors.Wrap(err, `error reading next value`)
	}
	return nil
}
