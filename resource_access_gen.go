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

// ResourceAccess describes the access, resource, and metadata associated with them
type ResourceAccess struct {
	actions     []string
	datatypes   []string
	identifier  *string
	locations   []string
	typ         *string
	extraFields map[string]interface{}
}

func (c *ResourceAccess) Get(key string) (interface{}, bool) {
	switch key {
	case "actions":
		if len(c.actions) == 0 {
			return nil, false
		}
		return c.actions, true
	case "datatypes":
		if len(c.datatypes) == 0 {
			return nil, false
		}
		return c.datatypes, true
	case "identifier":
		if c.identifier == nil {
			return nil, false
		}
		return c.identifier, true
	case "locations":
		if len(c.locations) == 0 {
			return nil, false
		}
		return c.locations, true
	case "type":
		if c.typ == nil {
			return nil, false
		}
		return c.typ, true
	default:
		if c.extraFields == nil {
			return nil, false
		}
		v, ok := c.extraFields[key]
		return v, ok
	}
}

func (c *ResourceAccess) Set(key string, value interface{}) error {
	switch key {
	case "actions":
		if v, ok := value.([]string); ok {
			c.actions = v
		} else {
			return errors.Errorf(`invalid type for "actions" (%T)`, value)
		}
	case "datatypes":
		if v, ok := value.([]string); ok {
			c.datatypes = v
		} else {
			return errors.Errorf(`invalid type for "datatypes" (%T)`, value)
		}
	case "identifier":
		if v, ok := value.(string); ok {
			c.identifier = &v
		} else if value == nil {
			c.identifier = nil
		} else {
			return errors.Errorf(`invalid type for "identifier" (%T)`, value)
		}
	case "locations":
		if v, ok := value.([]string); ok {
			c.locations = v
		} else {
			return errors.Errorf(`invalid type for "locations" (%T)`, value)
		}
	case "type":
		if v, ok := value.(string); ok {
			c.typ = &v
		} else if value == nil {
			c.typ = nil
		} else {
			return errors.Errorf(`invalid type for "type" (%T)`, value)
		}
	default:
		if c.extraFields == nil {
			c.extraFields = make(map[string]interface{})
		}
		c.extraFields[key] = value
	}
	return nil
}

func (c *ResourceAccess) AddActions(v ...string) {
	c.actions = append(c.actions, v...)
}

func (c *ResourceAccess) Actions() []string {
	return c.actions
}

func (c *ResourceAccess) AddDataTypes(v ...string) {
	c.datatypes = append(c.datatypes, v...)
}

func (c *ResourceAccess) DataTypes() []string {
	return c.datatypes
}

func (c *ResourceAccess) SetIdentifier(v string) {
	c.identifier = &v
}

func (c *ResourceAccess) Identifier() string {
	if c.identifier == nil {
		return ""
	}
	return *(c.identifier)
}

func (c *ResourceAccess) AddLocations(v ...string) {
	c.locations = append(c.locations, v...)
}

func (c *ResourceAccess) Locations() []string {
	return c.locations
}

func (c *ResourceAccess) SetType(v string) {
	c.typ = &v
}

func (c *ResourceAccess) Type() string {
	if c.typ == nil {
		return ""
	}
	return *(c.typ)
}

func (c ResourceAccess) MarshalJSON() ([]byte, error) {
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

func (c *ResourceAccess) UnmarshalJSON(data []byte) error {
	c.actions = nil
	c.datatypes = nil
	c.identifier = nil
	c.locations = nil
	c.typ = nil
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
			case "actions":
				if err := dec.Decode(&(c.actions)); err != nil {
					return errors.Wrap(err, `error reading actions`)
				}
			case "datatypes":
				if err := dec.Decode(&(c.datatypes)); err != nil {
					return errors.Wrap(err, `error reading datatypes`)
				}
			case "identifier":
				var tmp string
				if err := dec.Decode(&tmp); err != nil {
					return errors.Wrap(err, `error reading identifier`)
				}
				c.identifier = &tmp
			case "locations":
				if err := dec.Decode(&(c.locations)); err != nil {
					return errors.Wrap(err, `error reading locations`)
				}
			case "type":
				var tmp string
				if err := dec.Decode(&tmp); err != nil {
					return errors.Wrap(err, `error reading type`)
				}
				c.typ = &tmp
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

func (c *ResourceAccess) makePairs() []*mapiter.Pair {
	var pairs []*mapiter.Pair
	if tmp := c.actions; len(tmp) > 0 {
		pairs = append(pairs, &mapiter.Pair{Key: "actions", Value: tmp})
	}
	if tmp := c.datatypes; len(tmp) > 0 {
		pairs = append(pairs, &mapiter.Pair{Key: "datatypes", Value: tmp})
	}
	if tmp := c.identifier; tmp != nil {
		pairs = append(pairs, &mapiter.Pair{Key: "identifier", Value: *tmp})
	}
	if tmp := c.locations; len(tmp) > 0 {
		pairs = append(pairs, &mapiter.Pair{Key: "locations", Value: tmp})
	}
	if tmp := c.typ; tmp != nil {
		pairs = append(pairs, &mapiter.Pair{Key: "type", Value: *tmp})
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

func (c *ResourceAccess) Iterate(ctx context.Context) mapiter.Iterator {
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
