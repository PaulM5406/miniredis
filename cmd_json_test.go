package miniredis

import (
	"testing"

	"github.com/alicebob/miniredis/v2/proto"
)

// Test JSON.SET - JSON.GET
func TestJson(t *testing.T) {
	s, err := Run()
	ok(t, err)
	defer s.Close()
	c, err := proto.Dial(s.Addr())
	ok(t, err)
	defer c.Close()

	t.Run("key does not exist", func(t *testing.T) {
		mustNil(t, c,
			"JSON.GET", "unknown_key", "$",
		)

		mustNil(t, c,
			"JSON.GET", "unknown_key",
		)
	})

	t.Run("basic root paths", func(t *testing.T) {
		mustOK(t, c,
			"JSON.SET", "basic", "$", `"silence"`,
		)

		mustDo(t, c,
			"JSON.GET", "basic", "$",
			proto.Inline(`["silence"]`),
		)

		mustDo(t, c,
			"JSON.GET", "basic",
			proto.Inline(`"silence"`),
		)

		mustOK(t, c,
			"JSON.SET", "basic", "$", `"noisy"`,
		)

		mustDo(t, c,
			"JSON.GET", "basic", "$",
			proto.Inline(`["noisy"]`),
		)

		mustDo(t, c,
			"JSON.GET", "basic",
			proto.Inline(`"noisy"`),
		)

		mustDo(t, c,
			"JSON.GET", "basic", "$", "$",
			proto.Inline(`{"$":["noisy"]}`),
		)
	})

	t.Run("not root paths", func(t *testing.T) {
		mustOK(t, c,
			"JSON.SET", "doc", "$", `{"a":2, "b": 2}`,
		)

		mustDo(t, c,
			"JSON.GET", "doc", "$.a",
			proto.Inline("[2]"),
		)

		// mustDo(t, c,
		// 	"JSON.GET", "doc", ".a",
		// 	proto.Inline("2"),
		// )

		// mustDo(t, c,
		// 	"JSON.GET", "doc", "a",
		// 	proto.Inline("2"),
		// )

		mustDo(t, c,
			"JSON.GET", "doc", "$..a",
			proto.Inline("[2]"),
		)

		// mustDo(t, c,
		// 	"JSON.GET", "doc", "..a",
		// 	proto.Inline("2"),
		// )

		mustDo(t, c,
			"JSON.GET", "doc", "$.a", "$.a",
			proto.Inline(`{"$.a":[2]}`),
		)

		// mustDo(t, c,
		// 	"JSON.GET", "doc", ".a", ".a",
		// 	proto.Inline(`{"$.a": 2}`),
		// )

		// mustDo(t, c,
		// 	"JSON.GET", "doc", "a", "a",
		// 	proto.Inline(`{"$.a": 2}`),
		// )

		// mustDo(t, c,
		// 	"JSON.GET", "doc", "$.a", ".a", "a",
		// 	proto.Inline(`{"$.a": [2], ".a": [2], "a": [2]}`),
		// )
	})

	t.Run("path does not exist", func(t *testing.T) {
		mustOK(t, c,
			"JSON.SET", "ex", "$", `{"a":2, "b": 2}`,
		)

		mustDo(t, c,
			"JSON.GET", "ex", "$.c",
			proto.Inline("[]"),
		)

		mustDo(t, c,
			"JSON.GET", "ex", "$.c", "$.c",
			proto.Inline(`{"$.c":[]}`),
		)

		// mustDo(t, c,
		// 	"JSON.GET", "doc", ".c",
		// 	proto.Error()
		// )

		// mustDo(t, c,
		// 	"JSON.GET", "doc", ".c", ".c",
		// 	proto.Error()
		// )
	})

	t.Run("more complex paths", func(t *testing.T) {

	})

	t.Run("for new key, the path must be the root", func(t *testing.T) {
		mustDo(t, c,
			"JSON.SET", "newKey", "$.a", `"silence"`,
			proto.Error(msgRootToCreateObject),
		)
	})

	t.Run("option NX, sets the key only if it does not already exist", func(t *testing.T) {
		mustOK(t, c,
			"JSON.SET", "nx", "$", `"silence"`, "NX",
		)

		mustDo(t, c,
			"JSON.GET", "nx", "$",
			proto.Inline(`["silence"]`),
		)

		mustNil(t, c,
			"JSON.SET", "nx", "$", `"silence"`, "NX",
		)
	})

	t.Run("option XX, sets the key only if it already exists", func(t *testing.T) {
		mustNil(t, c,
			"JSON.SET", "xx", "$", `"silence"`, "XX",
		)

		mustOK(t, c,
			"JSON.SET", "xx", "$", `"silence"`,
		)

		mustOK(t, c,
			"JSON.SET", "xx", "$", `"noisy"`, "XX",
		)
	})

	t.Run("as long as it starts with a valid json", func(t *testing.T) {
		mustOK(t, c,
			"JSON.SET", "key", "$", `{}[`,
		)

		mustDo(t, c,
			"JSON.GET", "key", "$",
			proto.Inline(`[{}]`),
		)
	})

	t.Run("errors related to invalid json input", func(t *testing.T) {
		mustDo(t, c,
			"JSON.SET", "key", "$", `{`,
			proto.Error(msgInvalidJson),
		)
	})

	t.Run("errors related to incorrect inputs", func(t *testing.T) {
		mustDo(t, c,
			"JSON.SET",
			proto.Error(errWrongNumber("JSON.SET")),
		)

		mustDo(t, c,
			"JSON.SET", "key",
			proto.Error(errWrongNumber("JSON.SET")),
		)

		mustDo(t, c,
			"JSON.SET", "key", "$",
			proto.Error(errWrongNumber("JSON.SET")),
		)

		mustDo(t, c,
			"JSON.SET", "key", "$", `"silence"`, "YY",
			proto.Error(msgSyntaxError),
		)

		mustDo(t, c,
			"JSON.SET", "key", "$", `"silence"`, "XX", "TooMany",
			proto.Error(msgSyntaxError),
		)

		mustDo(t, c,
			"JSON.GET",
			proto.Error(errWrongNumber("JSON.GET")),
		)
	})

}
