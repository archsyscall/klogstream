package matcher

import (
	"testing"
)

func TestJSONMatcher_ShouldMerge(t *testing.T) {
	matcher := NewJSONMatcher()

	tests := []struct {
		name     string
		lines    []string
		expected []bool
	}{
		{
			name: "simple json object",
			lines: []string{
				"{",
				"  \"key\": \"value\"",
				"}",
				"Next log",
			},
			expected: []bool{
				true,  // After '{', still in object
				true,  // After '  "key": "value"', still in object
				false, // After '}', no longer in object
			},
		},
		{
			name: "nested json object",
			lines: []string{
				"{",
				"  \"outer\": {",
				"    \"inner\": \"value\"",
				"  }",
				"}",
				"Next log",
			},
			expected: []bool{
				true,  // After '{', still in object
				true,  // After '  "outer": {', deeper in object
				true,  // After '    "inner": "value"', still in object
				true,  // After '  }', still in outer object
				false, // After '}', no longer in object
			},
		},
		{
			name: "continuation with backslash",
			lines: []string{
				"This is a line \\",
				"that continues here",
			},
			expected: []bool{
				true, // After backslash, should continue
			},
		},
		{
			name: "single line json",
			lines: []string{
				"{ \"key\": \"value\" }",
				"Next log",
			},
			expected: []bool{
				false, // Object opens and closes on same line
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset matcher for each test
			matcher = NewJSONMatcher()

			for i := 0; i < len(tt.lines)-1; i++ {
				previous := tt.lines[i]
				next := tt.lines[i+1]

				got := matcher.ShouldMerge(previous, next)
				if got != tt.expected[i] {
					t.Errorf("Line %d: JSONMatcher.ShouldMerge(%q, %q) = %v, want %v",
						i, previous, next, got, tt.expected[i])
				}
			}
		})
	}
}
