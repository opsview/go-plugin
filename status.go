package plugin

type Status int

const (
	OK Status = iota
	WARNING
	CRITICAL
	UNKNOWN
)

func (st Status) ExitCode() int {
	return int(st)
}

func (st Status) String() string {
	switch st {
	case 0:
		return "OK"
	case 1:
		return "WARNING"
	case 2:
		return "CRITICAL"
	default:
		return "UNKNOWN"
	}
}
