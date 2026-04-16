package numbering

import (
	"errors"
	"math"
	"sort"
)

// ErrCounterOverflow is returned when incrementing a counter would exceed
// the safe integer range.
var ErrCounterOverflow = errors.New("numbering: counter overflow")

// A Manager tracks the state of one or more OOXML numbering definitions as
// paragraphs are encountered in document order. Positions are monotonic
// indices supplied by the caller (typically paragraph index) — they are not
// interpreted, only compared.
//
// The Manager is not safe for concurrent use.
type Manager struct {
	counters map[string]map[int]map[int]int    // numId → level → pos → value
	starts   map[string]map[int]startSettings  // numId → level → settings
	abstract map[string]string                 // numId → abstractId
	cache    map[string]map[int]map[int][]int  // numId → level → pos → path
	cacheOn  bool
}

type startSettings struct {
	Start           int
	Restart         *int
	StartOverridden bool
}

// NewManager returns a fresh Manager with no state.
func NewManager() *Manager {
	return &Manager{
		counters: map[string]map[int]map[int]int{},
		starts:   map[string]map[int]startSettings{},
		abstract: map[string]string{},
		cache:    map[string]map[int]map[int][]int{},
	}
}

// SetStart configures the starting counter value for (numId, level).
// Restart may be nil, or point to a level at/below which a restart should
// occur when that level is used.
func (m *Manager) SetStart(numID string, level, start int, restart *int, overridden bool) {
	mustValidID(numID)
	mustValidLevel(level)
	if m.starts[numID] == nil {
		m.starts[numID] = map[int]startSettings{}
	}
	m.starts[numID][level] = startSettings{Start: start, Restart: restart, StartOverridden: overridden}
}

// SetCounter explicitly records a counter value for (numId, level, pos).
// Typically called after Calculate so that later positions can see history.
func (m *Manager) SetCounter(numID string, level, pos, value int, abstractID string) {
	mustValidID(numID)
	mustValidLevel(level)
	mustValidPos(pos)
	m.abstract[numID] = abstractID
	if m.counters[numID] == nil {
		m.counters[numID] = map[int]map[int]int{}
	}
	if m.counters[numID][level] == nil {
		m.counters[numID][level] = map[int]int{}
	}
	m.counters[numID][level][pos] = value
	delete(m.cache, numID)
}

// Counter returns the counter recorded at (numId, level, pos), or 0,false
// if nothing has been set there.
func (m *Manager) Counter(numID string, level, pos int) (int, bool) {
	v, ok := m.counters[numID][level][pos]
	return v, ok
}

// Calculate returns the counter value for a paragraph at (numId, level, pos)
// based on prior state. It does NOT record the result — call SetCounter for
// that. Returns ErrCounterOverflow if the next value would not fit in int.
func (m *Manager) Calculate(numID string, level, pos int) (int, error) {
	mustValidID(numID)
	mustValidLevel(level)
	mustValidPos(pos)

	start := m.startOf(numID, level)
	restart := m.restartOf(numID, level)
	prevPos, prevCount, hasPrev := previousCounter(m.counters[numID][level], pos)
	if !hasPrev {
		prevCount = start - 1
	}

	// Restart == 0 means "never restart" in OOXML: keep incrementing.
	if restart != nil && *restart == 0 {
		return safeInc(prevCount)
	}
	if !hasPrev {
		return start, nil
	}

	// Did any lower level fire between prevPos and pos?
	used := usedLowerLevels(m.counters[numID], level, prevPos, pos)
	if len(used) == 0 {
		return safeInc(prevCount)
	}
	if restart == nil {
		return start, nil
	}
	for _, lvl := range used {
		if lvl <= *restart {
			return start, nil
		}
	}
	return safeInc(prevCount)
}

// AncestorPath returns the counter values at every level above `level`
// that apply at position `pos` — suitable for rendering "1.2.3" style
// markers. Levels with no prior counter use their configured start.
func (m *Manager) AncestorPath(numID string, level, pos int) []int {
	mustValidID(numID)
	mustValidLevel(level)
	mustValidPos(pos)

	if m.cacheOn {
		if cached := m.cache[numID][level][pos]; cached != nil {
			out := make([]int, len(cached))
			copy(out, cached)
			return out
		}
	}

	path := make([]int, 0, level)
	for lvl := 0; lvl < level; lvl++ {
		startCount := m.startOf(numID, lvl)
		levelData := m.counters[numID][lvl]
		if len(levelData) == 0 {
			path = append(path, startCount)
			continue
		}
		prev := lastPosBefore(levelData, pos)
		if prev < 0 {
			path = append(path, startCount)
			continue
		}
		path = append(path, levelData[prev])
	}

	if m.cacheOn {
		if m.cache[numID] == nil {
			m.cache[numID] = map[int]map[int][]int{}
		}
		if m.cache[numID][level] == nil {
			m.cache[numID][level] = map[int][]int{}
		}
		stored := make([]int, len(path))
		copy(stored, path)
		m.cache[numID][level][pos] = stored
	}
	return path
}

// FullPath appends the counter at (numId, level, pos) to the ancestor path,
// when one has been recorded.
func (m *Manager) FullPath(numID string, level, pos int) []int {
	path := m.AncestorPath(numID, level, pos)
	if v, ok := m.Counter(numID, level, pos); ok {
		path = append(path, v)
	}
	return path
}

// EnableCache turns on memoization for AncestorPath. Clears runtime state.
func (m *Manager) EnableCache() {
	m.cacheOn = true
	m.clearRuntime()
}

// DisableCache turns off memoization and clears runtime state.
func (m *Manager) DisableCache() {
	m.cacheOn = false
	m.clearRuntime()
}

// Reset clears all state including start settings.
func (m *Manager) Reset() {
	m.starts = map[string]map[int]startSettings{}
	m.clearRuntime()
}

func (m *Manager) clearRuntime() {
	m.counters = map[string]map[int]map[int]int{}
	m.cache = map[string]map[int]map[int][]int{}
	m.abstract = map[string]string{}
}

func (m *Manager) startOf(numID string, level int) int {
	if s, ok := m.starts[numID][level]; ok {
		return s.Start
	}
	return 1
}
func (m *Manager) restartOf(numID string, level int) *int {
	if s, ok := m.starts[numID][level]; ok {
		return s.Restart
	}
	return nil
}

func mustValidID(id string) {
	if id == "" {
		panic("numbering: empty numId")
	}
}
func mustValidLevel(l int) {
	if l < 0 {
		panic("numbering: negative level")
	}
}
func mustValidPos(p int) {
	if p < 0 {
		panic("numbering: negative position")
	}
}

func safeInc(v int) (int, error) {
	if v == math.MaxInt {
		return 0, ErrCounterOverflow
	}
	return v + 1, nil
}

func previousCounter(levelData map[int]int, pos int) (prevPos, prevCount int, ok bool) {
	prevPos = -1
	for p := range levelData {
		if p >= pos {
			continue
		}
		if p > prevPos {
			prevPos = p
		}
	}
	if prevPos < 0 {
		return 0, 0, false
	}
	return prevPos, levelData[prevPos], true
}

func lastPosBefore(levelData map[int]int, pos int) int {
	best := -1
	for p := range levelData {
		if p < pos && p > best {
			best = p
		}
	}
	return best
}

// usedLowerLevels returns the lower levels that fired strictly between
// (prevPos, pos), sorted ascending.
func usedLowerLevels(levels map[int]map[int]int, level, prevPos, pos int) []int {
	var used []int
	for lvl := 0; lvl < level; lvl++ {
		data := levels[lvl]
		for p := range data {
			if p > prevPos && p < pos {
				used = append(used, lvl)
				break
			}
		}
	}
	sort.Ints(used)
	return used
}
