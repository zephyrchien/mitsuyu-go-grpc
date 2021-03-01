package common

import (
	"sync"
	"time"
)

type Statistician struct {
	enable bool

	uplock   sync.Mutex
	downlock sync.Mutex

	uplimit     int
	downlimit   int
	upleft      int
	downleft    int
	uprefresh   time.Time
	downrefresh time.Time

	uptraffic   uint64
	downtraffic uint64
}

func NewStatistician(uplimit, downlimit int) *Statistician {
	return &Statistician{uplimit: uplimit, downlimit: downlimit}
}

func (s *Statistician) Config(b bool) {
	s.enable = b
}

func (s *Statistician) Restore(up, down uint64) {
	s.uptraffic = up
	s.downtraffic = down
}

func (s *Statistician) RecordUplink(n int) {
	if s.enable {
		s.uplock.Lock()
		defer s.uplock.Unlock()
		s.uptraffic += uint64(n)
		if s.uplimit == 0 {
			return
		}
		if s.upleft -= n; s.upleft < 0 {
			now := time.Now()
			if now.Before(s.uprefresh) {
				time.Sleep(s.uprefresh.Sub(now))
			}
			s.upleft = s.uplimit
			s.uprefresh = time.Now().Add(1 * time.Second)
		}
	}
}

func (s *Statistician) RecordDownlink(n int) {
	if s.enable {
		s.downlock.Lock()
		defer s.downlock.Unlock()
		s.downtraffic += uint64(n)
		if s.downlimit == 0 {
			return
		}
		if s.downleft -= n; s.downleft < 0 {
			now := time.Now()
			if now.Before(s.downrefresh) {
				time.Sleep(s.downrefresh.Sub(now))
			}
			s.downleft = s.downlimit
			s.downrefresh = time.Now().Add(1 * time.Second)
		}
	}
}

func (s *Statistician) StartRecord() {
	s.enable = true
}

func (s *Statistician) StopRecord() {
	s.enable = false
}

func (s *Statistician) GetTraffic() (uint64, uint64) {
	return s.uptraffic, s.downtraffic
}
