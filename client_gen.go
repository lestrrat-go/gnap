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

type Client struct {
	classID     *string
	instanceID  *string
	key         *Key
	extraFields map[string]interface{}
}

func (c *Client) Get(key string) (interface{}, bool) {
	switch key {
	case "class_id":
		if c.classID == nil {
			return nil, false
		}
		return c.classID, true
	case "instance_id":
		if c.instanceID == nil {
			return nil, false
		}
		return c.instanceID, true
	case "key":
		if c.key == nil {
			return nil, false
		}
		return c.key, true
	default:
		if c.extraFields == nil {
			return nil, false
		}
		v, ok := c.extraFields[key]
		return v, ok
	}
}

func (c *Client) Set(key string, value interface{}) error {
	switch key {
	case "class_id":
		if v, ok := value.(string); ok {
			c.classID = &v
		} else if value == nil {
			c.classID = nil
		} else {
			return errors.Errorf(`invalid type for "class_id" (%T)`, value)
		}
	case "instance_id":
		if v, ok := value.(string); ok {
			c.instanceID = &v
		} else if value == nil {
			c.instanceID = nil
		} else {
			return errors.Errorf(`invalid type for "instance_id" (%T)`, value)
		}
	case "key":
		if v, ok := value.(*Key); ok {
			c.key = v
		} else {
			return errors.Errorf(`invalid type for "key" (%T)`, value)
		}
	default:
		if c.extraFields == nil {
			c.extraFields = make(map[string]interface{})
		}
		c.extraFields[key] = value
	}
	return nil
}

func (c *Client) SetClassID(v string) {
	c.classID = &v
}

func (c *Client) ClassID() string {
	if c.classID == nil {
		return ""
	}
	return *(c.classID)
}

func (c *Client) SetInstanceID(v string) {
	c.instanceID = &v
}

func (c *Client) InstanceID() string {
	if c.instanceID == nil {
		return ""
	}
	return *(c.instanceID)
}

func (c *Client) SetKey(v *Key) {
	c.key = v
}

func (c *Client) Key() *Key {
	return c.key
}

func (c Client) MarshalJSON() ([]byte, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	buf.WriteByte('{')
	var pairs []*mapiter.Pair
	for iter := c.Iterate(ctx); iter.Next(ctx); {
		pairs = append(pairs, iter.Pair())
	}
	if len(pairs) == 1 && pairs[0].Key.(string) == "instance_id" {
		return []byte(strconv.Quote(pairs[0].Value.(string))), nil
	}
	for i, pair := range pairs {
		if i > 0 {
			buf.WriteByte(',')
		}
		buf.WriteString(strconv.Quote(pair.Key.(string)))
		buf.WriteByte(':')
		if err := enc.Encode(pair.Value); err != nil {
			return nil, errors.Wrapf(err, `failed to encode %s`, pair.Key.(string))
		}
	}
	buf.WriteByte('}')
	return buf.Bytes(), nil
}

func (c *Client) UnmarshalJSON(data []byte) error {
	c.classID = nil
	c.instanceID = nil
	c.key = nil
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
	case string:
		c.instanceID = &tok
		return nil
	default:
		return errors.Errorf(`expected '{' or string, but got '%c'`, tok)
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
			case "class_id":
				var tmp string
				if err := dec.Decode(&tmp); err != nil {
					return errors.Wrap(err, `error reading class_id`)
				}
				c.classID = &tmp
			case "instance_id":
				var tmp string
				if err := dec.Decode(&tmp); err != nil {
					return errors.Wrap(err, `error reading instance_id`)
				}
				c.instanceID = &tmp
			case "key":
				if err := dec.Decode(&(c.key)); err != nil {
					return errors.Wrap(err, `error reading key`)
				}
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

func (c *Client) makePairs() []*mapiter.Pair {
	var pairs []*mapiter.Pair
	if tmp := c.classID; tmp != nil {
		pairs = append(pairs, &mapiter.Pair{Key: "class_id", Value: *tmp})
	}
	if tmp := c.instanceID; tmp != nil {
		pairs = append(pairs, &mapiter.Pair{Key: "instance_id", Value: *tmp})
	}
	if tmp := c.key; tmp != nil {
		pairs = append(pairs, &mapiter.Pair{Key: "key", Value: *tmp})
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

func (c *Client) Iterate(ctx context.Context) mapiter.Iterator {
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
