package filter

import (
	"regexp"
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/labels"
)

func TestLogFilter_IsEmpty(t *testing.T) {
	tests := []struct {
		name   string
		filter *LogFilter
		want   bool
	}{
		{
			name:   "empty filter",
			filter: NewLogFilter(),
			want:   true,
		},
		{
			name: "non-empty filter with pod regex",
			filter: &LogFilter{
				PodNameRegex:   regexp.MustCompile("test"),
				ContainerState: DefaultContainerState,
			},
			want: false,
		},
		{
			name: "non-empty filter with namespace",
			filter: &LogFilter{
				Namespaces:     []string{"default"},
				ContainerState: DefaultContainerState,
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.filter.IsEmpty(); got != tt.want {
				t.Errorf("LogFilter.IsEmpty() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLogFilter_Validate(t *testing.T) {
	future := time.Now().Add(time.Hour)
	selector := labels.SelectorFromSet(labels.Set{"app": "test"})

	tests := []struct {
		name    string
		filter  *LogFilter
		wantErr error
	}{
		{
			name:    "empty filter",
			filter:  NewLogFilter(),
			wantErr: ErrEmptyFilter,
		},
		{
			name: "no namespace",
			filter: &LogFilter{
				PodNameRegex: regexp.MustCompile("test"),
			},
			wantErr: ErrNoNamespaceSpecified,
		},
		{
			name: "invalid container state",
			filter: &LogFilter{
				PodNameRegex:   regexp.MustCompile("test"),
				Namespaces:     []string{"default"},
				ContainerState: "invalid",
			},
			wantErr: ErrInvalidContainerState,
		},
		{
			name: "future since time",
			filter: &LogFilter{
				PodNameRegex:   regexp.MustCompile("test"),
				Namespaces:     []string{"default"},
				ContainerState: "all",
				Since:          &future,
			},
			wantErr: ErrInvalidSinceTime,
		},
		{
			name: "valid filter",
			filter: &LogFilter{
				PodNameRegex:   regexp.MustCompile("test"),
				ContainerRegex: regexp.MustCompile("web"),
				LabelSelector:  selector,
				IncludeRegex:   regexp.MustCompile("ERROR"),
				Namespaces:     []string{"default"},
				ContainerState: "running",
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.filter.Validate()
			if err != tt.wantErr {
				t.Errorf("LogFilter.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
