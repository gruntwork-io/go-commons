package collections

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMakeCopyOfList(t *testing.T) {
	originalStr := []string{"foo", "bar", "baz"}
	copyOfListStr := MakeCopyOfList(originalStr)
	assert.Equal(t, originalStr, copyOfListStr)

	originalInt := []int{1, 2, 3}
	copyOfListInt := MakeCopyOfList(originalInt)
	assert.Equal(t, originalInt, copyOfListInt)
}

func TestListContainsElement(t *testing.T) {
	t.Parallel()

	testCasesStr := []struct {
		list     []string
		element  string
		expected bool
	}{
		{[]string{}, "", false},
		{[]string{}, "foo", false},
		{[]string{"foo"}, "foo", true},
		{[]string{"bar", "foo", "baz"}, "foo", true},
		{[]string{"bar", "foo", "baz"}, "nope", false},
		{[]string{"bar", "foo", "baz"}, "", false},
	}

	for _, testCase := range testCasesStr {
		actual := ListContainsElement(testCase.list, testCase.element)
		assert.Equal(t, testCase.expected, actual, "For list %v and element %s", testCase.list, testCase.element)
	}

	testCasesInt := []struct {
		list     []int
		element  int
		expected bool
	}{
		{[]int{}, 0, false},
		{[]int{}, 1, false},
		{[]int{1}, 1, true},
		{[]int{1, 2, 3}, 1, true},
		{[]int{1, 2, 3}, 4, false},
		{[]int{1, 2, 3}, 0, false},
	}

	for _, testCase := range testCasesInt {
		actual := ListContainsElement(testCase.list, testCase.element)
		assert.Equal(t, testCase.expected, actual, "For list %v and element %s", testCase.list, testCase.element)
	}
}

func TestRemoveElementFromList(t *testing.T) {
	t.Parallel()

	testCasesStr := []struct {
		list     []string
		element  string
		expected []string
	}{
		{[]string{}, "", []string{}},
		{[]string{}, "foo", []string{}},
		{[]string{"foo"}, "foo", []string{}},
		{[]string{"bar"}, "foo", []string{"bar"}},
		{[]string{"bar", "foo", "baz"}, "foo", []string{"bar", "baz"}},
		{[]string{"bar", "foo", "baz"}, "nope", []string{"bar", "foo", "baz"}},
		{[]string{"bar", "foo", "baz"}, "", []string{"bar", "foo", "baz"}},
	}

	for _, testCase := range testCasesStr {
		actual := RemoveElementFromList(testCase.list, testCase.element)
		assert.Equal(t, testCase.expected, actual, "For list %v and element %s", testCase.list, testCase.element)
	}

	testCasesInt := []struct {
		list     []int
		element  int
		expected []int
	}{
		{[]int{}, 0, []int{}},
		{[]int{}, 1, []int{}},
		{[]int{1}, 1, []int{}},
		{[]int{1}, 2, []int{1}},
		{[]int{1, 2, 3}, 1, []int{2, 3}},
		{[]int{1, 2, 3}, 4, []int{1, 2, 3}},
		{[]int{1, 2, 3}, 0, []int{1, 2, 3}},
	}

	for _, testCase := range testCasesInt {
		actual := RemoveElementFromList(testCase.list, testCase.element)
		assert.Equal(t, testCase.expected, actual, "For list %v and element %s", testCase.list, testCase.element)
	}

}

func TestBatchListIntoGroupsOf(t *testing.T) {
	t.Parallel()

	testCasesStr := []struct {
		stringList []string
		n          int
		result     [][]string
	}{
		{
			[]string{"macaroni", "gentoo", "magellanic", "adelie", "little", "king", "emperor"},
			2,
			[][]string{
				[]string{"macaroni", "gentoo"},
				[]string{"magellanic", "adelie"},
				[]string{"little", "king"},
				[]string{"emperor"},
			},
		},
		{
			[]string{"macaroni", "gentoo", "magellanic", "adelie", "king", "emperor"},
			2,
			[][]string{
				[]string{"macaroni", "gentoo"},
				[]string{"magellanic", "adelie"},
				[]string{"king", "emperor"},
			},
		},
		{
			[]string{"macaroni", "gentoo", "magellanic"},
			5,
			[][]string{
				[]string{"macaroni", "gentoo", "magellanic"},
			},
		},
		{
			[]string{"macaroni", "gentoo", "magellanic"},
			-1,
			nil,
		},
		{
			[]string{"macaroni", "gentoo", "magellanic"},
			0,
			nil,
		},
		{
			[]string{},
			7,
			[][]string{},
		},
	}

	for idx, testCase := range testCasesStr {
		t.Run(fmt.Sprintf("%s_%d", t.Name(), idx), func(t *testing.T) {
			t.Parallel()
			original := MakeCopyOfList(testCase.stringList)
			assert.Equal(t, BatchListIntoGroupsOf(testCase.stringList, testCase.n), testCase.result)
			// Make sure the function doesn't modify the original list
			assert.Equal(t, testCase.stringList, original)
		})
	}

	testCasesInt := []struct {
		intList []int
		n       int
		result  [][]int
	}{
		{
			[]int{1, 2, 3, 4, 5, 6, 7},
			2,
			[][]int{
				[]int{1, 2},
				[]int{3, 4},
				[]int{5, 6},
				[]int{7},
			},
		},
		{
			[]int{1, 2, 3, 4, 5, 6},
			2,
			[][]int{
				[]int{1, 2},
				[]int{3, 4},
				[]int{5, 6},
			},
		},
		{
			[]int{1, 2, 3},
			5,
			[][]int{
				[]int{1, 2, 3},
			},
		},
		{
			[]int{1, 2, 3},
			-1,
			nil,
		},
		{
			[]int{1, 2, 3},
			0,
			nil,
		},
		{
			[]int{},
			7,
			[][]int{},
		},
	}

	for idx, testCase := range testCasesInt {
		t.Run(fmt.Sprintf("%s_%d", t.Name(), idx), func(t *testing.T) {
			t.Parallel()
			original := MakeCopyOfList(testCase.intList)
			assert.Equal(t, BatchListIntoGroupsOf(testCase.intList, testCase.n), testCase.result)
			// Make sure the function doesn't modify the original list
			assert.Equal(t, testCase.intList, original)
		})
	}
}
