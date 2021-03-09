package gnap

import (
	"bytes"
	"context"
	"encoding/json"
	"sort"
	"strconv"

	"github.com/lestrrat-go/iter/mapiter"
	"github.com/lestrrat-go/jwx/jwk"
	"github.com/pkg/errors"
)

type AccessToken struct {
	access      []ResourceAccess
	bound       *bool
	durable     *bool
	expires_in  *int64
	key         jwk.Key
	label       *string
	manage      *string
	split       *bool
	value       *string
	extraFields map[string]interface{}
}

func (c *AccessToken) Validate() error {
	if len(c.access) == 0 {
		return errors.Errorf(`field "access" is required`)
	}
	if c.value == nil {
		return errors.Errorf(`field "value" is required`)
	}
	return nil
}

func (c *AccessToken) Get(key string) (interface{}, bool) {
	switch key {
	case "access":
		if len(c.access) == 0 {
			return nil, false
		}
		return c.access, true
	case "bound":
		if c.bound == nil {
			return nil, false
		}
		return c.bound, true
	case "durable":
		if c.durable == nil {
			return nil, false
		}
		return c.durable, true
	case "expires_in":
		if c.expires_in == nil {
			return nil, false
		}
		return c.expires_in, true
	case "key":
		if c.key == nil {
			return nil, false
		}
		return c.key, true
	case "label":
		if c.label == nil {
			return nil, false
		}
		return c.label, true
	case "manage":
		if c.manage == nil {
			return nil, false
		}
		return c.manage, true
	case "split":
		if c.split == nil {
			return nil, false
		}
		return c.split, true
	case "value":
		if c.value == nil {
			return nil, false
		}
		return c.value, true
	default:
		if c.extraFields == nil {
			return nil, false
		}
		v, ok := c.extraFields[key]
		return v, ok
	}
}

func (c *AccessToken) Set(key string, value interface{}) error {
	switch key {
	case "access":
		if v, ok := value.([]ResourceAccess); ok {
			c.access = v
		} else {
			return errors.Errorf(`invalid type for "access" (%T)`, value)
		}
	case "bound":
		if v, ok := value.(*bool); ok {
			c.bound = v
		} else {
			return errors.Errorf(`invalid type for "bound" (%T)`, value)
		}
	case "durable":
		if v, ok := value.(*bool); ok {
			c.durable = v
		} else {
			return errors.Errorf(`invalid type for "durable" (%T)`, value)
		}
	case "expires_in":
		if v, ok := value.(*int64); ok {
			c.expires_in = v
		} else {
			return errors.Errorf(`invalid type for "expires_in" (%T)`, value)
		}
	case "key":
		if v, ok := value.(jwk.Key); ok {
			c.key = v
		} else {
			return errors.Errorf(`invalid type for "key" (%T)`, value)
		}
	case "label":
		if v, ok := value.(string); ok {
			c.label = &v
		} else if value == nil {
			c.label = nil
		} else {
			return errors.Errorf(`invalid type for "label" (%T)`, value)
		}
	case "manage":
		if v, ok := value.(string); ok {
			c.manage = &v
		} else if value == nil {
			c.manage = nil
		} else {
			return errors.Errorf(`invalid type for "manage" (%T)`, value)
		}
	case "split":
		if v, ok := value.(*bool); ok {
			c.split = v
		} else {
			return errors.Errorf(`invalid type for "split" (%T)`, value)
		}
	case "value":
		if v, ok := value.(string); ok {
			c.value = &v
		} else if value == nil {
			c.value = nil
		} else {
			return errors.Errorf(`invalid type for "value" (%T)`, value)
		}
	default:
		if c.extraFields == nil {
			c.extraFields = make(map[string]interface{})
		}
		c.extraFields[key] = value
	}
	return nil
}

func (c *AccessToken) AddAccess(v ...ResourceAccess) *AccessToken {
	c.access = append(c.access, v...)
	return c
}

func (c *AccessToken) Access() []ResourceAccess {
	return c.access
}

func (c *AccessToken) SetBound(v *bool) {
	c.bound = v
}

func (c *AccessToken) Bound() *bool {
	return c.bound
}

func (c *AccessToken) SetDurable(v *bool) {
	c.durable = v
}

func (c *AccessToken) Durable() *bool {
	return c.durable
}

func (c *AccessToken) SetExpiresIn(v *int64) {
	c.expires_in = v
}

func (c *AccessToken) ExpiresIn() *int64 {
	return c.expires_in
}

func (c *AccessToken) SetKey(v jwk.Key) {
	c.key = v
}

func (c *AccessToken) Key() jwk.Key {
	return c.key
}

func (c *AccessToken) SetLabel(v string) {
	c.label = &v
}

func (c *AccessToken) Label() string {
	if c.label == nil {
		return ""
	}
	return *(c.label)
}

func (c *AccessToken) SetManage(v string) {
	c.manage = &v
}

func (c *AccessToken) Manage() string {
	if c.manage == nil {
		return ""
	}
	return *(c.manage)
}

func (c *AccessToken) SetSplit(v *bool) {
	c.split = v
}

func (c *AccessToken) Split() *bool {
	return c.split
}

func (c *AccessToken) SetValue(v string) {
	c.value = &v
}

func (c *AccessToken) Value() string {
	if c.value == nil {
		return ""
	}
	return *(c.value)
}

func (c AccessToken) MarshalJSON() ([]byte, error) {
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

func (c *AccessToken) UnmarshalJSON(data []byte) error {
	c.access = nil
	c.bound = nil
	c.durable = nil
	c.expires_in = nil
	c.key = nil
	c.label = nil
	c.manage = nil
	c.split = nil
	c.value = nil
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
			case "bound":
				if err := dec.Decode(&(c.bound)); err != nil {
					return errors.Wrap(err, `error reading bound`)
				}
			case "durable":
				if err := dec.Decode(&(c.durable)); err != nil {
					return errors.Wrap(err, `error reading durable`)
				}
			case "expires_in":
				if err := dec.Decode(&(c.expires_in)); err != nil {
					return errors.Wrap(err, `error reading expires_in`)
				}
			case "key":
				if err := dec.Decode(&(c.key)); err != nil {
					return errors.Wrap(err, `error reading key`)
				}
			case "label":
				var tmp string
				if err := dec.Decode(&tmp); err != nil {
					return errors.Wrap(err, `error reading label`)
				}
				c.label = &tmp
			case "manage":
				var tmp string
				if err := dec.Decode(&tmp); err != nil {
					return errors.Wrap(err, `error reading manage`)
				}
				c.manage = &tmp
			case "split":
				if err := dec.Decode(&(c.split)); err != nil {
					return errors.Wrap(err, `error reading split`)
				}
			case "value":
				var tmp string
				if err := dec.Decode(&tmp); err != nil {
					return errors.Wrap(err, `error reading value`)
				}
				c.value = &tmp
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

func (c *AccessToken) makePairs() []*mapiter.Pair {
	var pairs []*mapiter.Pair
	if tmp := c.access; len(tmp) > 0 {
		pairs = append(pairs, &mapiter.Pair{Key: "access", Value: tmp})
	}
	if tmp := c.bound; tmp != nil {
		pairs = append(pairs, &mapiter.Pair{Key: "bound", Value: *tmp})
	}
	if tmp := c.durable; tmp != nil {
		pairs = append(pairs, &mapiter.Pair{Key: "durable", Value: *tmp})
	}
	if tmp := c.expires_in; tmp != nil {
		pairs = append(pairs, &mapiter.Pair{Key: "expires_in", Value: *tmp})
	}
	if tmp := c.label; tmp != nil {
		pairs = append(pairs, &mapiter.Pair{Key: "label", Value: *tmp})
	}
	if tmp := c.manage; tmp != nil {
		pairs = append(pairs, &mapiter.Pair{Key: "manage", Value: *tmp})
	}
	if tmp := c.split; tmp != nil {
		pairs = append(pairs, &mapiter.Pair{Key: "split", Value: *tmp})
	}
	if tmp := c.value; tmp != nil {
		pairs = append(pairs, &mapiter.Pair{Key: "value", Value: *tmp})
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

func (c *AccessToken) Iterate(ctx context.Context) mapiter.Iterator {
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
