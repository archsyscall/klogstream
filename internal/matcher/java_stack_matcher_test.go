package matcher

import (
	"testing"
)

func TestJavaStackMatcher_ShouldMerge(t *testing.T) {
	matcher := NewJavaStackMatcher()

	tests := []struct {
		name     string
		previous string
		next     string
		want     bool
	}{
		{
			name:     "tab indentation",
			previous: "Exception in thread \"main\" java.lang.NullPointerException",
			next:     "\tat com.example.Main.method(Main.java:10)",
			want:     true,
		},
		{
			name:     "at line",
			previous: "Exception in thread \"main\" java.lang.NullPointerException",
			next:     "    at com.example.Main.method(Main.java:10)",
			want:     true,
		},
		{
			name:     "caused by line",
			previous: "java.lang.RuntimeException: Something went wrong",
			next:     "Caused by: java.lang.NullPointerException",
			want:     true,
		},
		{
			name:     "continuation with backslash",
			previous: "This is a long line \\",
			next:     "that continues here",
			want:     true,
		},
		{
			name:     "empty line",
			previous: "Exception in thread \"main\" java.lang.NullPointerException",
			next:     "",
			want:     false,
		},
		{
			name:     "whitespace line",
			previous: "Exception in thread \"main\" java.lang.NullPointerException",
			next:     "   ",
			want:     false,
		},
		{
			name:     "new log line",
			previous: "Exception in thread \"main\" java.lang.NullPointerException",
			next:     "INFO: Starting application",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := matcher.ShouldMerge(tt.previous, tt.next)
			if got != tt.want {
				t.Errorf("JavaStackMatcher.ShouldMerge(%q, %q) = %v, want %v",
					tt.previous, tt.next, got, tt.want)
			}
		})
	}
}
