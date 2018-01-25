package rpc

import (
	"regexp"
	"testing"

	"github.com/eleme/log"
)

const (
	NexLogPattern    = `(\d{4}-\d{2}-\d{2} \d{1,}:\d{1,}:\d{1,}.\d{0,}) info (\S+)\[(\S+)\]: \[(\S+) (.*)\] (.*)## (.+)$`
	NexSyslogPattern = `\[(\S+) (.*)\] (.*)## (.+)$`
)

func TestELogFormatter(t *testing.T) {
	nexLogRegex, _ := regexp.Compile(NexLogPattern)
	nexSyslogRegex, _ := regexp.Compile(NexSyslogPattern)
	record := NewELogRecord("test", 1, log.INFO, "test", "111", "111")

	formatter := NewELogFormatter(NexLog, false)
	nexLog := formatter.Format(record)
	if !nexLogRegex.Match(nexLog) {
		t.Error("Nex log pattern failed.")
	}

	formatter = NewELogFormatter(NexSyslog, false)
	nexSyslog := formatter.Format(record)
	if !nexSyslogRegex.Match(nexSyslog) {
		t.Error("Nex syslog pattern failed.")
	}
}
