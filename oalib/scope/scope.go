package scope

import (
	"slices"
	"strings"

	mapset "github.com/deckarep/golang-set/v2"
)

type Scope struct {
	set       mapset.Set[string]
	setstr    string
	setlength int

	slice       []string
	slicestr    string
	slicelength int
}

func New(raw string) *Scope {
	slice := make([]string, 0)
	for _, scope := range strings.Split(raw, " ") {
		scope = strings.TrimSpace(scope)
		if scope == "" {
			continue
		}
		slice = append(slice, scope)
	}
	slices.Sort(slice)
	slicelength := len(slice)
	slicestr := strings.Join(slice, " ")

	set := mapset.NewSet(slice...)
	setslice := set.ToSlice()
	slices.Sort(setslice)
	setstr := strings.Join(setslice, " ")
	setlength := set.Cardinality()
	return &Scope{set, setstr, setlength, slice, slicestr, slicelength}
}

func NewWith(values ...string) *Scope {
	return New(strings.Join(values, " "))
}

func (s *Scope) Copy() *Scope {
	cs := New(s.slicestr)
	return cs
}

func (s *Scope) Slice() []string {
	return s.slice
}

func (s *Scope) SliceString() string {
	return s.slicestr
}

func (s *Scope) SliceLength() int {
	return s.slicelength
}

func (s *Scope) Set() mapset.Set[string] {
	return s.set
}

func (s *Scope) SetString() string {
	return s.setstr
}

func (s *Scope) SetLength() int {
	return s.setlength
}

func (s *Scope) Contains(scope string) bool {
	return s.set.Contains(scope)
}
