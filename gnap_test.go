package gnap_test

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/lestrrat-go/gnap"
	"github.com/stretchr/testify/assert"
)

func datatypeRoundtrip(t *testing.T, src string, expected interface{}) bool {
	t.Helper()

	dst := reflect.New(reflect.TypeOf(expected).Elem()).Interface()

	if !assert.NoError(t, json.Unmarshal([]byte(src), dst), `json.Unmarshal should succeed`) {
		return false
	}

	if !assert.Equal(t, expected, dst, `values should match`) {
		return false
	}

	buf, err := json.Marshal(dst)
	if !assert.NoError(t, err, `json.Marshal should succeed`) {
		return false
	}

	if !assert.Equal(t, src, string(buf), `produced JSON should match`) {
		return false
	}
	return true
}

func TestDataTypes(t *testing.T) {
	t.Run("GrantRequest", func(t *testing.T) {
		t.Run("Single Access Token", func(t *testing.T) {
			const src = `   {
       "access_token": {
           "access": [
               {
                   "type": "photo-api",
                   "actions": [
                       "read",
                       "write",
                       "dolphin"
                   ],
                   "locations": [
                       "https://server.example.net/",
                       "https://resource.local/other"
                   ],
                   "datatypes": [
                       "metadata",
                       "images"
                   ]
               }
           ]
       }
		}`
			var expected gnap.GrantRequest
			if !assert.NoError(t, json.Unmarshal([]byte(src), &expected), `json.Unmarshal should succeed`) {
				return
			}
		})
	})
	t.Run("AccessTokenRequest", func(t *testing.T) {
		const src = `{"access":[{"actions":["read","write","delete"],"datatypes":["metadata","images"],"locations":["https://server.example.net/","https://resource.local/other"],"type":"photo-api"},{"actions":["foo","bar"],"datatypes":["data","pictures","walrus whiskers"],"locations":["https://resource.other/"],"type":"walrus-access"}],"flags":["split"],"label":"token1-23"}`
		var expected gnap.AccessTokenRequest
		expected.SetLabel("token1-23")
		expected.AddFlags(gnap.Split)

		var ra1 gnap.ResourceAccess
		ra1.SetType("photo-api")
		ra1.AddActions("read", "write", "delete")
		ra1.AddLocations("https://server.example.net/", "https://resource.local/other")
		ra1.AddDataTypes("metadata", "images")
		expected.AddAccess(&ra1)

		var ra2 gnap.ResourceAccess
		ra2.SetType("walrus-access")
		ra2.AddActions("foo", "bar")
		ra2.AddLocations("https://resource.other/")
		ra2.AddDataTypes("data", "pictures", "walrus whiskers")
		expected.AddAccess(&ra2)

		t.Run("Roundtrip", func(t *testing.T) {
			datatypeRoundtrip(t, src, &expected)
		})
	})
	t.Run("InteractionFinish", func(t *testing.T) {
		const src = `{"method":"redirect","nonce":"LKLTI25DK82FX4T4QFZC","uri":"https://client.example.net/return/123455"}`

		var expected gnap.InteractionFinish
		expected.SetMethod(gnap.FinishRedirect)
		expected.SetURI("https://client.example.net/return/123455")
		expected.SetNonce("LKLTI25DK82FX4T4QFZC")

		t.Run("Roundtrip", func(t *testing.T) {
			datatypeRoundtrip(t, src, &expected)
		})
	})
	t.Run("ResourceAccess", func(t *testing.T) {
		const src = `{"actions":["read"],"datatypes":["file"],"extra":"foo","identifier":"gnap.go","locations":["https://github.com/lestrrat-go/gnap"],"type":"sourcecode"}`

		var expected gnap.ResourceAccess
		expected.SetType("sourcecode")
		expected.AddActions("read")
		expected.AddLocations("https://github.com/lestrrat-go/gnap")
		expected.AddDataTypes("file")

		//nolint:errcheck
		expected.SetIdentifier("gnap.go")
		//nolint:errcheck
		expected.Set("extra", "foo")

		t.Run("Roundtrip", func(t *testing.T) {
			datatypeRoundtrip(t, src, &expected)
		})
	})
}
