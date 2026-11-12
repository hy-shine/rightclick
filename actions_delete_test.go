package main

import (
	"fmt"
	"os"
	"testing"
)

// TestDeletePermanentlyRefusesProtectedPaths checks that the root and the home
// folder are refused before the confirmation dialog, so no dialog opens for them.
func TestDeletePermanentlyRefusesProtectedPaths(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range []string{"/", home} {
		if err := deletePermanently([]string{p}); err == nil {
			t.Errorf("deletePermanently(%q) = nil, want a refusal", p)
		}
	}
}

// TestDeleteDetailListsNamesAndCountsTheRest checks the text under the delete
// question: the first names, then how many more items there are.
func TestDeleteDetailListsNamesAndCountsTheRest(t *testing.T) {
	useLocale(t, "en")
	var paths []string
	for i := range 7 {
		paths = append(paths, fmt.Sprintf("/tmp/item%d", i))
	}
	want := "item0\nitem1\nitem2\nitem3\nitem4\n…and 2 more items\n\nThis cannot be undone."
	if got := deleteDetail(paths); got != want {
		t.Errorf("deleteDetail() =\n%q\nwant\n%q", got, want)
	}
}
