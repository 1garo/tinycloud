package tui

import "testing"

func TestSlug(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  string
	}{
		{name: "simple name", value: "Expense Tracker", want: "expense-tracker"},
		{name: "punctuation", value: " team/dashboard! ", want: "team-dashboard"},
		{name: "empty", value: "---", want: ""},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := slug(test.value); got != test.want {
				t.Fatalf("slug(%q) = %q, want %q", test.value, got, test.want)
			}
		})
	}
}
