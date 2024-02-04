package main

import (
	"testing"
)

func TestEntryListFromJSON(t *testing.T) {
	el := EntryList{}

	if err := el.fromJSONFile("testdata/entries.json"); err != nil {
		t.Errorf("Expected no error, got %s", err)
	}

	if len(el.Entries) != 4 {
		t.Errorf("Expected 4 entries, got %d", len(el.Entries))
	}

	//for _, entry := range el.Entries {
	//	if entry.ID == 0 {
	//		t.Errorf("Expected non-zero ID, got %d", entry.ID)
	//	}
	//}
}
