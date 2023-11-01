package membership

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// Member is a struct
type Member struct {
	ID             string
	Heartbeat      int
	LastUpdateTime int64
	State          State
	Incarnation    int
}

// NewMemberSelf creates a new member of itself information
func NewMemberSelf() (*Member, error) {
	name, err := os.Hostname()
	if err != nil {
		return nil, fmt.Errorf("failed to get hostname: %w", err)
	}
	return &Member{
		ID:             fmt.Sprintf("%s_%d", name, time.Now().Unix()),
		Heartbeat:      0,
		LastUpdateTime: time.Now().UnixMilli(),
		State:          ALIVE,
		Incarnation:    0,
	}, nil
}

func (m *Member) UpdateState(heartbeat int, state State) {
	m.logChange(state, heartbeat, m.Incarnation)
	m.Heartbeat = heartbeat
	m.State = state
	m.LastUpdateTime = time.Now().UnixMilli()
}

func (m *Member) UpdateHeartbeatAndIncarnation(heartbeat, incarnation int) {
	m.logChange(m.State, heartbeat, incarnation)
	m.Heartbeat = heartbeat
	m.Incarnation = incarnation
	m.LastUpdateTime = time.Now().UnixMilli()
}

func (m *Member) UpdateStateAndIncarnation(state State, incarnation int) {
	m.logChange(state, m.Heartbeat, incarnation)
	m.State = state
	m.Incarnation = incarnation
	m.LastUpdateTime = time.Now().UnixMilli()
}

func (m *Member) UpdateStateHeartbeatAndIncarnation(state State, heartbeat, incarnation int) {
	m.logChange(state, heartbeat, incarnation)
	m.State = state
	m.Heartbeat = heartbeat
	m.Incarnation = incarnation
	m.LastUpdateTime = time.Now().UnixMilli()
}

func (m *Member) IncreaseHeartbeat() {
	m.logChange(m.State, m.Heartbeat+1, m.Incarnation)
	m.Heartbeat++
	m.LastUpdateTime = time.Now().UnixMilli()
}

func (m *Member) logChange(state State, heartbeat, incarnation int) {
	logrus.WithFields(logrus.Fields{
		"ID": m.ID,
	}).Debugf(
		"\n\tState: %s -> %s\n\tHeartbeat: %d -> %d\n\tIncarnation: %d -> %d",
		m.State,
		state,
		m.Heartbeat,
		heartbeat,
		m.Incarnation,
		incarnation,
	)
	if m.State != state || m.Incarnation != incarnation {
		logrus.WithFields(logrus.Fields{
			"ID": m.ID,
		}).Infof("[STATE CHANGE] (%s, %d) to (%s, %d)", m.State, m.Incarnation, state, incarnation)
	}
}

func (m *Member) GetName() string {
	return strings.Split(m.ID, "_")[0]
}

func (m *Member) String() string {
	return fmt.Sprintf("\n{\n\tID: %s\n\tHeartbeat: %d\n\tLastUpdateTime: %d\n\tState: %s\n\tIncarnation: %d\n}", m.ID, m.Heartbeat, m.LastUpdateTime, m.State, m.Incarnation)
}
