package crdt

import "fmt"

type SiteID uint32
type Counter uint32

type Dot struct {
	SiteID  SiteID
	Counter Counter
}

func (d Dot) Cmp(dot *Dot) int {
	if d.SiteID < dot.SiteID {
		return -1
	}
	if d.SiteID > dot.SiteID {
		return 1
	}
	if d.Counter < dot.Counter {
		return -1
	}
	if d.Counter > dot.Counter {
		return 1
	}
	return 0
}

type Summary struct {
	summary map[SiteID]Counter
}

func NewSummary() *Summary {
	return &Summary{
		summary: make(map[SiteID]Counter),
	}
}

func (s *Summary) Clone() *Summary {
	summary := make(map[SiteID]Counter, len(s.summary))
	for i, c := range s.summary {
		summary[i] = c
	}
	return &Summary{summary}
}

func (s *Summary) Counter(siteID SiteID) Counter {
	c, ok := s.summary[siteID]
	if ok {
		return c
	}
	return 0
}

func (s *Summary) Dot(siteID SiteID) Dot {
	c := s.Increment(siteID)
	return Dot{siteID, c}
}

func (s *Summary) Increment(siteID SiteID) Counter {
	c, ok := s.summary[siteID]
	if ok {
		c = c + 1
	} else {
		c = 1
	}
	s.summary[siteID] = c
	return c
}

func (s *Summary) Contains(dot *Dot) bool {
	c, ok := s.summary[dot.SiteID]
	if ok {
		return c >= dot.Counter
	}
	return false
}

func (s *Summary) Insert(dot *Dot) {
	c, ok := s.summary[dot.SiteID]
	if !ok || c < dot.Counter {
		s.summary[dot.SiteID] = dot.Counter
	}
}

func (s *Summary) Eq(summary *Summary) bool {
	if len(s.summary) != len(summary.summary) {
		return false
	}
	for id, c1 := range s.summary {
		c2, ok := summary.summary[id]
		if !ok || c1 != c2 {
			return false
		}
	}
	return true
}

func CheckSiteID(siteID SiteID) {
	if uint32(siteID) == 0 {
		panic(fmt.Sprintf("Invalid site id: %d", siteID))
	}
}

func DotExists(dots []Dot, dot Dot) bool {
	for _, d := range dots {
		if dot == d {
			return true
		}
	}
	return false
}
