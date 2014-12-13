package overlay

import (
	"log"
	"time"
)

type Liveness struct {
	members map[int64]time.Time
}

func NewLiveness() *Liveness {
	l := new(Liveness)
	l.members = make(map[int64]time.Time)
	return l
}

func (l *Liveness) SetAlive(m *Member) {
	l.members[m.GetId()] = time.Now()
	log.Printf("Member %d is alive", m.GetId())
}

func (l *Liveness) Check(members []*Member, onTimeout func(*Member)) {
	for _, m := range members {
		if time.Since(l.members[m.GetId()]) > 10*time.Second {
			onTimeout(m)
		}
	}
}
