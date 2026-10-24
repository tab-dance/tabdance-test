package terminal

import "testing"

func TestCreateGeneratesUniqueIDs(t *testing.T) {
	m := NewTerminalManager()

	ids := map[string]bool{}
	for i := 0; i < 50; i++ {
		mt, err := m.Create(CreateConfig{Cols: 20, Rows: 5})
		if err != nil {
			t.Fatalf("create terminal %d: %v", i, err)
		}
		defer mt.Close()

		if ids[mt.ID] {
			t.Fatalf("duplicate terminal ID %q after %d creates", mt.ID, i)
		}
		ids[mt.ID] = true
	}
}

func TestGetUnknownTerminal(t *testing.T) {
	m := NewTerminalManager()

	if _, err := m.Get("does_not_exist"); err == nil {
		t.Fatal("expected error for unknown terminal ID")
	}
}

func TestDeleteRemovesTerminal(t *testing.T) {
	m := NewTerminalManager()

	mt, err := m.Create(CreateConfig{Cols: 20, Rows: 5})
	if err != nil {
		t.Fatalf("create terminal: %v", err)
	}

	if err := m.Delete(mt.ID); err != nil {
		t.Fatalf("delete terminal: %v", err)
	}

	if _, err := m.Get(mt.ID); err == nil {
		t.Fatal("expected deleted terminal to be gone")
	}

	if err := m.Delete(mt.ID); err == nil {
		t.Fatal("expected error deleting already-removed terminal")
	}
}

func TestGenerateRandomNameListsUnique(t *testing.T) {
	seen := map[string]bool{}
	for _, n := range nouns {
		if seen[n] {
			t.Fatalf("duplicate noun %q in name lists", n)
		}
		seen[n] = true
	}

	seenAdj := map[string]bool{}
	for _, a := range adjectives {
		if seenAdj[a] {
			t.Fatalf("duplicate adjective %q in name lists", a)
		}
		seenAdj[a] = true
	}
}
