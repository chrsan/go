package crdt

type SiteID uint32
type Counter uint32

type Dot struct {
	SiteID  SiteID
	Counter Counter
}

func (d Dot) Cmp(dot Dot) int {
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
	return Dot{
		SiteID:  siteID,
		Counter: c,
	}
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

func (s *Summary) Contains(dot Dot) bool {
	return s.ContainsPair(dot.SiteID, dot.Counter)
}

func (s *Summary) ContainsPair(siteId SiteID, counter Counter) bool {
	c, ok := s.summary[siteId]
	if ok {
		return c >= counter
	}
	return false
}

func (s *Summary) Insert(dot Dot) {
	s.InsertPair(dot.SiteID, dot.Counter)
}

func (s *Summary) InsertPair(siteId SiteID, counter Counter) {
	c, ok := s.summary[siteId]
	if !ok || c < counter {
		s.summary[siteId] = counter
	}
}
