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

type Interaction struct {
	finish      []*InteractionFinish
	start       []StartMode
	extraFields map[string]interface{}
}

func (c *Interaction) Validate() error {
	return nil
}

func (c *Interaction) Get(key string) (interface{}, bool) {
	switch key {
	case "finish":
		if len(c.finish) == 0 {
			return nil, false
		}
		return c.finish, true
	case "start":
		if len(c.start) == 0 {
			return nil, false
		}
		return c.start, true
	default:
		if c.extraFields == nil {
			return nil, false
		}
		v, ok := c.extraFields[key]
		return v, ok
	}
}

func (c *Interaction) Set(key string, value interface{}) error {
	switch key {
	case "finish":
		if v, ok := value.([]*InteractionFinish); ok {
			c.finish = v
		} else {
			return errors.Errorf(`invalid type for "finish" (%T)`, value)
		}
	case "start":
		if v, ok := value.([]StartMode); ok {
			c.start = v
		} else {
			return errors.Errorf(`invalid type for "start" (%T)`, value)
		}
	default:
		if c.extraFields == nil {
			c.extraFields = make(map[string]interface{})
		}
		c.extraFields[key] = value
	}
	return nil
}

func (c *Interaction) AddFinish(v ...*InteractionFinish) *Interaction {
	c.finish = append(c.finish, v...)
	return c
}

func (c *Interaction) Finish() []*InteractionFinish {
	return c.finish
}

func (c *Interaction) AddStart(v ...StartMode) *Interaction {
	c.start = append(c.start, v...)
	return c
}

func (c *Interaction) Start() []StartMode {
	return c.start
}

func (c Interaction) MarshalJSON() ([]byte, error) {
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

func (c *Interaction) UnmarshalJSON(data []byte) error {
	c.finish = nil
	c.start = nil
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
			case "finish":
				if err := dec.Decode(&(c.finish)); err != nil {
					return errors.Wrap(err, `error reading finish`)
				}
			case "start":
				if err := dec.Decode(&(c.start)); err != nil {
					return errors.Wrap(err, `error reading start`)
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

func (c *Interaction) makePairs() []*mapiter.Pair {
	var pairs []*mapiter.Pair
	if tmp := c.finish; len(tmp) > 0 {
		pairs = append(pairs, &mapiter.Pair{Key: "finish", Value: tmp})
	}
	if tmp := c.start; len(tmp) > 0 {
		pairs = append(pairs, &mapiter.Pair{Key: "start", Value: tmp})
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

func (c *Interaction) Iterate(ctx context.Context) mapiter.Iterator {
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
