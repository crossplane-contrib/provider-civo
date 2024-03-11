package civokubernetes

import "testing"

func TestStringSlicesNeedUpdate_False(t *testing.T) {

	type testCaseConfig struct {
		name   string
		a, b   []string
		expect bool
	}

	testCases := []testCaseConfig{
		{
			name:   "same-length",
			a:      []string{"apple", "banana", "pear"},
			b:      []string{"apple", "banana", "pear"},
			expect: false,
		},
		{
			name:   "zero-length",
			a:      []string{},
			b:      []string{},
			expect: false,
		},
		{
			name:   "add-element",
			a:      []string{"apple", "banana", "pear"},
			b:      []string{"apple", "banana", "pear", "strawberry"},
			expect: true,
		},
		{
			name:   "remove-element",
			a:      []string{"apple", "banana", "pear"},
			b:      []string{"apple", "banana"},
			expect: true,
		},
		{
			name:   "replace-element",
			a:      []string{"apple", "banana", "pear"},
			b:      []string{"apple", "banana", "strawberry"},
			expect: true,
		},
		{
			name:   "out-of-order-same-elements",
			a:      []string{"pear", "banana", "apple"},
			b:      []string{"apple", "pear", "banana"},
			expect: false,
		},
		{
			name:   "out-of-order-different-elements",
			a:      []string{"pear", "banana", "apple"},
			b:      []string{"apple", "strawberry", "banana"},
			expect: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			result := stringSlicesNeedUpdate(testCase.a, testCase.b)
			if result != testCase.expect {
				t.Errorf("expected %v, actual %v", testCase.expect, result)
			}
		})
	}
}
