package gnap

import (
	"bytes"
	"context"
	"encoding/json"
	"sort"
	"strconv"

	"github.com/lestrrat-go/iter/mapiter"
	"github.com/pkg/errors"
)

type AccessTokenRequest struct {
	access      []*ResourceAccess
	flags       []AccessTokenAttribute
	label       *string
	extraFields map[string]interface{}
}

func (c *AccessTokenRequest) Get(key string) (interface{}, bool) {
	switch key {
	case "access":
		if len(c.access) == 0 {
			return nil, false
		}
		return c.access, true
	case "flags":
		if len(c.flags) == 0 {
			return nil, false
		}
		return c.flags, true
	case "label":
		if c.label == nil {
			return nil, false
		}
		return c.label, true
	default:
		if c.extraFields == nil {
			return nil, false
		}
		v, ok := c.extraFields[key]
		return v, ok
	}
}

func (c *AccessTokenRequest) Set(key string, value interface{}) error {
	switch key {
	case "access":
		if v, ok := value.([]*ResourceAccess); ok {
			c.access = v
		} else {
			return errors.Errorf(`invalid type for "access" (%T)`, value)
		}
	case "flags":
		if v, ok := value.([]AccessTokenAttribute); ok {
			c.flags = v
		} else {
			return errors.Errorf(`invalid type for "flags" (%T)`, value)
		}
	case "label":
		if v, ok := value.(string); ok {
			c.label = &v
		} else if value == nil {
			c.label = nil
		} else {
			return errors.Errorf(`invalid type for "label" (%T)`, value)
		}
	default:
		if c.extraFields == nil {
			c.extraFields = make(map[string]interface{})
		}
		c.extraFields[key] = value
	}
	return nil
}

func (c *AccessTokenRequest) AddAccess(v ...*ResourceAccess) {
	c.access = append(c.access, v...)
}

func (c *AccessTokenRequest) AddFlags(v ...AccessTokenAttribute) {
	c.flags = append(c.flags, v...)
}

func (c *AccessTokenRequest) SetLabel(v string) {
	c.label = &v
}

func (c AccessTokenRequest) MarshalJSON() ([]byte, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	buf.WriteByte('{')
	var i int
	for iter := c.Iterate(ctx); iter.Next(ctx); {
		pair := iter.Pair()
		if i > 0 {
			buf.WriteByte(',')
		}
		i++
		buf.WriteString(strconv.Quote(pair.Key.(string)))
		buf.WriteByte(':')
		if err := enc.Encode(pair.Value); err != nil {
			return nil, errors.Wrapf(err, `failed to encode %s`, pair.Key.(string))
		}
	}
	buf.WriteByte('}')
	return buf.Bytes(), nil
}

func (c *AccessTokenRequest) UnmarshalJSON(data []byte) error {
	c.access = nil
	c.flags = nil
	c.label = nil
	dec := json.NewDecoder(bytes.NewReader(data))
	tok, err := dec.Token()
	if err != nil {
		return errors.Wrap(err, `error reading token`)
	}
	switch tok := tok.(type) {
	case json.Delim:
		if tok != '{' {
			return errors.Errorf(`expected '{', but got '%c'`, tok)
		}
	default:
		return errors.Errorf(`expected '{', but got '%c'`, tok)
	}
LOOP:
	for {
		tok, err := dec.Token()
		if err != nil {
			return errors.Wrap(err, `error reading token`)
		}
		switch tok := tok.(type) {
		case json.Delim:
			if tok == '}' { // End of object
				break LOOP
			}
			return errors.Errorf(`unexpected delimiter '%c'`, tok)
		case string:
			switch tok {
			case "access":
				if err := dec.Decode(&(c.access)); err != nil {
					return errors.Wrap(err, `error reading access`)
				}
			case "flags":
				if err := dec.Decode(&(c.flags)); err != nil {
					return errors.Wrap(err, `error reading flags`)
				}
			case "label":
				var tmp string
				if err := dec.Decode(&tmp); err != nil {
					return errors.Wrap(err, `error reading label`)
				}
				c.label = &tmp
			default:
				var tmp interface{}
				if err := dec.Decode(&tmp); err != nil {
					return errors.Wrapf(err, `error reading %s`, tok)
				}
				if c.extraFields == nil {
					c.extraFields = map[string]interface{}{}
				}
				c.extraFields[tok] = tmp
			}
		}
	}
	return nil
}

func (c *AccessTokenRequest) makePairs() []*mapiter.Pair {
	var pairs []*mapiter.Pair
	if tmp := c.access; len(tmp) > 0 {
		pairs = append(pairs, &mapiter.Pair{Key: "access", Value: tmp})
	}
	if tmp := c.flags; len(tmp) > 0 {
		pairs = append(pairs, &mapiter.Pair{Key: "flags", Value: tmp})
	}
	if tmp := c.label; tmp != nil {
		pairs = append(pairs, &mapiter.Pair{Key: "label", Value: *tmp})
	}
	var extraKeys []string
	for k := range c.extraFields {
		extraKeys = append(extraKeys, k)
	}
	sort.Strings(extraKeys)
	for _, k := range extraKeys {
		pairs = append(pairs, &mapiter.Pair{Key: k, Value: c.extraFields[k]})
	}
	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].Key.(string) < pairs[j].Key.(string)
	})
	return pairs
}

func (c *AccessTokenRequest) Iterate(ctx context.Context) mapiter.Iterator {
	pairs := c.makePairs()
	ch := make(chan *mapiter.Pair, len(pairs))
	go func(ctx context.Context, ch chan *mapiter.Pair, pairs []*mapiter.Pair) {
		defer close(ch)
		for _, pair := range pairs {
			select {
			case <-ctx.Done():
				return
			case ch <- pair:
			}
		}
	}(ctx, ch, pairs)
	return mapiter.New(ch)
}
