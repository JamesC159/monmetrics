package features

import "testing"

func TestEnabledFallsBackToDeclaredDefault(t *testing.T) {
	r := NewRegistry(nil)
	for _, d := range Definitions {
		if got := r.Enabled(d.Key); got != d.Default {
			t.Fatalf("%s = %v, want default %v", d.Key, got, d.Default)
		}
	}
}

// Every gated surface ships off: a flag exists because its data source does not.
func TestAllFlagsDefaultOff(t *testing.T) {
	for _, d := range Definitions {
		if d.Default {
			t.Fatalf("%s defaults on; gated surfaces must default off", d.Key)
		}
		if d.Blocker == "" {
			t.Fatalf("%s has no blocker explaining what it waits on", d.Key)
		}
	}
}

func TestUnknownKeysAreRejected(t *testing.T) {
	if IsKnown("nope.not.a.flag") {
		t.Fatal("unknown key should be rejected so admins cannot write arbitrary docs")
	}
	if !IsKnown(KeyEbaySoldListings) {
		t.Fatal("declared key should be known")
	}
}

// A nil registry must not panic, so handlers constructed in tests still work.
func TestNilRegistryUsesDefaults(t *testing.T) {
	var r *Registry
	if r.Enabled(KeyEbayGrades) {
		t.Fatal("nil registry should report the default, not true")
	}
}

func TestStatesCoverEveryDefinition(t *testing.T) {
	states := NewRegistry(nil).States()
	if len(states) != len(Definitions) {
		t.Fatalf("States returned %d entries, want %d", len(states), len(Definitions))
	}
	for i, s := range states {
		if s.Key != Definitions[i].Key {
			t.Fatalf("States[%d] = %s, want declaration order %s", i, s.Key, Definitions[i].Key)
		}
	}
}
