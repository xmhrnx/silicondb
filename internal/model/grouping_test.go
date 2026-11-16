package model

import "testing"

func TestGenerations(t *testing.T) {
	chips := []*Chip{
		{ID: "a", Released: "2023-10", Generation: "Core 14th Gen", Specs: map[string]any{"market": "Desktop"}},
		{ID: "b", Released: "2022-10", Generation: "Core 13th Gen", Specs: map[string]any{"market": "Desktop"}},
		{ID: "c", Released: "2024-01", Generation: "Core 14th Gen", Specs: map[string]any{"market": "Laptop"}},
	}
	gens := Generations(chips)
	if len(gens) != 2 || gens[0].Name != "Core 14th Gen" || gens[0].Chips[0].ID != "c" || gens[0].First != 2023 || gens[0].Last != 2024 {
		t.Fatalf("unexpected grouping: %+v", gens[0])
	}
	if got := Segments(chips); len(got) != 2 || got[0] != "Desktop" || got[1] != "Laptop" {
		t.Fatalf("Segments = %v", got)
	}
}
