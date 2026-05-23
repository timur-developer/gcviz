package runner

import (
	"reflect"
	"strings"
	"testing"
)

func TestNormalizeGODEBUG(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  []string
	}{
		{
			name:  "empty",
			input: "",
			want:  []string{"gctrace=1", "gcpacertrace=1"},
		},
		{
			name:  "already set",
			input: "gctrace=1,gcpacertrace=1",
			want:  []string{"gctrace=1", "gcpacertrace=1"},
		},
		{
			name:  "override values",
			input: "gctrace=0,gcpacertrace=0",
			want:  []string{"gctrace=1", "gcpacertrace=1"},
		},
		{
			name:  "keeps other flags",
			input: "panic=1,gctrace=0",
			want:  []string{"panic=1", "gctrace=1", "gcpacertrace=1"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := normalizeGODEBUG(tc.input)
			gotParts := strings.Split(got, ",")
			if !containsAll(gotParts, tc.want) {
				t.Fatalf("expected %v, got %v", tc.want, gotParts)
			}
		})
	}
}

func TestMergeEnv(t *testing.T) {
	base := []string{
		"PATH=/bin",
		"GODEBUG=panic=1,gctrace=0",
	}

	extra := map[string]string{
		"FOO": "bar",
	}

	merged := mergeEnv(base, extra)
	got := envSliceToMap(merged)

	want := map[string]string{
		"PATH":    "/bin",
		"FOO":     "bar",
		"GODEBUG": "panic=1,gctrace=1,gcpacertrace=1",
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("expected %v, got %v", want, got)
	}
}

func envSliceToMap(items []string) map[string]string {
	result := make(map[string]string, len(items))
	for _, item := range items {
		parts := strings.SplitN(item, "=", 2)
		if len(parts) != 2 {
			continue
		}
		result[parts[0]] = parts[1]
	}
	return result
}

func containsAll(haystack []string, needles []string) bool {
	set := make(map[string]struct{}, len(haystack))
	for _, item := range haystack {
		set[item] = struct{}{}
	}
	for _, needle := range needles {
		if _, ok := set[needle]; !ok {
			return false
		}
	}
	return true
}

