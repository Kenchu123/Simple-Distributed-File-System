package membership

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"gitlab.engr.illinois.edu/ckchu2/cs425-mp3/internal/config"
)

// MemberList is a struct that contains a map of Members
type Membership struct {
	ID           string             // ID of the node
	Members      map[string]*Member // map of members
	mu           sync.Mutex         // mutex
	rrobin       []*Member
	rrobinIndex  int    // round robin index
	targetNumber int    // number of targets
	introducer   string // whether the node is the introducer
}

// New creates a new membership
func New(introducer string, targetNumber int) (*Membership, error) {
	member, err := NewMemberSelf()
	if err != nil {
		return nil, fmt.Errorf("failed to create a new membership: %w", err)
	}
	if strings.Contains(member.ID, introducer) {
		logrus.Infof("I am the introducer")
	}
	return &Membership{
		ID: member.ID, Members: map[string]*Member{
			member.ID: member,
		},
		rrobin:       []*Member{},
		rrobinIndex:  0,
		targetNumber: targetNumber,
		introducer:   introducer,
	}, nil
}

// NewEmpty creates a new empty membership
func NewEmpty() *Membership {
	return &Membership{Members: map[string]*Member{}}
}

// IncreaseSelfHeartbeat increases the heartbeat of itself
func (m *Membership) IncreaseSelfHeartbeat() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Members[m.ID].IncreaseHeartbeat()
}

func (m *Membership) UpdateSelfState(state State) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Members[m.ID].UpdateState(m.Members[m.ID].Heartbeat, state)
}

// Update updates the membership list
func (m *Membership) Update(ms *Membership) {
	m.mu.Lock()
	defer m.mu.Unlock()
	// Iterate through the members in the membership list
	for _, member := range ms.Members {
		// Update the membership list with the member
		m.updateMember(member)
	}
}

// UpdateMember updates the membership list with a new member
func (m *Membership) updateMember(member *Member) {
	// case 1: member is not in the membership list
	if _, ok := m.Members[member.ID]; !ok {
		// case 1.1: new member is marked as failed or left, don't update the state
		if member.State == FAILED || member.State == LEFT {
			return
		}
		member.LastUpdateTime = time.Now().UnixMilli()
		m.Members[member.ID] = member
		// TODO: prettier log
		logrus.Infof("[JOINED] %s with state %s", member.ID, member.State)
		return
	}
	// case 2: member is in the membership list
	// case 2.1: member is marked as failed, don't update the state
	if m.Members[member.ID].State == FAILED {
		return
	}
	// case 2.2: member is marked as left, don't update the state
	if m.Members[member.ID].State == LEFT {
		if member.State == FAILED {
			m.Members[member.ID].UpdateState(member.Heartbeat, member.State)
			// case self failed
			if member.State == FAILED && m.ID == member.ID {
				logrus.Fatalf("[FAILED] I am marked as failed")
			}
		}
		return
	}
	// case 2.3: member is marked as alive
	if m.Members[member.ID].State == ALIVE {
		// case 2.3.1: new member is marked as failed or left, update the state
		if member.State == FAILED || member.State == LEFT {
			m.Members[member.ID].UpdateState(member.Heartbeat, member.State)
			// case self failed
			if member.State == FAILED && m.ID == member.ID {
				logrus.Fatalf("[FAILED] I am marked as failed")
			}
		}
		// case 2.3.2: new member is marked as alive with higher heartbeat number and with equal or higher incarnation number, update the heartbeat
		if member.State == ALIVE && m.Members[member.ID].Heartbeat < member.Heartbeat && m.Members[member.ID].Incarnation <= member.Incarnation {
			m.Members[member.ID].UpdateHeartbeatAndIncarnation(member.Heartbeat, member.Incarnation)
		}
		// case 2.3.4: new member is marked as suspected with equal or higher incarnation number, update state, and incarnation number
		if member.State == SUSPECTED && m.Members[member.ID].Incarnation <= member.Incarnation {
			// case self alive but received suspected
			if m.ID == member.ID {
				m.Members[member.ID].UpdateStateAndIncarnation(ALIVE, member.Incarnation+1)
			} else {
				m.Members[member.ID].UpdateStateAndIncarnation(member.State, member.Incarnation)
			}
		}
		return
	}
	// case 2.4: member is marked as suspected
	if m.Members[member.ID].State == SUSPECTED {
		// case 2.4.1: new member is marked as failed or left, update the state
		if member.State == FAILED || member.State == LEFT {
			m.Members[member.ID].UpdateState(member.Heartbeat, member.State)
			// case self failed
			if member.State == FAILED && m.ID == member.ID {
				logrus.Fatalf("[FAILED] I am marked as failed")
			}
		}
		// case 2.4.2: new member is marked as alive with higher incarnation number, update the state, heartbeat, and incarnation number
		if member.State == ALIVE && m.Members[member.ID].Incarnation < member.Incarnation {
			m.Members[member.ID].UpdateStateHeartbeatAndIncarnation(member.State, member.Heartbeat, member.Incarnation)
		}
		// case 2.4.3: new member is marked as suspected with higher incarnation number, update the incarnation number
		if member.State == SUSPECTED && m.Members[member.ID].Incarnation < member.Incarnation {
			m.Members[member.ID].UpdateStateAndIncarnation(member.State, member.Incarnation)
		}
		return
	}
}

// DetectFailure detects failure
func (m *Membership) DetectFailure(config config.FailureDetect) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, member := range m.Members {
		if member.State == ALIVE {
			if config.Suspicion.Enabled {
				if time.Now().UnixMilli() > member.LastUpdateTime+config.Suspicion.SuspectTimeout.Milliseconds() {
					member.UpdateState(member.Heartbeat, SUSPECTED)
				}
			} else {
				if time.Now().UnixMilli() > member.LastUpdateTime+config.FailureTimeout.Milliseconds() {
					member.UpdateState(member.Heartbeat, FAILED)
				}
			}
			continue
		}
		if member.State == SUSPECTED && time.Now().UnixMilli() > member.LastUpdateTime+config.Suspicion.FailureTimeout.Milliseconds() {
			member.UpdateState(member.Heartbeat, FAILED)
			continue
		}
	}
}

// CleanUp cleans up the membership list
func (m *Membership) CleanUp(cleanupTimeout time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for id, member := range m.Members {
		if (member.State == FAILED || member.State == LEFT) && time.Now().UnixMilli() > member.LastUpdateTime+cleanupTimeout.Milliseconds() {
			delete(m.Members, id)
			logrus.Infof("[REMOVE] %s with state %s", id, member.State)
		}
	}
}

// String
func (m *Membership) String() string {
	return fmt.Sprintf("SelfID: %s\nMembership: %s\n", m.ID, m.Members)
}

// Get name of membership's owner
func (m *Membership) GetName() string {
	return strings.Split(m.ID, "_")[0]
}

// Check whether the hostname is already in the hostnames list
func checkHostname(hostnames []string, hostname string) bool {
	for _, h := range hostnames {
		if h == hostname {
			return true
		}
	}
	return false
}

// Get heartbeat target members' hostnames
func (m *Membership) GetHeartbeatTargetMembers(machines []config.Machine) []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	// update rrobin list according to membership list
	rrobinSet := map[string]bool{}
	for _, member := range m.rrobin {
		rrobinSet[member.ID] = true
	}
	for _, member := range m.Members {
		// check if member is self
		if member.ID == m.ID {
			continue
		}
		// check if member is in rrobinSet
		if _, ok := rrobinSet[member.ID]; ok {
			continue
		}
		// add member to rrobinSet
		rrobinSet[member.ID] = true
		// insert member to rrobin list on index m.index
		m.rrobin = append(m.rrobin[:m.rrobinIndex], append([]*Member{member}, m.rrobin[m.rrobinIndex:]...)...)
	}
	// case 1: if rrobin list is empty, return introducer
	if len(m.rrobin) == 0 {
		return []string{m.introducer}
	}
	// case 2: if rrobin list is not empty, return m.target members
	hostnames := []string{}
	for {
		// if rrobin list is empty, return introducer
		if len(m.rrobin) == 0 {
			return []string{m.introducer}
		}
		// if m.index is out of range, reset m.index to 0
		if m.rrobinIndex >= len(m.rrobin) {
			m.rrobinIndex = 0
		}
		// if the member is not in membership list, delete it from rrobin list
		if _, ok := m.Members[m.rrobin[m.rrobinIndex].ID]; !ok {
			m.rrobin = append(m.rrobin[:m.rrobinIndex], m.rrobin[m.rrobinIndex+1:]...)
			continue
		}
		// if the member start with index 0, shuffle the rrobin list
		if m.rrobinIndex == 0 {
			rand.Shuffle(len(m.rrobin), func(i, j int) {
				m.rrobin[i], m.rrobin[j] = m.rrobin[j], m.rrobin[i]
			})
		}
		// append to hostname
		hostnames = append(hostnames, m.rrobin[m.rrobinIndex].GetName())
		m.rrobinIndex++
		// if hostnames is full, break
		if len(hostnames) == m.targetNumber || len(hostnames) == len(m.rrobin) {
			break
		}
	}
	return hostnames
}

// SerializedMember is a struct that contains the heartbeat, state, and incarnation of a member
// Used for serialization and deserialization
type SerializedMember struct {
	H int
	S State
	I int
}
type SerializedMembership map[string]SerializedMember

// Serialize serializes the membership list
func Serialize(m *Membership) ([]byte, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	members := SerializedMembership{}
	for id, member := range m.Members {
		members[id] = SerializedMember{
			H: member.Heartbeat,
			S: member.State,
			I: member.Incarnation,
		}
	}
	buf, err := json.Marshal(members)
	if err != nil {
		return nil, err
	}
	return buf, nil
}

// Deserialize deserializes the membership list
func Deserialize(b []byte) (*Membership, error) {
	m := &SerializedMembership{}
	err := json.Unmarshal(b, m)
	if err != nil {
		fmt.Printf("failed to deserialize membership from buf %s to membershiplist: %v\n", string(b), m)
		return nil, err
	}
	members := &Membership{Members: map[string]*Member{}}
	for id, member := range *m {
		members.Members[id] = &Member{
			ID:          id,
			Heartbeat:   member.H,
			State:       member.S,
			Incarnation: member.I,
		}
	}
	return members, nil
}

// Get alive members in the membership list
func (m *Membership) GetAliveMembers() map[string]*Member {
	m.mu.Lock()
	defer m.mu.Unlock()
	members := map[string]*Member{}
	for _, member := range m.Members {
		if member.State == ALIVE {
			members[member.GetName()] = member
		}
	}
	return members
}
