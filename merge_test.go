package yaml

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"go.yaml.in/yaml/v3"
)

// From http://yaml.org/type/merge.html
var mergeTests = `
anchors:
  list:
    - &CENTER { "x": 1, "y": 2 }
    - &LEFT   { "x": 0, "y": 2 }
    - &BIG    { "r": 10 }
    - &SMALL  { "r": 1 }

# All the following maps are equal:

plain:
  # Explicit keys
  "x": 1
  "y": 2
  "r": 10
  label: center/big

mergeOne:
  # Merge one map
  << : *CENTER
  "r": 10
  label: center/big

mergeMultiple:
  # Merge multiple maps
  << : [ *CENTER, *BIG ]
  label: center/big

override:
  # Override
  << : [ *BIG, *LEFT, *SMALL ]
  "x": 1
  label: center/big

shortTag:
  # Explicit short merge tag
  !!merge "<<" : [ *CENTER, *BIG ]
  label: center/big

longTag:
  # Explicit merge long tag
  !<tag:yaml.org,2002:merge> "<<" : [ *CENTER, *BIG ]
  label: center/big

inlineMap:
  # Inlined map
  << : {"x": 1, "y": 2, "r": 10}
  label: center/big

inlineSequenceMap:
  # Inlined map in sequence
  << : [ *CENTER, {"r": 10} ]
  label: center/big
`

func TestMergeYAMLMapNodes(t *testing.T) {
	tests := []struct {
		name     string
		dst      string
		src      string
		expected string
	}{
		{
			name:     "scalar values",
			dst:      `"original"`,
			src:      `"updated"`,
			expected: `"updated"`,
		},
		{
			name:     "empty to empty",
			dst:      "",
			src:      "",
			expected: "",
		},
		{
			name:     "merge to empty",
			dst:      "",
			src:      "foo: bar",
			expected: "foo: bar",
		},
		{
			name:     "empty to something",
			dst:      "foo: bar",
			src:      "",
			expected: "foo: bar",
		},
		{
			name:     "empty({}) to something",
			dst:      "foo: bar",
			src:      "{}",
			expected: "foo: bar",
		},
		{
			name: "mapping nodes - simple merge to empty",
			dst:  "",
			src: `
key2: updated_value2
key3: value3
`,
			expected: `
key2: updated_value2
key3: value3
`,
		},
		{
			name: "mapping nodes - simple merge",
			dst: `
key1: value1
key2: value2
`,
			src: `
key2: updated_value2
key3: value3
`,
			expected: `
key1: value1
key2: updated_value2
key3: value3
`,
		},
		{
			name: "mapping nodes - nested merge",
			dst: `
nested:
  key1: value1
  key2: value2
`,
			src: `
nested:
  key2: updated_value2
  key3: value3
`,
			expected: `
nested:
  key1: value1
  key2: updated_value2
  key3: value3
`,
		},
		{
			name: "sequence nodes - replaced",
			dst: `
- item1
- item2
`,
			src: `
- item3
- item4
`,
			expected: `
- item3
- item4
`,
		},
		{
			name:     "different node kinds - does nothing with error",
			dst:      `"scalar"`,
			src:      `[sequence]`,
			expected: `error: invalid node kinds`,
		},
		{
			name: "empty destination {} - if dst is empty replace with src",
			dst:  `{}`,
			src: `
key1: value1
key2: value2
`,
			expected: "{key1: value1, key2: value2}\n",
		},
		{
			name: "empty destination '' - if dst is empty replace with src",
			dst:  ``,
			src: `
key1: value1
key2: value2
`,
			expected: `
key1: value1
key2: value2
`,
		},
		{
			name: "empty source - if src is empty do nothing",
			dst: `
key1: value1
key2: value2
`,
			src: `{}`,
			expected: `
key1: value1
key2: value2
`,
		},
		{
			name: "single docs",
			dst: `
---
key1: value1
key2: value2
`,
			src: `
---
key2: updated_value2
key3: value3
`,
			expected: `
key1: value1
key2: updated_value2
key3: value3
`,
		},
		{
			name: "multiple docs",
			dst: `
---
key1: value1
key2: value2
---
key3: value3
key4: value4
`,
			src: `
---
key2: updated_value2
key3: value3
`,
			expected: `
key1: value1
key2: updated_value2
key3: value3
`,
		},
		{
			name: "merge test from yaml spec",
			dst:  mergeTests,
			src:  mergeTests,
			expected: `
anchors:
    list:
        - &CENTER {"x": 1, "y": 2}
        - &LEFT {"x": 0, "y": 2}
        - &BIG {"r": 10}
        - &SMALL {"r": 1}
# All the following maps are equal:
plain:
    # Explicit keys
    "x": 1
    "y": 2
    "r": 10
    label: center/big
mergeOne:
    # Merge one map
    !!merge <<: *CENTER
    "r": 10
    label: center/big
mergeMultiple:
    # Merge multiple maps
    !!merge <<: [*CENTER, *BIG]
    label: center/big
override:
    # Override
    !!merge <<: [*BIG, *LEFT, *SMALL]
    "x": 1
    label: center/big
shortTag:
    # Explicit short merge tag
    !!merge "<<": [*CENTER, *BIG]
    label: center/big
longTag:
    # Explicit merge long tag
    !!merge "<<": [*CENTER, *BIG]
    label: center/big
inlineMap:
    # Inlined map
    !!merge <<: {"x": 1, "y": 2, "r": 10}
    label: center/big
inlineSequenceMap:
    # Inlined map in sequence
    !!merge <<: [*CENTER, {"r": 10}]
    label: center/big
`,
		},
		{
			name: "merge test 2 and lists got replaced",
			dst: `
anchors:
    list:
        - &BIG {"r": 10}
mergeOne:
    !!merge <<: *BIG
    "a": 1
    "b": 2
`,
			src: `
anchors:
    list:
        - &SMALL {"r": 1}
mergeOne:
    !!merge <<: *SMALL
    "b": 22
    "c": 33
`,
			expected: `
anchors:
    list:
        - &SMALL {"r": 1}
mergeOne:
    !!merge <<: *SMALL
    "a": 1
    "b": 22
    "c": 33
`,
		},
		{
			name: "merge test 3 - extend anchor",
			dst: `
# 1
anchors:
    # 2
    list:
        &BIG {"a": 10}
    # 3
# 4
mergeOne:
    !!merge <<: *BIG
    "a": 1
    "b": 2
# 5
`,
			src: `
# 1x
anchors:
    # 2x
    list:
        &BIG {"b": 10}
    # 3x
# 4x
mergeOne:
    !!merge <<: *BIG
    "b": 22
    "c": 33
# 5x
`,
			expected: `
# 1x
anchors:
    # 2x
    list:
        &BIG {"a": 10, "b": 10}
    # 3x
# 4x
mergeOne:
    !!merge <<: *BIG
    "a": 1
    "b": 22
    "c": 33
# 5x
`,
		},
		{
			name: "merge test 4 - anchor collision",
			dst: `
inner: &I1
    a: 1

outter:
    in: *I1
`,
			src: `
inner: &I2
    b: 2

outter:
    in: *I2
`,
			// I2 is not allowed to replace I1.
			expected: "error: unmergable error",
		},
		{
			name: "empty keys",
			dst: `
foo:
    bar:
x:
`,
			src: `
foo:
    zoo:
`,
			expected: `
foo:
    bar:
    zoo:
x:
`,
		},
		{
			name: "map onto empty map that is null scalar",
			dst: `
foo:
`,
			src: `
foo:
    zoo:
`,
			expected: `
foo:
    zoo:
`,
		},
		{
			name: "map onto scalar",
			dst: `
foo: 1
`,
			src: `
foo:
    zoo:
`,
			expected: "error: invalid node kinds",
		},
		{
			name: "null scalar onto map",
			dst: `
foo:
    zoo:
`,
			src: `
foo:
`,
			expected: `
foo:
`,
		},
		{
			name: "nested null scalar onto null scalar",
			dst: `
foo:
    zoo:
`,
			src: `
foo:
    zoo:
`,
			expected: `
foo:
    zoo:
`,
		},
		{
			name: "list to map",
			dst: `
foo:
    zoo:
`,
			src: `
foo:
    - blah
`,
			expected: "error: invalid node kinds",
		},
		{
			name: "set list to null scalar",
			dst: `
foo:
`,
			src: `
foo:
    - x
    - y
    - z
`,
			expected: `
foo:
    - x
    - y
    - z
`,
		},
		{
			name: "erase list",
			dst: `
foo:
    - x
    - y
    - z
`,
			src: `
foo:
`,
			expected: `
foo:
`,
		},
		{
			name: "set scalar",
			dst: `
foo:
`,
			src: `
foo: 1
`,
			expected: `
foo: 1
`,
		},
		{
			name: "erase scalar",
			dst: `
foo: 1
`,
			src: `
foo:
`,
			expected: `
foo:
`,
		},
		{
			name: "null nodes",
			dst: `
a: 1
b:
  - a
foo:
  a: 1
c: null
d: null
e: null
f: null
g: null
`,
			src: `
a: null
b: null
foo: null
c: null
d: 1
e:
f:
  a: 1
g:
- a
`,
			expected: `
a: null
b: null
foo: null
c: null
d: 1
e:
f:
  a: 1
g:
- a
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var dst, src, expected yaml.Node

			NoError(t, yaml.Unmarshal([]byte(tt.dst), &dst))
			NoError(t, yaml.Unmarshal([]byte(tt.src), &src))
			NoError(t, yaml.Unmarshal([]byte(tt.expected), &expected))

			// fmt.Println(">>", (*DebugYamlNode)(&dst).String())
			// fmt.Println(">>", (*DebugYamlNode)(&src).String())

			err := MergeYAMLNodes(&dst, &src)

			// fmt.Println(">>", (*DebugYamlNode)(&dst).String())

			if strings.HasPrefix(tt.expected, "error: ") {
				ErrorContains(t, err, tt.expected[len("error: "):])
				return
			} else {
				NoError(t, err)
			}

			dstBytes, err := yaml.Marshal(&dst)
			NoError(t, err)

			expectedBytes, err := yaml.Marshal(&expected)
			NoError(t, err)

			if string(expectedBytes) != string(dstBytes) {
				Fail(t, errors.New("unexpected result"), "%s\n!=\n\n%s", string(expectedBytes), string(dstBytes))
			}
		})
	}
}

func TestFilterYAMLNullNodes(t *testing.T) {
	tests := []struct {
		name     string
		original string
		explicit string
		implicit string
		both     string
	}{
		{
			name:     "null scalar",
			original: "null",
			explicit: "",
			implicit: "",
			both:     "",
		},
		{
			name:     "to empty doc",
			original: "foo:\n  - null",
			explicit: "",
			implicit: "foo:\n  - null",
			both:     "",
		},
		{
			name: "null nodes",
			original: `
a:
b:
  a:
  x: 1
  b:
    c: 5
  b: null
foo:
  a:
bar:
  a: null
  b:
    c:
g: null
`,
			explicit: `
a:
b:
  a:
  x: 1
  b:
    c: 5
foo:
  a:
bar:
  b:
    c:
`,
			implicit: `
b:
  x: 1
  b:
    c: 5
  b: null
bar:
  a: null
g: null
`,
			both: `
b:
  x: 1
  b:
    c: 5
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var t1, t2, t3, ex, im, both yaml.Node

			NoError(t, yaml.Unmarshal([]byte(tt.original), &t1))
			NoError(t, yaml.Unmarshal([]byte(tt.original), &t2))
			NoError(t, yaml.Unmarshal([]byte(tt.original), &t3))

			NoError(t, yaml.Unmarshal([]byte(tt.explicit), &ex))
			NoError(t, yaml.Unmarshal([]byte(tt.implicit), &im))
			NoError(t, yaml.Unmarshal([]byte(tt.both), &both))

			// fmt.Println(">>", (*DebugYamlNode)(&src).String())

			FilterYAMLNullNodes(&t1, true, false)
			FilterYAMLNullNodes(&t2, false, true)
			FilterYAMLNullNodes(&t3, true, true)

			// fmt.Println(">>", (*DebugYamlNode)(&src).String())

			for _, p := range [][]*yaml.Node{
				{&t1, &ex},
				{&t2, &im},
				{&t3, &both},
			} {
				srcBytes, err := yaml.Marshal(p[0])
				NoError(t, err, (*DebugYamlNode)(p[0]).String())

				expectedBytes, err := yaml.Marshal(p[1])
				NoError(t, err)

				if string(expectedBytes) != string(srcBytes) {
					Fail(t, errors.New("unexpected result"), "%s\n!=\n\n%s", string(expectedBytes), string(srcBytes))
				}
			}
		})
	}
}

func NoError(t *testing.T, err error, msgAndArgs ...any) {
	if err != nil {
		Fail(t, err, msgAndArgs...)
	}
}

func ErrorContains(t *testing.T, theError error, contains string, msgAndArgs ...any) {
	if !strings.Contains(fmt.Sprintf("%v", theError), contains) {
		Fail(t, theError, msgAndArgs...)
	}
}

func Fail(t *testing.T, err error, msgAndArgs ...any) {
	t.Errorf("Received unexpected error:\n%+v", err)
	if len(msgAndArgs) > 0 {
		t.Errorf(msgAndArgs[0].(string), msgAndArgs[1:]...)
	}
	t.FailNow()
}
