package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/lestrrat-go/codegen"
	"github.com/lestrrat-go/xstrings"
	"github.com/pkg/errors"
)

type fielddef struct {
	name        string
	pubname     string
	jsonname    string
	typ         string
	allowSingle bool
	required    bool
}

type datadef struct {
	name            string
	fields          []*fielddef
	extraValidation string
	comment         string
	// sometimes spec allows to use a string instead of an object, and the string
	// is a field of the otherwise empty object. string value specifies
	// the field name
	allowString string
	clientCmd   bool
}

var types = []*datadef{
	{
		name: "AccessToken",
		fields: []*fielddef{
			{
				name:     "value",
				required: true,
				typ:      "*string",
			},
			{
				name: "bound",
				typ:  "*bool",
			},
			{
				name: "label",
				typ:  "*string",
			},
			{
				name: "manage",
				typ:  "*string",
			},
			{
				name:     "access",
				required: true,
				typ:      "[]ResourceAccess",
			},
			{
				name: "expires_in",
				typ:  "*int64",
			},
			{
				name: "key",
				typ:  "jwk.Key",
			},
			{
				name: "durable",
				typ:  "*bool",
			},
			{
				name: "split",
				typ:  "*bool",
			},
		},
	},
	{
		name: "RequestContinuation",
		fields: []*fielddef{
			{
				name:     "uri",
				pubname:  "URI",
				required: true,
				typ:      "*string",
			},
			{
				name:     "accessToken",
				jsonname: "access_token",
				required: true,
				typ:      "*AccessToken",
			},
			{
				name: "wait",
				typ:  "*int64",
			},
		},
	},
	{
		name:      "GrantRequest",
		clientCmd: true,
		extraValidation: "\nif len(c.accessTokens) > 0 {" +
			"\n  for _, access := range c.accessTokens {" +
			"\n    if access.label == nil {" +
			"\n      return errors.Errorf(`\"label\" is required in \"access\" field for multiple access token requests (2.1.1)`)" +
			"\n    }" +
			"\n  }" +
			"\n}",
		fields: []*fielddef{
			{
				name:        "accessTokens",
				jsonname:    "access_token",
				typ:         "[]*AccessTokenRequest",
				allowSingle: true,
			},
			{
				name: "capabilities",
				typ:  "[]string",
			},
			{
				name: "interact",
				typ:  "*Interaction",
			},
			{
				name: "client",
				typ:  "*Client",
			},
			{
				name: "subject",
				typ:  "*SubjectRequest",
			},
		},
	},
	{
		name: "GrantResponse",
		fields: []*fielddef{
			{
				name: "error",
				typ: "*string",
			},
		},
	},
	{
		name: "SubjectRequest",
		fields: []*fielddef{
			{
				name: "subIDs",
				typ:  "[]string",
			},
			{
				name: "assertions",
				typ:  "[]string",
			},
		},
	},
	{
		name: "UserCode",
		fields: []*fielddef{
			{
				name:     "code",
				required: true,
				typ:      "*string",
			},
			{
				name:    "url",
				pubname: "URL",
				typ:     "*string",
			},
		},
	},
	{
		name:        "Client",
		allowString: "instanceID",
		fields: []*fielddef{
			{
				name:     "instanceID",
				pubname:  "InstanceID",
				jsonname: "instance_id",
				typ:      "*string",
			},
			{
				name:     "key",
				required: true,
				typ:      "*Key",
			},
			{
				name: "classID",
				typ:  "*string",
			},
		},
	},
	{
		name: "ClientDisplay",
		fields: []*fielddef{
			{
				name: "name",
				typ:  "*string",
			},
			{
				name:    "uri",
				pubname: "URI",
				typ:     "*string",
			},
			{
				name: "logo_uri",
				typ:  "*string",
			},
		},
	},
	{
		name: "Key",
		fields: []*fielddef{
			{
				name: "proof",
				typ:  "*ProofForm",
			},
			{
				name:    "jwk",
				pubname: "JWK",
				typ:     "jwk.Key",
			},
			{
				name: "cert",
				typ:  "*string",
			},
			{
				name:     "certS256",
				jsonname: "cert#S256",
				typ:      "*string",
			},
		},
	},
	{
		name: "AccessTokenRequest",
		fields: []*fielddef{
			{
				name: "access",
				typ:  "[]*ResourceAccess",
			},
			{
				name: "label",
				typ:  "*string",
			},
			{
				name: "flags",
				typ:  "[]AccessTokenAttribute",
			},
		},
	},
	{
		name: "InteractionFinish",
		fields: []*fielddef{
			{
				name:     "method",
				required: true,
				typ:      "*FinishMode",
			},
			{
				name:     "uri",
				pubname:  "URI",
				required: true,
				typ:      "*string",
			},
			{
				name:     "nonce",
				required: true,
				typ:      "*string",
			},
			{
				name:     "hash_method",
				jsonname: "hash_method",
				typ:      "*string",
			},
		},
	},
	{
		name: "Interaction",
		fields: []*fielddef{
			{
				name:     "start",
				required: true,
				typ:      "[]StartMode",
			},
			{
				name: "finish",
				typ:  "[]*InteractionFinish",
			},
			{
				name: "hints",
				typ: "*InteractionHint",
			},
		},
	},
	{
		name: "InteractionHint",
		fields: []*fielddef{
			{
				name: "uiLocales",
				typ: "[]string",
			},
		},
	},
	{
		name:    "ResourceAccess",
		comment: "ResourceAccess describes the access, resource, and metadata associated with them",
		fields: []*fielddef{
			{
				name:    "typ",
				pubname: "Type",
				typ:     "*string",
			},
			{
				name: "actions",
				typ:  "[]string",
			},
			{
				name: "locations",
				typ:  "[]string",
			},
			{
				name:     "datatypes",
				jsonname: "datatypes", // don't snake it
				pubname:  "DataTypes",
				typ:      "[]string",
			},
			{
				name: "identifier",
				typ:  "*string",
			},
		},
	},
}

func main() {
	if err := _main(); err != nil {
		fmt.Printf("%s", err)
		os.Exit(1)
	}
}

func _main() error {
	for _, ddef := range types {
		// Preprocess each data type
		sort.Slice(ddef.fields, func(i, j int) bool {
			return ddef.fields[i].name < ddef.fields[j].name
		})
		for _, fdef := range ddef.fields {
			if fdef.pubname == "" {
				fdef.pubname = xstrings.Camel(fdef.name)
			}
			if fdef.jsonname == "" {
				fdef.jsonname = xstrings.Snake(fdef.pubname)
			}
		}

		if err := genType(ddef); err != nil {
			return errors.Wrapf(err, `failed to generate definitions for %s`, ddef.name)
		}

		if err := genClientCmd(ddef); err != nil {
			return errors.Wrapf(err, `failed to generate definitions for %s`, ddef.name)
		}
	}
	return nil
}

// type name, true if we need to take the pointer of the value
func intype(v string) (string, bool) {
	switch v {
	case "*string", "*FinishMode":
		return strings.TrimPrefix(v, "*"), true
	}
	return v, false
}

func zeroval(v string) string {
	switch v {
	case "string", "FinishMode":
		return `""`
	}
	return "nil"
}

func qualifyPkg(v string) string {
	nptr := strings.TrimPrefix(v, "*")
	var isPtr bool
	if nptr != v {
		isPtr = true
	}
	for _, ddef := range types {
		if ddef.name == nptr {
			ret := "gnap." + nptr
			if isPtr {
				ret = "*" + ret
			}
			return ret
		}
	}
	return v
}

func genClientCmd(ddef *datadef) error {
	if !ddef.clientCmd {
		return nil
	}

	var buf bytes.Buffer

	fmt.Fprintf(&buf, "\npackage client")
	fmt.Fprintf(&buf, "\n\ntype %sCmd struct {", ddef.name)
	fmt.Fprintf(&buf, "\nclient *Client")
	fmt.Fprintf(&buf, "\npayload *gnap.%s", ddef.name)
	fmt.Fprintf(&buf, "\n}")

	var requiredFields []*fielddef
	for _, field := range ddef.fields {
		if field.required {
			requiredFields = append(requiredFields, field)
		}
	}

	fmt.Fprintf(&buf, "\n\nfunc (client *Client) New%s(", ddef.name)
	for i, field := range requiredFields {
		if i > 0 {
			fmt.Fprintf(&buf, ", ")
		}
		fmt.Fprintf(&buf, "%s %s", field.name, field.typ)
	}
	fmt.Fprintf(&buf, ") *%sCmd {", ddef.name)
	fmt.Fprintf(&buf, "\nreturn &%sCmd{", ddef.name)
	fmt.Fprintf(&buf, "\nclient: client,")
	fmt.Fprintf(&buf, "\npayload: gnap.New%s(),", ddef.name)
	fmt.Fprintf(&buf, "\n}")
	fmt.Fprintf(&buf, "\n}")

	for _, fdef := range ddef.fields {
		switch {
		case strings.HasPrefix(fdef.typ, "[]"):
			fmt.Fprintf(&buf, "\n\nfunc (cmd *%[1]sCmd) Add%[2]s(v ...%[3]s) *%[1]sCmd {", ddef.name, fdef.pubname, qualifyPkg(strings.TrimPrefix(fdef.typ, "[]")))
			fmt.Fprintf(&buf, "\ncmd.payload.Add%s(v...)", fdef.pubname)
			fmt.Fprintf(&buf, "\nreturn cmd")
			fmt.Fprintf(&buf, "\n}")
		default:
			intyp, _ := intype(fdef.typ)
			fmt.Fprintf(&buf, "\n\nfunc (cmd *%[1]sCmd) %[2]s(v %[3]s) *%[1]sCmd {", ddef.name, fdef.pubname, qualifyPkg(intyp))
			fmt.Fprintf(&buf, "\ncmd.payload.Set%s(v)", fdef.pubname)
			fmt.Fprintf(&buf, "\nreturn cmd")
			fmt.Fprintf(&buf, "\n}")
		}
	}

	filename := filepath.Join("client", xstrings.Snake(ddef.name)+"_cmd_gen.go")
	if err := codegen.WriteFile(filename, &buf, codegen.WithFormatCode(true)); err != nil {
		if cfe, ok := err.(codegen.CodeFormatError); ok {
			fmt.Fprint(os.Stderr, cfe.Source())
		}
		return errors.Wrapf(err, `failed to write to %s`, filename)
	}

	return nil
}

func genType(ddef *datadef) error {
	var buf bytes.Buffer

	var requiredFields []*fielddef
	for _, field := range ddef.fields {
		if field.required {
			requiredFields = append(requiredFields, field)
		}
	}

	fmt.Fprintf(&buf, "package gnap")
	fmt.Fprintf(&buf, "\n\n")
	if ddef.comment != "" {
		fmt.Fprintf(&buf, "// %s\n", ddef.comment)
	}
	fmt.Fprintf(&buf, "type %s struct {", ddef.name)
	for _, fdef := range ddef.fields {
		fmt.Fprintf(&buf, "\n%s %s", fdef.name, fdef.typ)
	}
	fmt.Fprintf(&buf, "\nextraFields map[string]interface{}")
	fmt.Fprintf(&buf, "\n}")

	fmt.Fprintf(&buf, "\n\nfunc New%s(", ddef.name)
	for i, field := range requiredFields {
		if i > 0 {
			fmt.Fprintf(&buf, ", ")
		}

		intyp := field.typ
		// If this is a pointer, accept a non-pointer version
		intyp = strings.TrimPrefix(intyp, "*")

		// If this is a slice, accept 1 element in the constructor
		intyp = strings.TrimPrefix(intyp, "[]")

		fmt.Fprintf(&buf, "%s %s", field.name, intyp)
	}
	fmt.Fprintf(&buf, ") *%s {", ddef.name)
	fmt.Fprintf(&buf, "\nreturn &%s{", ddef.name)
	for _, field := range requiredFields {
		if strings.HasPrefix(field.typ, "[]") {
			// If this is a slice, assign as a such
			fmt.Fprintf(&buf, "\n%[1]s: %[2]s{%[1]s},", field.name, field.typ)
		} else if strings.HasPrefix(field.typ, "*") {
			// If this is a slice, assign as a such
			fmt.Fprintf(&buf, "\n%[1]s: &%[1]s,", field.name)
		} else {
			fmt.Fprintf(&buf, "\n%[1]s: %[1]s,", field.name)
		}
	}
	fmt.Fprintf(&buf, "\n}")
	fmt.Fprintf(&buf, "\n}")

	fmt.Fprintf(&buf, "\n\nfunc (c *%s) Validate() error {", ddef.name)
	for _, fdef := range ddef.fields {
		if fdef.required {
			if strings.HasPrefix(fdef.typ, "[]") {
				fmt.Fprintf(&buf, "\nif len(c.%s) == 0 {", fdef.name)
			} else {
				fmt.Fprintf(&buf, "\nif c.%s == nil {", fdef.name)
			}
			fmt.Fprintf(&buf, "\nreturn errors.Errorf(`field %#v is required`)", fdef.name)
			fmt.Fprintf(&buf, "\n}")
		}
	}
	if code := ddef.extraValidation; code != "" {
		fmt.Fprintf(&buf, code)
	}
	fmt.Fprintf(&buf, "\nreturn nil")
	fmt.Fprintf(&buf, "\n}")
	fmt.Fprintf(&buf, "\n\nfunc (c *%s) Get(key string) (interface{}, bool) {", ddef.name)
	fmt.Fprintf(&buf, "\nswitch key {")
	for _, fdef := range ddef.fields {
		fmt.Fprintf(&buf, "\ncase %#v:", fdef.jsonname)
		if strings.HasPrefix(fdef.typ, "[]") {
			fmt.Fprintf(&buf, "\nif len(c.%s) == 0 {", fdef.name)
		} else {
			fmt.Fprintf(&buf, "\nif c.%s == nil {", fdef.name)
		}
		fmt.Fprintf(&buf, "\nreturn nil, false")
		fmt.Fprintf(&buf, "\n}")
		fmt.Fprintf(&buf, "\nreturn c.%s, true", fdef.name)
	}
	fmt.Fprintf(&buf, "\ndefault:")
	fmt.Fprintf(&buf, "\nif c.extraFields == nil {")
	fmt.Fprintf(&buf, "\nreturn nil, false")
	fmt.Fprintf(&buf, "\n}")
	fmt.Fprintf(&buf, "\nv, ok := c.extraFields[key]")
	fmt.Fprintf(&buf, "\nreturn v, ok")
	fmt.Fprintf(&buf, "\n}") // end switch
	fmt.Fprintf(&buf, "\n}")
	fmt.Fprintf(&buf, "\n\nfunc (c *%s) Set(key string, value interface{}) error {", ddef.name)
	fmt.Fprintf(&buf, "\nswitch key {")
	for _, fdef := range ddef.fields {
		fmt.Fprintf(&buf, "\ncase %#v:", fdef.jsonname)
		if fdef.typ == "*string" {
			fmt.Fprintf(&buf, "\nif v, ok := value.(string); ok {")
			fmt.Fprintf(&buf, "\nc.%s = &v", fdef.name)
			fmt.Fprintf(&buf, "\n} else if value == nil {")
			fmt.Fprintf(&buf, "\nc.%s = nil", fdef.name)
			fmt.Fprintf(&buf, "\n} else {")
			fmt.Fprintf(&buf, "\nreturn errors.Errorf(`invalid type for %#v (%%T)`, value)", fdef.jsonname)
			fmt.Fprintf(&buf, "\n}")
		} else {
			fmt.Fprintf(&buf, "\nif v, ok := value.(%s); ok {", fdef.typ)
			fmt.Fprintf(&buf, "\nc.%s = v", fdef.name)
			fmt.Fprintf(&buf, "\n} else {")
			fmt.Fprintf(&buf, "\nreturn errors.Errorf(`invalid type for %#v (%%T)`, value)", fdef.jsonname)
			fmt.Fprintf(&buf, "\n}")
		}
	}
	fmt.Fprintf(&buf, "\ndefault:")
	fmt.Fprintf(&buf, "\nif c.extraFields == nil {")
	fmt.Fprintf(&buf, "\nc.extraFields = make(map[string]interface{})")
	fmt.Fprintf(&buf, "\n}")
	fmt.Fprintf(&buf, "\nc.extraFields[key] = value")
	fmt.Fprintf(&buf, "\n}") // end switch
	fmt.Fprintf(&buf, "\nreturn nil")
	fmt.Fprintf(&buf, "\n}")

	// Pre-defined fields have convenience setters
	for _, fdef := range ddef.fields {
		switch {
		case strings.HasPrefix(fdef.typ, "[]"):
			fmt.Fprintf(&buf, "\n\nfunc (c *%[1]s) Add%[2]s(v ...%[3]s) *%[1]s {", ddef.name, fdef.pubname, strings.TrimPrefix(fdef.typ, "[]"))
			fmt.Fprintf(&buf, "\nc.%[1]s = append(c.%[1]s, v...)", fdef.name)
			fmt.Fprintf(&buf, "\nreturn c")
			fmt.Fprintf(&buf, "\n}")

			fmt.Fprintf(&buf, "\n\nfunc (c *%s) %s() %s {", ddef.name, fdef.pubname, fdef.typ)
			fmt.Fprintf(&buf, "\nreturn c.%s", fdef.name)
			fmt.Fprintf(&buf, "\n}")
		default:
			intyp, takePtr := intype(fdef.typ)
			zv := zeroval(intyp)

			fmt.Fprintf(&buf, "\n\nfunc (c *%s) Set%s(v %s) {", ddef.name, fdef.pubname, intyp)
			if takePtr {
				fmt.Fprintf(&buf, "\nc.%s = &v", fdef.name)
			} else {
				fmt.Fprintf(&buf, "\nc.%s = v", fdef.name)
			}
			fmt.Fprintf(&buf, "\n}")

			fmt.Fprintf(&buf, "\n\nfunc (c *%s) %s() %s {", ddef.name, fdef.pubname, intyp)
			if !takePtr {
				fmt.Fprintf(&buf, "\nreturn c.%s", fdef.name)
			} else {
				fmt.Fprintf(&buf, "\nif c.%s == nil {", fdef.name)
				fmt.Fprintf(&buf, "\nreturn %s", zv)
				fmt.Fprintf(&buf, "\n}")
				fmt.Fprintf(&buf, "\nreturn *(c.%s)", fdef.name)
			}
			fmt.Fprintf(&buf, "\n}")
		}
	}

	fmt.Fprintf(&buf, "\n\nfunc (c %s) MarshalJSON() ([]byte, error) {", ddef.name)
	fmt.Fprintf(&buf, "\nctx, cancel := context.WithCancel(context.Background())")
	fmt.Fprintf(&buf, "\ndefer cancel()")
	fmt.Fprintf(&buf, "\nvar buf bytes.Buffer")
	fmt.Fprintf(&buf, "\nenc := json.NewEncoder(&buf)")
	fmt.Fprintf(&buf, "\nbuf.WriteByte('{')")
	if fieldname := ddef.allowString; fieldname != "" {
		fmt.Fprintf(&buf, "\nvar pairs []*mapiter.Pair")
		fmt.Fprintf(&buf, "\nfor iter := c.Iterate(ctx); iter.Next(ctx); {")
		fmt.Fprintf(&buf, "\npairs = append(pairs, iter.Pair())")
		fmt.Fprintf(&buf, "\n}")
		// dangit, we can't just translate field name to json name
		var jsonname string
		for _, fdef := range ddef.fields {
			if fdef.name == fieldname {
				jsonname = fdef.jsonname
				break
			}
		}
		fmt.Fprintf(&buf, "\nif len(pairs) == 1 && pairs[0].Key.(string) == %#v {", jsonname)
		fmt.Fprintf(&buf, "\nreturn []byte(strconv.Quote(pairs[0].Value.(string))), nil")
		fmt.Fprintf(&buf, "\n}")
		fmt.Fprintf(&buf, "\nfor i, pair := range pairs {")
	} else {
		fmt.Fprintf(&buf, "\nvar i int")
		fmt.Fprintf(&buf, "\nfor iter := c.Iterate(ctx); iter.Next(ctx); {")
		fmt.Fprintf(&buf, "\npair := iter.Pair()")
	}

	fmt.Fprintf(&buf, "\nif i > 0 {")
	fmt.Fprintf(&buf, "\nbuf.WriteByte(',')")
	fmt.Fprintf(&buf, "\n}")
	if fieldname := ddef.allowString; fieldname == "" {
		fmt.Fprintf(&buf, "\ni++")
	}
	fmt.Fprintf(&buf, "\nbuf.WriteString(strconv.Quote(pair.Key.(string)))")
	fmt.Fprintf(&buf, "\nbuf.WriteByte(':')")

	var singles []*fielddef
	for _, fdef := range ddef.fields {
		if fdef.allowSingle {
			singles = append(singles, fdef)
		}
	}

	if len(singles) > 0 {
		fmt.Fprintf(&buf, "\nswitch pair.Key.(string) {")
		for _, fdef := range singles {
			fmt.Fprintf(&buf, "\ncase %#v:", fdef.jsonname)
			fmt.Fprintf(&buf, "\nv := pair.Value.(%s)", fdef.typ)
			fmt.Fprintf(&buf, "\nif len(v) == 1 {")
			fmt.Fprintf(&buf, "\npair.Value = v[0]")
			fmt.Fprintf(&buf, "\n}") // end if
		}
		fmt.Fprintf(&buf, "\n}") // end switch
	}

	fmt.Fprintf(&buf, "\nif err := enc.Encode(pair.Value); err != nil {")
	fmt.Fprintf(&buf, "\nreturn nil, errors.Wrapf(err, `failed to encode %%s`, pair.Key.(string))")
	fmt.Fprintf(&buf, "\n}") // end if
	fmt.Fprintf(&buf, "\n}") // end for
	fmt.Fprintf(&buf, "\nbuf.WriteByte('}')")
	fmt.Fprintf(&buf, "\nreturn buf.Bytes(), nil")
	fmt.Fprintf(&buf, "\n}")

	fmt.Fprintf(&buf, "\n\nfunc (c *%s) UnmarshalJSON(data []byte) error {", ddef.name)
	for _, fdef := range ddef.fields {
		switch fdef.typ {
		case "string":
			fmt.Fprintf(&buf, "\nc.%s = \"\"", fdef.name)
		default:
			fmt.Fprintf(&buf, "\nc.%s = nil", fdef.name)
		}
	}

	fmt.Fprintf(&buf, "\ndec := json.NewDecoder(bytes.NewReader(data))")
	fmt.Fprintf(&buf, "\ntok, err := dec.Token()")
	fmt.Fprintf(&buf, "\nif err != nil {")
	fmt.Fprintf(&buf, "\nreturn errors.Wrap(err, `error reading token`)")
	fmt.Fprintf(&buf, "\n}")
	fmt.Fprintf(&buf, "\nswitch tok := tok.(type) {")
	fmt.Fprintf(&buf, "\ncase json.Delim:")
	fmt.Fprintf(&buf, "\nif tok != '{' { ")
	fmt.Fprintf(&buf, "\nreturn errors.Errorf(`expected '{', but got '%%c'`, tok)")
	fmt.Fprintf(&buf, "\n}")
	if fieldname := ddef.allowString; fieldname != "" {
		fmt.Fprintf(&buf, "\ncase string:")
		fmt.Fprintf(&buf, "\nc.%s = &tok", fieldname)
		fmt.Fprintf(&buf, "\nreturn nil")
		fmt.Fprintf(&buf, "\ndefault:")
		fmt.Fprintf(&buf, "\nreturn errors.Errorf(`expected '{' or string, but got '%%c'`, tok)")
		fmt.Fprintf(&buf, "\n}")
	} else {
		fmt.Fprintf(&buf, "\ndefault:")
		fmt.Fprintf(&buf, "\nreturn errors.Errorf(`expected '{', but got '%%c'`, tok)")
		fmt.Fprintf(&buf, "\n}")
	}

	fmt.Fprintf(&buf, "\nLOOP:")
	fmt.Fprintf(&buf, "\nfor {")
	fmt.Fprintf(&buf, "\ntok, err := dec.Token()")
	fmt.Fprintf(&buf, "\nif err != nil {")
	fmt.Fprintf(&buf, "\nreturn errors.Wrap(err, `error reading token`)")
	fmt.Fprintf(&buf, "\n}")
	fmt.Fprintf(&buf, "\nswitch tok := tok.(type) {")
	fmt.Fprintf(&buf, "\ncase json.Delim:")
	fmt.Fprintf(&buf, "\nif tok == '}' { // End of object")
	fmt.Fprintf(&buf, "\nbreak LOOP")
	fmt.Fprintf(&buf, "\n}")
	fmt.Fprintf(&buf, "\nreturn errors.Errorf(`unexpected delimiter '%%c'`, tok)")
	fmt.Fprintf(&buf, "\ncase string:")
	fmt.Fprintf(&buf, "\nswitch tok {")
	for _, fdef := range ddef.fields {
		fmt.Fprintf(&buf, "\ncase %s:", strconv.Quote(fdef.jsonname))
		switch fdef.typ {
		case "*string":
			fmt.Fprintf(&buf, "\nvar tmp string")
			fmt.Fprintf(&buf, "\nif err := dec.Decode(&tmp); err != nil {")
			fmt.Fprintf(&buf, "\nreturn errors.Wrap(err, `error reading %s`)", fdef.jsonname)
			fmt.Fprintf(&buf, "\n}")
			fmt.Fprintf(&buf, "\nc.%s = &tmp", fdef.name)
		default:
			if !strings.HasPrefix(fdef.typ, "[]") || !fdef.allowSingle {
				fmt.Fprintf(&buf, "\nif err := dec.Decode(&(c.%s)); err != nil {", fdef.name)
				fmt.Fprintf(&buf, "\nreturn errors.Wrap(err, `error reading %s`)", fdef.jsonname)
				fmt.Fprintf(&buf, "\n}")
			} else {
				// Decode the next thing into a json.RawMessage, and then do heuristics
				fmt.Fprintf(&buf, "\nvar nextThing json.RawMessage")
				fmt.Fprintf(&buf, "\nif err := dec.Decode(&nextThing); err != nil {")
				fmt.Fprintf(&buf, "\nreturn errors.Wrap(err, `error reading next token %s`)", fdef.jsonname)
				fmt.Fprintf(&buf, "\n}")
				fmt.Fprintf(&buf, "\nif bytes.HasPrefix(nextThing, []byte{'['}) {")
				fmt.Fprintf(&buf, "\nif err := json.Unmarshal(nextThing, &(c.%s)); err != nil {", fdef.name)
				fmt.Fprintf(&buf, "\nreturn errors.Wrap(err, `error decoding %s`)", fdef.jsonname)
				fmt.Fprintf(&buf, "\n}")
				fmt.Fprintf(&buf, "\n} else {")

				// either []thing or []*thing. if []*thing, remember we want pointers at the end
				typ := strings.TrimPrefix(fdef.typ, "[]")
				var ptr bool
				if strings.HasPrefix(typ, "*") {
					typ = strings.TrimPrefix(typ, "*")
					ptr = true
				}
				fmt.Fprintf(&buf, "\nvar tmp %s", typ)
				fmt.Fprintf(&buf, "\nif err := json.Unmarshal(nextThing, &tmp); err != nil {")
				fmt.Fprintf(&buf, "\nreturn errors.Wrap(err, `error reading %s`)", fdef.jsonname)
				fmt.Fprintf(&buf, "\n}")
				fmt.Fprintf(&buf, "\nc.%[1]s = append(c.%[1]s, ", fdef.name)
				if ptr {
					fmt.Fprintf(&buf, "&")
				}
				fmt.Fprintf(&buf, "tmp)")
				fmt.Fprintf(&buf, "\n}") // end else
			}
		}
	}
	fmt.Fprintf(&buf, "\ndefault:")
	fmt.Fprintf(&buf, "\nvar tmp interface{}")
	fmt.Fprintf(&buf, "\nif err := dec.Decode(&tmp); err != nil {")
	fmt.Fprintf(&buf, "\nreturn errors.Wrapf(err, `error reading %%s`, tok)")
	fmt.Fprintf(&buf, "\n}")
	fmt.Fprintf(&buf, "\nif c.extraFields == nil {")
	fmt.Fprintf(&buf, "\nc.extraFields = map[string]interface{}{}")
	fmt.Fprintf(&buf, "\n}")
	fmt.Fprintf(&buf, "\nc.extraFields[tok] = tmp")
	fmt.Fprintf(&buf, "\n}") // end switch (inside case string)
	fmt.Fprintf(&buf, "\n}") // end switch
	fmt.Fprintf(&buf, "\n}") // end for
	fmt.Fprintf(&buf, "\nreturn nil")
	fmt.Fprintf(&buf, "\n}") // end method

	fmt.Fprintf(&buf, "\n\nfunc (c *%s) makePairs() []*mapiter.Pair {", ddef.name)
	fmt.Fprintf(&buf, "\nvar pairs []*mapiter.Pair")
	for _, fdef := range ddef.fields {
		switch {
		case strings.HasPrefix(fdef.typ, "*"):
			fmt.Fprintf(&buf, "\nif tmp := c.%s; tmp != nil {", fdef.name)
			fmt.Fprintf(&buf, "\npairs = append(pairs, &mapiter.Pair{Key: %#v, Value: *tmp})", fdef.jsonname)
			fmt.Fprintf(&buf, "\n}")
		case strings.HasPrefix(fdef.typ, "[]"):
			fmt.Fprintf(&buf, "\nif tmp := c.%s; len(tmp) > 0 {", fdef.name)
			fmt.Fprintf(&buf, "\npairs = append(pairs, &mapiter.Pair{Key: %#v, Value: tmp})", fdef.jsonname)
			fmt.Fprintf(&buf, "\n}")
		case fdef.typ == "string":
			fmt.Fprintf(&buf, "\nif tmp := c.%s; tmp != \"\" {", fdef.name)
			fmt.Fprintf(&buf, "\npairs = append(pairs, &mapiter.Pair{Key: %#v, Value: tmp})", fdef.jsonname)
			fmt.Fprintf(&buf, "\n}")
		}
	}
	fmt.Fprintf(&buf, "\nvar extraKeys []string")
	fmt.Fprintf(&buf, "\nfor k := range c.extraFields {")
	fmt.Fprintf(&buf, "\nextraKeys = append(extraKeys, k)")
	fmt.Fprintf(&buf, "\n}")
	fmt.Fprintf(&buf, "\nfor _, k := range extraKeys {")
	fmt.Fprintf(&buf, "\npairs = append(pairs, &mapiter.Pair{Key: k, Value: c.extraFields[k]})")
	fmt.Fprintf(&buf, "\n}")
	fmt.Fprintf(&buf, "\nsort.Slice(pairs, func(i, j int) bool {")
	fmt.Fprintf(&buf, "\nreturn pairs[i].Key.(string) < pairs[j].Key.(string)")
	fmt.Fprintf(&buf, "\n})")
	fmt.Fprintf(&buf, "\nreturn pairs")
	fmt.Fprintf(&buf, "\n}")

	fmt.Fprintf(&buf, "\n\nfunc (c *%s) Iterate(ctx context.Context) mapiter.Iterator {", ddef.name)
	fmt.Fprintf(&buf, "\npairs := c.makePairs()")
	fmt.Fprintf(&buf, "\nch := make(chan *mapiter.Pair, len(pairs))")
	fmt.Fprintf(&buf, "\ngo func(ctx context.Context, ch chan *mapiter.Pair, pairs []*mapiter.Pair) {")
	fmt.Fprintf(&buf, "\ndefer close(ch)")
	fmt.Fprintf(&buf, "\nfor _, pair := range pairs {")
	fmt.Fprintf(&buf, "\nselect {")
	fmt.Fprintf(&buf, "\ncase <-ctx.Done():")
	fmt.Fprintf(&buf, "\nreturn")
	fmt.Fprintf(&buf, "\ncase ch <- pair:")
	fmt.Fprintf(&buf, "\n}") // end select
	fmt.Fprintf(&buf, "\n}") // end for
	fmt.Fprintf(&buf, "\n}(ctx, ch, pairs)")
	fmt.Fprintf(&buf, "\nreturn mapiter.New(ch)")
	fmt.Fprintf(&buf, "\n}") // end func

	filename := xstrings.Snake(ddef.name) + "_gen.go"
	if err := codegen.WriteFile(filename, &buf, codegen.WithFormatCode(true)); err != nil {
		if cfe, ok := err.(codegen.CodeFormatError); ok {
			fmt.Fprint(os.Stderr, cfe.Source())
		}
		return errors.Wrapf(err, `failed to write to %s`, filename)
	}
	return nil
}
