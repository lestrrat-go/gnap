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

type InteractionRequest struct {
	finish      []*InteractionFinish
	hints       *InteractionHint
	start       []StartMode
	extraFields map[string]interface{}
}

func NewInteractionRequest(start StartMode) *InteractionRequest {
	return &InteractionRequest{
		start: []StartMode{start},
	}
}

func (c *InteractionRequest) Validate() error {
	if len(c.start) == 0 {
		return errors.Errorf(`field "start" is required`)
	}
	return nil
}

func (c *InteractionRequest) Get(key string) (interface{}, bool) {
	switch key {
	case "finish":
		if len(c.finish) == 0 {
			return nil, false
		}
		return c.finish, true
	case "hints":
		if c.hints == nil {
			return nil, false
		}
		return c.hints, true
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

func (c *InteractionRequest) Set(key string, value interface{}) error {
	switch key {
	case "finish":
		if v, ok := value.([]*InteractionFinish); ok {
			c.finish = v
		} else {
			return errors.Errorf(`invalid type for "finish" (%T)`, value)
		}
	case "hints":
		if v, ok := value.(*InteractionHint); ok {
			c.hints = v
		} else {
			return errors.Errorf(`invalid type for "hints" (%T)`, value)
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

func (c *InteractionRequest) AddFinish(v ...*InteractionFinish) *InteractionRequest {
	c.finish = append(c.finish, v...)
	return c
}

func (c *InteractionRequest) Finish() []*InteractionFinish {
	return c.finish
}

func (c *InteractionRequest) SetHints(v *InteractionHint) {
	c.hints = v
}

func (c *InteractionRequest) Hints() *InteractionHint {
	return c.hints
}

func (c *InteractionRequest) AddStart(v ...StartMode) *InteractionRequest {
	c.start = append(c.start, v...)
	return c
}

func (c *InteractionRequest) Start() []StartMode {
	return c.start
}

func (c InteractionRequest) MarshalJSON() ([]byte, error) {
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

func (c *InteractionRequest) UnmarshalJSON(data []byte) error {
	c.finish = nil
	c.hints = nil
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
			case "hints":
				if err := dec.Decode(&(c.hints)); err != nil {
					return errors.Wrap(err, `error reading hints`)
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

func (c *InteractionRequest) makePairs() []*mapiter.Pair {
	var pairs []*mapiter.Pair
	if tmp := c.finish; len(tmp) > 0 {
		pairs = append(pairs, &mapiter.Pair{Key: "finish", Value: tmp})
	}
	if tmp := c.hints; tmp != nil {
		pairs = append(pairs, &mapiter.Pair{Key: "hints", Value: *tmp})
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

func (c *InteractionRequest) Iterate(ctx context.Context) mapiter.Iterator {
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
