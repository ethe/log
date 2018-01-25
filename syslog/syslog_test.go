package syslog

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/eleme/log"
	"github.com/eleme/log/rpc"
)

func TestSyslogtpl(t *testing.T) {
	buf := bytes.NewBuffer(make([]byte, 0, 100))

	l := log.NewWithWriter("name", nil)

	f := rpc.NewELogFormatter(rpc.NexSyslog, false)
	if err := f.ParseFormat(log.TplSyslog); err != nil {
		t.Error(err)
		return
	}

	hdr := log.NewStreamHandler(buf, f)
	l.AddHandler(hdr)

	recordFactory := rpc.NewELogRecordFactory("", "")
	l.SetRecordFactory(recordFactory)

	log.SetGlobalAppID("samaritan.test")
	defer log.SetGlobalAppID("")
	l.Info("TEST_TEST")

	strs := strings.Split(buf.String(), " ")
	fmt.Println(strs[4])

	if strs[0] != "[samaritan.test" || strs[4] != "TEST_TEST\n" ||
		strs[1] != "-" || strs[2] != "-]" {
		t.Errorf("SyslogTpl Error: %s", buf.String())
	}
}
