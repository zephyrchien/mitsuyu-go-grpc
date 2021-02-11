package common

type Statistician struct {
	enable      bool
	done        chan struct{}
	uplink      chan int
	downlink    chan int
	uptraffic   uint64
	downtraffic uint64
}

func NewStatistician() *Statistician {
	uplink := make(chan int, 20)
	downlink := make(chan int, 20)
	return &Statistician{uplink: uplink, downlink: downlink}
}

func (s *Statistician) Config(b bool) {
	s.enable = b
}

func (s *Statistician) Restore(up, down uint64) {
	s.uptraffic = up
	s.downtraffic = down
}

func (s *Statistician) GetUplink() chan int {
	return s.uplink
}
func (s *Statistician) GetDownlink() chan int {
	return s.downlink
}

func (s *Statistician) RecordUplink(n int) {
	if s.enable {
		s.uplink <- n
	}
}

func (s *Statistician) RecordDownlink(n int) {
	if s.enable {
		s.downlink <- n
	}
}

func (s *Statistician) StartRecord() {
	s.done = make(chan struct{}, 0)
	for {
		select {
		case <-s.done:
			return
		case n := <-s.uplink:
			s.uptraffic += uint64(n)
		case n := <-s.downlink:
			s.downtraffic += uint64(n)
		}
	}
}

func (s *Statistician) StopRecord() {
	defer func() {
		recover()
	}()
	close(s.done)
}

func (s *Statistician) GetTraffic() (uint64, uint64) {
	return s.uptraffic, s.downtraffic
}
