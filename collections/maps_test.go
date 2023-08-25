package collections

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMergeMaps(t *testing.T) {
	t.Parallel()

	testCasesStr := []struct {
		name        string
		mapsToMerge []map[string]interface{}
		expected    map[string]interface{}
	}{
		{
			"No maps",
			nil,
			map[string]interface{}{},
		},
		{
			"One empty map",
			[]map[string]interface{}{{}},
			map[string]interface{}{},
		},
		{
			"Multiple empty maps",
			[]map[string]interface{}{{}, {}, {}},
			map[string]interface{}{},
		},
		{
			"One non-empty map",
			[]map[string]interface{}{{"key1": "value1"}},
			map[string]interface{}{"key1": "value1"},
		},
		{
			"Two non-empty maps",
			[]map[string]interface{}{{"key1": "value1"}, {"key2": "value2"}},
			map[string]interface{}{"key1": "value1", "key2": "value2"},
		},
		{
			"Two non-empty maps, overlapping keys",
			[]map[string]interface{}{
				{"key1": "value1", "key3": "value3"},
				{"key1": "replacement", "key2": "value2"},
			},
			map[string]interface{}{"key1": "replacement", "key2": "value2", "key3": "value3"},
		},
		{
			"Multiple non-empty maps, overlapping keys",
			[]map[string]interface{}{
				{"key1": "value1", "key3": "value3"},
				{"key1": "replacement", "key2": "value2"},
				{"key1": "replacement-two", "key3": "replacement-two", "key4": "value4"},
			},
			map[string]interface{}{"key1": "replacement-two", "key2": "value2", "key3": "replacement-two", "key4": "value4"},
		},
	}

	for _, testCase := range testCasesStr {
		// The following is necessary to make sure testCase's values don't
		// get updated due to concurrency within the scope of t.Run(..) below
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			actual := MergeMaps(testCase.mapsToMerge...)
			assert.Equal(t, testCase.expected, actual)
		})
	}

	testCasesInt := []struct {
		name        string
		mapsToMerge []map[int]int
		expected    map[int]int
	}{
		{
			"No maps",
			nil,
			map[int]int{},
		},
		{
			"One empty map",
			[]map[int]int{{}},
			map[int]int{},
		},
		{
			"Multiple empty maps",
			[]map[int]int{{}, {}, {}},
			map[int]int{},
		},
		{
			"One non-empty map",
			[]map[int]int{{1: 1}},
			map[int]int{1: 1},
		},
		{
			"Two non-empty maps",
			[]map[int]int{{1: 1}, {2: 2}},
			map[int]int{1: 1, 2: 2},
		},
		{
			"Two non-empty maps, overlapping keys",
			[]map[int]int{
				{1: 1, 3: 3},
				{1: 1000, 2: 2},
			},
			map[int]int{1: 1000, 2: 2, 3: 3},
		},
		{
			"Multiple non-empty maps, overlapping keys",
			[]map[int]int{
				{1: 1, 3: 3},
				{1: 1000, 2: 2},
				{1: 1000000, 3: 3000, 4: 4},
			},
			map[int]int{1: 1000000, 2: 2, 3: 3000, 4: 4},
		},
	}

	for _, testCase := range testCasesInt {
		// The following is necessary to make sure testCase's values don't
		// get updated due to concurrency within the scope of t.Run(..) below
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			actual := MergeMaps(testCase.mapsToMerge...)
			assert.Equal(t, testCase.expected, actual)
		})
	}
}

func TestKeys(t *testing.T) {
	t.Parallel()

	testCasesStr := []struct {
		input    map[string]string
		expected []string
	}{
		{map[string]string{}, []string{}},
		{map[string]string{"a": "foo"}, []string{"a"}},
		{map[string]string{"a": "foo", "b": "bar", "c": "baz"}, []string{"a", "b", "c"}},
	}

	for _, testCase := range testCasesStr {
		t.Run(fmt.Sprintf("%v", testCase.input), func(t *testing.T) {
			actual := Keys(testCase.input)
			assert.Equal(t, testCase.expected, actual)
		})
	}

	testCasesInt := []struct {
		input    map[int]int
		expected []int
	}{
		{map[int]int{}, []int{}},
		{map[int]int{1: 1}, []int{1}},
		{map[int]int{1: 1, 2: 2, 3: 3}, []int{1, 2, 3}},
	}

	for _, testCase := range testCasesInt {
		t.Run(fmt.Sprintf("%v", testCase.input), func(t *testing.T) {
			actual := Keys(testCase.input)
			assert.Equal(t, testCase.expected, actual)
		})
	}
}

func TestKeyValueStringSlice(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		input    map[string]string
		expected []string
	}{
		{map[string]string{"a": "foo"}, []string{"a=foo"}},
		{map[string]string{"a": "foo", "b": "bar", "c": "baz"}, []string{"a=foo", "b=bar", "c=baz"}},
	}

	for _, testCase := range testCases {
		t.Run(fmt.Sprintf("%v", testCase.input), func(t *testing.T) {
			actual := KeyValueStringSlice(testCase.input)
			assert.Equal(t, testCase.expected, actual)
		})
	}
}

func TestKeyValueStringSliceWithFormat(t *testing.T) {
	t.Parallel()

	testCasesStr := []struct {
		input    map[string]string
		format   string
		expected []string
	}{
		{map[string]string{"a": "foo"}, "%s='%s'", []string{"a='foo'"}},
		{map[string]string{"a": "foo"}, "%s%s", []string{"afoo"}},
		{map[string]string{"a": "foo", "b": "bar", "c": "baz"}, "%s=%s", []string{"a=foo", "b=bar", "c=baz"}},
	}

	for _, testCase := range testCasesStr {
		t.Run(fmt.Sprintf("%v", testCase.input), func(t *testing.T) {
			actual := KeyValueStringSliceWithFormat(testCase.input, testCase.format)
			assert.Equal(t, testCase.expected, actual)
		})
	}

	testCasesInt := []struct {
		input    map[int]int
		format   string
		expected []string
	}{
		{map[int]int{1: 11}, "%d='%d'", []string{"1='11'"}},
		{map[int]int{1: 11}, "%d%d", []string{"111"}},
		{map[int]int{1: 11, 2: 22, 3: 33}, "%d=%d", []string{"1=11", "2=22", "3=33"}},
	}

	for _, testCase := range testCasesInt {
		t.Run(fmt.Sprintf("%v", testCase.input), func(t *testing.T) {
			actual := KeyValueStringSliceWithFormat(testCase.input, testCase.format)
			assert.Equal(t, testCase.expected, actual)
		})
	}
}

func TestKeyValueStringSliceAsMap(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    []string
		expected map[string][]string
	}{
		{
			"BaseCase",
			[]string{"key=value"},
			map[string][]string{"key": []string{"value"}},
		},
		{"EmptyCase", []string{}, map[string][]string{}},
		{
			"RepeatedKey",
			[]string{"key=valueA", "foo=bar", "key=valueB"},
			map[string][]string{
				"key": []string{"valueA", "valueB"},
				"foo": []string{"bar"},
			},
		},
		{
			"EmptyValue",
			[]string{"key", "foo=", "foo=baz"},
			map[string][]string{
				"key": []string{""},
				"foo": []string{"", "baz"},
			},
		},
		{
			"EqualInValue",
			[]string{"key=foo=bar"},
			map[string][]string{
				"key": []string{"foo=bar"},
			},
		},
		{
			"EmptyString",
			[]string{""},
			map[string][]string{
				"": []string{""},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			actual := KeyValueStringSliceAsMap(testCase.input)
			assert.Equal(t, testCase.expected, actual)
		})
	}
}

func TestMapJoin(t *testing.T) {
	t.Parallel()

	var testCases = []struct {
		vals             any
		sliceSep, mapSep string
		expected         string
	}{
		{map[string]string{"color": "white", "number": "two"}, ",", "=", "color=white,number=two"},
		{map[int]int{10: 100, 20: 200}, " ", ":", "10:100 20:200"},
	}

	for i, testCase := range testCases {
		// to make sure testCase's values don't get updated due to concurrency within the scope of t.Run(..) below
		testCase := testCase

		t.Run(fmt.Sprintf("test-%d-vals-%v-expected-%s", i, testCase.vals, testCase.expected), func(t *testing.T) {
			t.Parallel()

			var actual string

			switch vals := testCase.vals.(type) {
			case map[string]string:
				actual = MapJoin(vals, testCase.sliceSep, testCase.mapSep)
			case map[int]int:
				actual = MapJoin(vals, testCase.sliceSep, testCase.mapSep)
			}
			assert.Equal(t, testCase.expected, actual)
		})
	}
}

func TestMapToSlice(t *testing.T) {
	t.Parallel()

	var testCases = []struct {
		vals     any
		sep      string
		expected []string
	}{
		{map[string]string{"color": "white", "number": "two"}, "=", []string{"color=white", "number=two"}},
		{map[int]int{10: 100, 20: 200}, ":", []string{"10:100", "20:200"}},
	}

	for i, testCase := range testCases {
		// to make sure testCase's values don't get updated due to concurrency within the scope of t.Run(..) below
		testCase := testCase

		t.Run(fmt.Sprintf("test-%d-vals-%v-expected-%s", i, testCase.vals, testCase.expected), func(t *testing.T) {
			t.Parallel()

			var actual []string

			switch vals := testCase.vals.(type) {
			case map[string]string:
				actual = MapToSlice(vals, testCase.sep)
			case map[int]int:
				actual = MapToSlice(vals, testCase.sep)
			}

			assert.Subset(t, testCase.expected, actual)
		})
	}
}
