package rpc

import (
	"fmt"

	"github.com/eleme/log"
)

const (
	NexLog = iota
	NexSyslog
)

// ELogFormatter is the formatter for elog.
type ELogFormatter struct {
	*log.BaseFormatter
	logType int
}

// NewELogFormatter create a ELogFormatter with colored.
func NewELogFormatter(logType int, colored bool) *ELogFormatter {
	ef := new(ELogFormatter)
	ef.BaseFormatter = log.NewBaseFormatter(colored)
	ef.SetColored(colored)
	ef.logType = logType
	return ef
}

// Format formats a Record with set format
func (f *ELogFormatter) Format(record log.Record) []byte {
	var result string

	switch f.logType {
	case NexLog:
		result = fmt.Sprintf(
			"%v %v %v[%v]: [%v %v %v] ## %v",
			f.Datetime(record),
			f.Level(record),
			f.Name(record),
			f.Pid(record),
			f.AppID(record),
			f._rpcID(record.(*ELogRecord)),
			f._requestID(record.(*ELogRecord)),
			record.String(),
		)
	case NexSyslog:
		result = fmt.Sprintf(
			"[%v %v %v] ## %v",
			f.AppID(record),
			f._rpcID(record.(*ELogRecord)),
			f._requestID(record.(*ELogRecord)),
			record.String(),
		)
	default:
	}
	return []byte(result)
}

func (f *ELogFormatter) _rpcID(r *ELogRecord) string {
	s := r.rpcID
	if s == "" {
		s = "-"
	}
	if f.Colored() {
		s = f.Paint(r.Level(), s)
	}
	return s
}

func (f *ELogFormatter) _requestID(r *ELogRecord) string {
	s := r.requestID
	if s == "" {
		s = "-"
	}
	if f.Colored() {
		s = f.Paint(r.Level(), s)
	}
	return s
}
