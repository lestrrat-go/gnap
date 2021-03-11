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

type InteractionFinish struct {
	hash_method *string
	method      *FinishMode
	nonce       *string
	uri         *string
	extraFields map[string]interface{}
}

func NewInteractionFinish(method FinishMode, nonce string, uri string) *InteractionFinish {
	return &InteractionFinish{
		method: &method,
		nonce:  &nonce,
		uri:    &uri,
	}
}

func (c *InteractionFinish) Validate() error {
	if c.method == nil {
		return errors.Errorf(`field "method" is required`)
	}
	if c.nonce == nil {
		return errors.Errorf(`field "nonce" is required`)
	}
	if c.uri == nil {
		return errors.Errorf(`field "uri" is required`)
	}
	return nil
}

func (c *InteractionFinish) Get(key string) (interface{}, bool) {
	switch key {
	case "hash_method":
		if c.hash_method == nil {
			return nil, false
		}
		return c.hash_method, true
	case "method":
		if c.method == nil {
			return nil, false
		}
		return c.method, true
	case "nonce":
		if c.nonce == nil {
			return nil, false
		}
		return c.nonce, true
	case "uri":
		if c.uri == nil {
			return nil, false
		}
		return c.uri, true
	default:
		if c.extraFields == nil {
			return nil, false
		}
		v, ok := c.extraFields[key]
		return v, ok
	}
}

func (c *InteractionFinish) Set(key string, value interface{}) error {
	switch key {
	case "hash_method":
		if v, ok := value.(string); ok {
			c.hash_method = &v
		} else if value == nil {
			c.hash_method = nil
		} else {
			return errors.Errorf(`invalid type for "hash_method" (%T)`, value)
		}
	case "method":
		if v, ok := value.(*FinishMode); ok {
			c.method = v
		} else {
			return errors.Errorf(`invalid type for "method" (%T)`, value)
		}
	case "nonce":
		if v, ok := value.(string); ok {
			c.nonce = &v
		} else if value == nil {
			c.nonce = nil
		} else {
			return errors.Errorf(`invalid type for "nonce" (%T)`, value)
		}
	case "uri":
		if v, ok := value.(string); ok {
			c.uri = &v
		} else if value == nil {
			c.uri = nil
		} else {
			return errors.Errorf(`invalid type for "uri" (%T)`, value)
		}
	default:
		if c.extraFields == nil {
			c.extraFields = make(map[string]interface{})
		}
		c.extraFields[key] = value
	}
	return nil
}

func (c *InteractionFinish) SetHashMethod(v string) {
	c.hash_method = &v
}

func (c *InteractionFinish) HashMethod() string {
	if c.hash_method == nil {
		return ""
	}
	return *(c.hash_method)
}

func (c *InteractionFinish) SetMethod(v FinishMode) {
	c.method = &v
}

func (c *InteractionFinish) Method() FinishMode {
	if c.method == nil {
		return ""
	}
	return *(c.method)
}

func (c *InteractionFinish) SetNonce(v string) {
	c.nonce = &v
}

func (c *InteractionFinish) Nonce() string {
	if c.nonce == nil {
		return ""
	}
	return *(c.nonce)
}

func (c *InteractionFinish) SetURI(v string) {
	c.uri = &v
}

func (c *InteractionFinish) URI() string {
	if c.uri == nil {
		return ""
	}
	return *(c.uri)
}

func (c InteractionFinish) MarshalJSON() ([]byte, error) {
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

func (c *InteractionFinish) UnmarshalJSON(data []byte) error {
	c.hash_method = nil
	c.method = nil
	c.nonce = nil
	c.uri = nil
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
			case "hash_method":
				var tmp string
				if err := dec.Decode(&tmp); err != nil {
					return errors.Wrap(err, `error reading hash_method`)
				}
				c.hash_method = &tmp
			case "method":
				if err := dec.Decode(&(c.method)); err != nil {
					return errors.Wrap(err, `error reading method`)
				}
			case "nonce":
				var tmp string
				if err := dec.Decode(&tmp); err != nil {
					return errors.Wrap(err, `error reading nonce`)
				}
				c.nonce = &tmp
			case "uri":
				var tmp string
				if err := dec.Decode(&tmp); err != nil {
					return errors.Wrap(err, `error reading uri`)
				}
				c.uri = &tmp
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

func (c *InteractionFinish) makePairs() []*mapiter.Pair {
	var pairs []*mapiter.Pair
	if tmp := c.hash_method; tmp != nil {
		pairs = append(pairs, &mapiter.Pair{Key: "hash_method", Value: *tmp})
	}
	if tmp := c.method; tmp != nil {
		pairs = append(pairs, &mapiter.Pair{Key: "method", Value: *tmp})
	}
	if tmp := c.nonce; tmp != nil {
		pairs = append(pairs, &mapiter.Pair{Key: "nonce", Value: *tmp})
	}
	if tmp := c.uri; tmp != nil {
		pairs = append(pairs, &mapiter.Pair{Key: "uri", Value: *tmp})
	}
	var extraKeys []string
	for k := range c.extraFields {
		extraKeys = append(extraKeys, k)
	}
	for _, k := range extraKeys {
		pairs = append(pairs, &mapiter.Pair{Key: k, Value: c.extraFields[k]})
	}
	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].Key.(string) < pairs[j].Key.(string)
	})
	return pairs
}

func (c *InteractionFinish) Iterate(ctx context.Context) mapiter.Iterator {
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
