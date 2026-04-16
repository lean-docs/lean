package numbering

import (
	"reflect"
	"testing"
)

func mustCalc(t *testing.T, m *Manager, numID string, level, pos int) int {
	t.Helper()
	v, err := m.Calculate(numID, level, pos)
	if err != nil {
		t.Fatalf("Calculate(%s,%d,%d): %v", numID, level, pos, err)
	}
	return v
}

func TestFlatListIncrements(t *testing.T) {
	m := NewManager()
	for i := 0; i < 5; i++ {
		v := mustCalc(t, m, "1", 0, i)
		if v != i+1 {
			t.Errorf("pos %d: got %d, want %d", i, v, i+1)
		}
		m.SetCounter("1", 0, i, v, "")
	}
}

func TestCustomStart(t *testing.T) {
	m := NewManager()
	m.SetStart("1", 0, 5, nil, false)
	if got := mustCalc(t, m, "1", 0, 0); got != 5 {
		t.Errorf("first with start=5: got %d", got)
	}
	m.SetCounter("1", 0, 0, 5, "")
	if got := mustCalc(t, m, "1", 0, 1); got != 6 {
		t.Errorf("second: got %d", got)
	}
}

func TestNestedListRestartsChild(t *testing.T) {
	// default restart semantics: child restarts when parent fires.
	m := NewManager()
	restart := 0
	m.SetStart("1", 1, 1, &restart, false) // level-1 restarts at level ≤ 0
	// pos 0: parent a
	m.SetCounter("1", 0, 0, mustCalc(t, m, "1", 0, 0), "")
	// pos 1, 2: two children
	m.SetCounter("1", 1, 1, mustCalc(t, m, "1", 1, 1), "")
	m.SetCounter("1", 1, 2, mustCalc(t, m, "1", 1, 2), "")
	if v, _ := m.Counter("1", 1, 2); v != 2 {
		t.Errorf("second child should be 2, got %d", v)
	}
	// pos 3: next parent
	m.SetCounter("1", 0, 3, mustCalc(t, m, "1", 0, 3), "")
	// pos 4: child should restart if restart semantics say so.
	// With restart=0 ("never restart"), it increments.
	c, _ := m.Calculate("1", 1, 4)
	if c != 3 {
		t.Errorf("restart=0 means continue; want 3, got %d", c)
	}
}

func TestRestartNilRestartsOnAnyLowerLevel(t *testing.T) {
	m := NewManager()
	// No restart configured → restart when any lower level fires.
	m.SetCounter("1", 0, 0, mustCalc(t, m, "1", 0, 0), "")
	m.SetCounter("1", 1, 1, mustCalc(t, m, "1", 1, 1), "")
	m.SetCounter("1", 1, 2, mustCalc(t, m, "1", 1, 2), "")
	m.SetCounter("1", 0, 3, mustCalc(t, m, "1", 0, 3), "")
	// next child at pos 4
	v, _ := m.Calculate("1", 1, 4)
	if v != 1 {
		t.Errorf("child should restart to 1 after parent, got %d", v)
	}
}

func TestAncestorPath(t *testing.T) {
	m := NewManager()
	m.SetCounter("1", 0, 0, mustCalc(t, m, "1", 0, 0), "") // 1
	m.SetCounter("1", 1, 1, mustCalc(t, m, "1", 1, 1), "") // 1.1
	m.SetCounter("1", 1, 2, mustCalc(t, m, "1", 1, 2), "") // 1.2
	m.SetCounter("1", 2, 3, mustCalc(t, m, "1", 2, 3), "") // 1.2.1
	got := m.FullPath("1", 2, 3)
	want := []int{1, 2, 1}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("FullPath = %v, want %v", got, want)
	}
}

func TestCacheReturnsCopy(t *testing.T) {
	m := NewManager()
	m.EnableCache()
	m.SetCounter("1", 0, 0, 1, "")
	m.SetCounter("1", 1, 1, 1, "")
	first := m.AncestorPath("1", 1, 1)
	first[0] = 999
	second := m.AncestorPath("1", 1, 1)
	if second[0] == 999 {
		t.Error("cache returned a shared slice; callers could mutate internal state")
	}
}

func TestInvalidInputsPanic(t *testing.T) {
	m := NewManager()
	for _, tc := range []struct {
		name string
		fn   func()
	}{
		{"empty id", func() { m.Calculate("", 0, 0) }},
		{"negative level", func() { m.Calculate("1", -1, 0) }},
		{"negative pos", func() { m.Calculate("1", 0, -1) }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Error("expected panic")
				}
			}()
			tc.fn()
		})
	}
}
