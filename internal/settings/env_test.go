package settings

import "testing"

// TestPluginkitListingMeansEnabled reads pluginkit's listing: a "+" line means the
// Finder extension is on, and anything else means it is off or not found.
func TestPluginkitListingMeansEnabled(t *testing.T) {
	cases := []struct {
		listing string
		want    bool
	}{
		{"+    com.example.rightclick.FinderSync(0.1.0)\n", true},
		{"  +    com.example.rightclick.FinderSync(0.1.0)\n", true},
		{"-    com.example.rightclick.FinderSync(0.1.0)\n", false},
		{"", false},
	}
	for _, tt := range cases {
		if got := pluginkitEnabled(tt.listing); got != tt.want {
			t.Errorf("pluginkitEnabled(%q) = %v, want %v", tt.listing, got, tt.want)
		}
	}
}
