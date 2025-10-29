package astraid

import (
	"errors"
	"net"
	"os"
	"strconv"
	"sync"
	"time"
)

type Settings struct {
	BitsSequence   int
	BitsMachineID  int
	TimeUnit       time.Duration
	StartTime      time.Time
	MachineID      func() (int, error)
	CheckMachineID func(int) bool
	ClusterIDKey   string
}

type AstraID struct {
	mutex        *sync.Mutex
	bitsTime     int
	bitsSequence int
	bitsMachine  int
	timeUnit     int64
	startTime    int64
	elapsedTime  int64
	sequence     int
	machine      int
	now          func() time.Time
	clusterIDKey string
}

var (
	ErrInvalidBitsTime      = errors.New("bit length for time must be 32 or more")
	ErrInvalidBitsSequence  = errors.New("invalid bit length for sequence number")
	ErrInvalidBitsMachineID = errors.New("invalid bit length for machine id")
	ErrInvalidTimeUnit      = errors.New("invalid time unit")
	ErrInvalidSequence      = errors.New("invalid sequence number")
	ErrInvalidMachineID     = errors.New("invalid machine id")
	ErrStartTimeAhead       = errors.New("start time is ahead")
	ErrOverTimeLimit        = errors.New("over the time limit")
	ErrNoPrivateAddress     = errors.New("no private ip address")
	ErrInvalidClusterID     = errors.New("invalid cluster id: must be 0-15")
)

const (
	defaultTimeUnit        = 1e7 // nsec, i.e. 10 msec
	defaultClusterTimeUnit = 2e7

	defaultBitsTime           = 39
	defaultBitsSequence       = 8
	defaultBitsMachine        = 16
	defaultClusterBitsMachine = 20
)

const (
	defaultClusterIDKey = "CLUSTER_ID"
	defaultEmptyStr     = ""
)

var defaultInterfaceAddrs = net.InterfaceAddrs

type InterfaceAddrs func() ([]net.Addr, error)

func NewDefault() (*AstraID, error) {
	return New(Settings{})
}

func New(st Settings) (*AstraID, error) {
	if st.BitsSequence < 0 || st.BitsSequence > 30 {
		return nil, ErrInvalidBitsSequence
	}
	if st.BitsMachineID < 0 || st.BitsMachineID > 30 {
		return nil, ErrInvalidBitsMachineID
	}
	if st.TimeUnit < 0 || (st.TimeUnit > 0 && st.TimeUnit < time.Millisecond) {
		return nil, ErrInvalidTimeUnit
	}
	if st.StartTime.After(time.Now()) {
		return nil, ErrStartTimeAhead
	}

	sf := new(AstraID)
	sf.mutex = new(sync.Mutex)
	sf.now = time.Now

	sf.clusterIDKey = st.ClusterIDKey
	if sf.clusterIDKey == defaultEmptyStr {
		sf.clusterIDKey = defaultClusterIDKey
	}

	if st.BitsSequence == 0 {
		sf.bitsSequence = defaultBitsSequence
	} else {
		sf.bitsSequence = st.BitsSequence
	}

	if st.BitsMachineID == 0 {
		if os.Getenv(sf.clusterIDKey) != defaultEmptyStr {
			sf.bitsMachine = defaultClusterBitsMachine
		} else {
			sf.bitsMachine = defaultBitsMachine
		}
	} else {
		sf.bitsMachine = st.BitsMachineID
	}

	sf.bitsTime = 63 - sf.bitsSequence - sf.bitsMachine
	if sf.bitsTime < 32 {
		return nil, ErrInvalidBitsTime
	}

	if st.TimeUnit == 0 {
		if os.Getenv(sf.clusterIDKey) != defaultEmptyStr {
			sf.timeUnit = defaultClusterTimeUnit
		} else {
			sf.timeUnit = defaultTimeUnit
		}

	} else {
		sf.timeUnit = int64(st.TimeUnit)
	}

	if st.StartTime.IsZero() {
		sf.startTime = sf.toInternalTime(time.Date(2025, 10, 1, 0, 0, 0, 0, time.UTC))
	} else {
		sf.startTime = sf.toInternalTime(st.StartTime)
	}

	sf.sequence = 1<<sf.bitsSequence - 1

	var err error
	if st.MachineID == nil {
		sf.machine, err = lower16BitPrivateIP(defaultInterfaceAddrs, sf.clusterIDKey)
	} else {
		sf.machine, err = st.MachineID()
	}
	if err != nil {
		return nil, err
	}

	if sf.machine < 0 || sf.machine >= 1<<sf.bitsMachine {
		return nil, ErrInvalidMachineID
	}

	if st.CheckMachineID != nil && !st.CheckMachineID(sf.machine) {
		return nil, ErrInvalidMachineID
	}

	return sf, nil
}

func (sf *AstraID) NextID() (int64, error) {
	maskSequence := 1<<sf.bitsSequence - 1

	sf.mutex.Lock()
	defer sf.mutex.Unlock()

	current := sf.currentElapsedTime()
	if sf.elapsedTime < current {
		sf.elapsedTime = current
		sf.sequence = 0
	} else {
		sf.sequence = (sf.sequence + 1) & maskSequence
		if sf.sequence == 0 {
			sf.elapsedTime++
			overtime := sf.elapsedTime - current
			sf.sleep(overtime)
		}
	}

	return sf.toID()
}

func (sf *AstraID) toInternalTime(t time.Time) int64 {
	return t.UTC().UnixNano() / sf.timeUnit
}

func (sf *AstraID) currentElapsedTime() int64 {
	return sf.toInternalTime(sf.now()) - sf.startTime
}

func (sf *AstraID) sleep(overtime int64) {
	sleepTime := time.Duration(overtime*sf.timeUnit) -
		time.Duration(sf.now().UTC().UnixNano()%sf.timeUnit)
	time.Sleep(sleepTime)
}

func (sf *AstraID) toID() (int64, error) {
	if sf.elapsedTime >= 1<<sf.bitsTime {
		return 0, ErrOverTimeLimit
	}

	return sf.elapsedTime<<(sf.bitsSequence+sf.bitsMachine) |
		int64(sf.sequence)<<sf.bitsMachine |
		int64(sf.machine), nil
}

func privateIPv4(interfaceAddrs InterfaceAddrs) (net.IP, error) {
	as, err := interfaceAddrs()
	if err != nil {
		return nil, err
	}

	for _, a := range as {
		ipnet, ok := a.(*net.IPNet)
		if !ok || ipnet.IP.IsLoopback() {
			continue
		}

		ip := ipnet.IP.To4()
		if isPrivateIPv4(ip) {
			return ip, nil
		}
	}
	return nil, ErrNoPrivateAddress
}

func isPrivateIPv4(ip net.IP) bool {
	return ip != nil &&
		(ip[0] == 10 || ip[0] == 172 && (ip[1] >= 16 && ip[1] < 32) || ip[0] == 192 && ip[1] == 168 || ip[0] == 169 && ip[1] == 254)
}

func lower16BitPrivateIP(interfaceAddrs InterfaceAddrs, clusterIDKey string) (int, error) {

	ip, err := privateIPv4(interfaceAddrs)
	if err != nil {
		return 0, err
	}

	ipSuffix := int(ip[2])<<8 + int(ip[3])

	clusterIDStr := os.Getenv(clusterIDKey)

	if clusterIDStr == defaultEmptyStr {
		return ipSuffix, nil
	}

	clusterID, err := strconv.Atoi(clusterIDStr)
	if err != nil {
		return 0, ErrInvalidClusterID
	}

	if clusterID < 0 || clusterID > 15 {
		return 0, ErrInvalidClusterID
	}

	return (clusterID << 16) | ipSuffix, nil

}

func (sf *AstraID) ToTime(id int64) time.Time {
	return time.Unix(0, (sf.startTime+sf.timePart(id))*sf.timeUnit)
}

func (sf *AstraID) Compose(t time.Time, sequence, machineID int) (int64, error) {
	elapsedTime := sf.toInternalTime(t.UTC()) - sf.startTime
	if elapsedTime < 0 {
		return 0, ErrStartTimeAhead
	}
	if elapsedTime >= 1<<sf.bitsTime {
		return 0, ErrOverTimeLimit
	}

	if sequence < 0 || sequence >= 1<<sf.bitsSequence {
		return 0, ErrInvalidSequence
	}

	if machineID < 0 || machineID >= 1<<sf.bitsMachine {
		return 0, ErrInvalidMachineID
	}

	return elapsedTime<<(sf.bitsSequence+sf.bitsMachine) |
		int64(sequence)<<sf.bitsMachine |
		int64(machineID), nil
}

func (sf *AstraID) Decompose(id int64) map[string]int64 {
	time := sf.timePart(id)
	sequence := sf.sequencePart(id)
	machine := sf.machinePart(id)
	return map[string]int64{
		"id":       id,
		"time":     time,
		"sequence": sequence,
		"machine":  machine,
	}
}

func (sf *AstraID) timePart(id int64) int64 {
	return id >> (sf.bitsSequence + sf.bitsMachine)
}

func (sf *AstraID) sequencePart(id int64) int64 {
	maskSequence := int64((1<<sf.bitsSequence - 1) << sf.bitsMachine)
	return (id & maskSequence) >> sf.bitsMachine
}

func (sf *AstraID) machinePart(id int64) int64 {
	maskMachine := int64(1<<sf.bitsMachine - 1)
	return id & maskMachine
}
