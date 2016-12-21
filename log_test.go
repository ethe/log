package log

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFileLine(t *testing.T) {
	buf := bytes.NewBuffer(make([]byte, 100))
	l := new(Logger)
	l.name = "name"
	l.lv = INFO
	l.handlers = make(map[Handler]bool)
	l.recordFactory = NewBaseRecordFactory()

	f := NewBaseFormatter(false)
	if err := f.ParseFormat("{{level}} {{date}} {{time}} {{name}} {{file_line}} {{}}"); err != nil {
		t.Fatalf("NewStreamHandler Error:%v", err)
	}
	hdr := NewStreamHandler(buf, f)

	l.AddHandler(hdr)
	SetGlobalAppID("samaritan.test")
	defer SetGlobalAppID("")
	l.Info("TEST_TEST")

	strs := strings.Split(buf.String(), " ")
	if strs[4] != "log_test.go:34" {
		t.Errorf("FileLine Error: %s", buf.String())
	}
}

func newLogger(t *testing.T, w io.Writer, f string) SimpleLogger {
	l := NewWithWriter("test", nil)
	formatter := NewBaseFormatter(false)
	if err := formatter.ParseFormat(f); err != nil {
		t.Error("error creating stream handler: ", err)
		t.FailNow()
	}
	h := NewStreamHandler(w, formatter)

	h.Colored(false)
	l.AddHandler(h)
	return l
}

func TestAsync(t *testing.T) {
	w := &fakeWriter{
		writed: make(chan bool, 10),
		buf:    bytes.NewBuffer(make([]byte, 0)),
	}
	l := newLogger(t, w, "{{}}")

	w1 := &fakeWriter{
		writed: make(chan bool, 10),
		buf:    bytes.NewBuffer(make([]byte, 0)),
	}
	f := NewBaseFormatter(false)
	f.ParseFormat("{{}}")
	h := NewStreamHandler(w1, f)
	h.Colored(false)
	l.AddHandler(h)

	l.SetAsync(true)
	expected := "Test_Async\n"
	l.Info("Test_Async")

	realWrited := 0
outer:
	for realWrited < 2 {
		select {
		case <-w.writed:
			realWrited++
		case <-w1.writed:
			realWrited++
		case <-time.After(time.Millisecond * 200):
			t.Errorf("Test Async Timeout, realWrited=%d", realWrited)
			break outer
		}
		if realWrited == 2 {
			if w.String() != expected || w1.String() != expected {
				t.Errorf("Test Async Error, want=%s want_w1=%s got=%s", w.String(), w1.String(), expected)
			}
		}
	}
}

func TestAsyncFATA(t *testing.T) {
	w := &appendWriter{
		l: &sync.Mutex{},
	}

	SetBufferSize(3000)

	l := newLogger(t, w, "{{}}")
	l.SetAsync(true)
	realLogger := l.(*Logger)

	expected := "Last_Fatal_Async\n"
	fataRecord := realLogger.recordFactory(realLogger.Name(), 2, FATA, "Last_Fatal_Async")

	max := 2000

	for i := 0; i < max; i++ {
		realLogger.Info("Test_Async")
	}
	<-time.After(time.Millisecond * 100)
	realLogger.Output(fataRecord)
	<-time.After(time.Millisecond * 100)
	for i := 0; i < 20; i++ {
		realLogger.Info("Test_Async")
	}
	<-time.After(time.Millisecond * 100)

	if len(w.lines) != max+1 {
		t.Errorf("Writed log lines not Match, whant=%d got=%d", max+1, len(w.lines))
	} else if w.lines[max] != expected {
		t.Errorf("Fatal Log not Match,\n want=%s got=%s", expected, w.lines[max])
	}
}

func TestGlobalLevel(t *testing.T) {
	expected := "W: WarnLog\n"
	var b bytes.Buffer
	l := newLogger(t, &b, "{{l}}: {{}}")
	SetGlobalLevel(WARN)
	defer SetGlobalLevel(NOTSET)

	l.Debug("DebugLog")
	l.Info("InfoLog")
	l.Warn("WarnLog")

	if b.String() != expected {
		t.Errorf("Expected:\n%v\nGot:\n%v", expected, b.String())
	}
}

func TestLevelPriority(t *testing.T) {
	l := New("test")
	ast := assert.New(t)

	ast.Equal(l.Level(), defaultLevel)

	SetGlobalLevel(FATA)
	defer SetGlobalLevel(NOTSET)
	ast.Equal(l.Level(), FATA)

	l.SetLevel(WARN)
	ast.Equal(l.Level(), WARN)
}

func TestLevel(t *testing.T) {
	expected := "E: ErrorLog\n"
	var b bytes.Buffer
	l := newLogger(t, &b, "{{l}}: {{}}")
	l.SetLevel(ERRO)
	l.Debug("DebugLog")
	l.Info("InfoLog")
	l.Warn("WarnLog")
	l.Error("ErrorLog")

	if b.String() != expected {
		t.Errorf("Expected:\n%v\nGot:\n%v", expected, b.String())
	}
}

func TestGlobalAppID(t *testing.T) {
	var buf bytes.Buffer
	l := newLogger(t, &buf, "[{{app_id}}] ## {{}}")

	expectedNil := "[-] ## InfoLog\n"
	l.Info("InfoLog")
	if buf.String() != expectedNil {
		t.Errorf("Expected:\n%s\nGot:\n%s", expectedNil, buf.String())
	}

	buf.Reset()
	expected := "[test.appid] ## InfoLog\n"
	SetGlobalAppID("test.appid")
	defer SetGlobalAppID("")

	l.Info("InfoLog")
	if buf.String() != expected {
		t.Errorf("Expected:\n%s\nGot:\n%s", expected, buf.String())
	}
}

func TestTemplate(t *testing.T) {
	expected := `long: INFO
short: I
duplicate: I
content: hi
`
	var b bytes.Buffer
	l := newLogger(t, &b, "long: {{ level }}\nshort: {{     l }}\nduplicate: {{l}}\ncontent: {{}}")
	l.Info("hi")

	if b.String() != expected {
		t.Errorf("Expected:\n%v\nGot:\n%v\n", expected, b.String())
	}
}

func ExampleLogger() {
	l := NewWithWriter("test", nil)
	f := NewBaseFormatter(false)
	if err := f.ParseFormat("{{level}} {{}}"); err != nil {
		fmt.Println("error creating stream handler: ", err)
		return
	}
	h := NewStreamHandler(os.Stdout, f)

	h.Colored(false)
	l.AddHandler(h)

	l.Debug("default level is INFO")
	l.Info("thus debug is not printed")

	l.SetLevel(DEBUG)
	l.Debug("this enables debug")
	// Output:
	// INFO thus debug is not printed
	// DEBUG this enables debug
}

func ExampleLevel() {
	l := NewWithWriter("test", nil)
	l.SetLevel(DEBUG)

	f := NewBaseFormatter(false)
	if err := f.ParseFormat("{{level}} {{}}"); err != nil {
		fmt.Println("error creating stream handler: ", err)
		return
	}
	h := NewStreamHandler(os.Stdout, f)
	h.Colored(false)
	l.AddHandler(h)
	l.Debug("Debug, turned off by default")
	l.Info("Info, default log level")
	l.Warn("Warning, errors are handled, need attention")
	// l.Fatal("Fatal, will os.Exit(1)")
	// FATA Fatal, will os.Exit(1)

	// Output:
	// DEBUG Debug, turned off by default
	// INFO Info, default log level
	// WARN Warning, errors are handled, need attention
}

// Benchmarks
func dateM(n time.Time) string {
	year, month, day := n.Date()
	return fmt.Sprintf("%04d-%02d-%02d", year, month, day)
}

func timeM(n time.Time) string {
	hour, min, sec := n.Clock()
	return fmt.Sprintf("%02d:%02d:%02d", hour, min, sec)
}

func BenchmarkDateM(b *testing.B) {
	n := time.Now()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dateM(n)
	}
}

func BenchmarkDate(b *testing.B) {
	hdr := defaultLogger.Handlers()[0].(*StreamHandler)
	r := &BaseRecord{
		now: time.Now(),
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hdr.Formatter.(*BaseFormatter)._date(r)
	}
}

func BenchmarkTimeM(b *testing.B) {
	n := time.Now()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		timeM(n)
	}
}

func BenchmarkTime(b *testing.B) {
	hdr := defaultLogger.Handlers()[0].(*StreamHandler)
	r := &BaseRecord{
		now: time.Now(),
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hdr.Formatter.(*BaseFormatter)._time(r)
	}
}

type emptyWriter struct {
	sync.Mutex
	times     int
	waitTimes int
	writed    chan bool
	w         io.WriteCloser
}

func newEmptyWriter(N int, key int) (*emptyWriter, error) {
	filename := "/tmp/empty_writer" + strconv.Itoa(key)
	f, err := os.OpenFile(filename, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}

	ew := &emptyWriter{
		sync.Mutex{},
		0,
		N,
		make(chan bool, 10),
		f,
	}
	return ew, nil
}

func (w *emptyWriter) Write(p []byte) (n int, err error) {

	n, err = w.w.Write(p)

	w.Lock()
	w.times++
	if w.times == w.waitTimes {
		w.writed <- true
	}
	w.Unlock()
	return
}

func initLockVisor() {
	writerLocks = newWriterLocker()
	wSupervisor = newWriterSupervisor(defaultBufferSize)
}

func BenchmarkLogSync(b *testing.B) {
	initLockVisor()
	w, err := newEmptyWriter(b.N, 11)
	if err != nil {
		b.Error(err)
	}
	defer w.w.Close()

	l := NewWithWriter("test", nil)
	f := NewBaseFormatter(false)
	f.ParseFormat("{{}}")
	h := NewStreamHandler(w, f)
	h.Colored(false)
	l.AddHandler(h)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.Info("TEST_TEST_TEST")
	}
}

func BenchmarkLogAsync(b *testing.B) {
	initLockVisor()
	w, err := newEmptyWriter(b.N, 12)
	if err != nil {
		b.Error(err)
	}
	defer w.w.Close()

	l := NewWithWriter("test", nil)
	f := NewBaseFormatter(false)
	f.ParseFormat("{{}}")
	h := NewStreamHandler(w, f)
	h.Colored(false)
	l.AddHandler(h)
	l.SetAsync(true)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.Info("TEST_TEST_TEST")
	}
}

func BenchmarkLogSync5Handlers(b *testing.B) {
	initLockVisor()
	var arr [5]*emptyWriter
	var err error
	for i := range arr {
		arr[i], err = newEmptyWriter(b.N, i)
		if err != nil {
			b.Error(err)
		}
		defer arr[i].w.Close()
	}

	l := NewWithWriter("test", nil)
	for _, w := range arr {
		f := NewBaseFormatter(false)
		f.ParseFormat("{{}}")
		h := NewStreamHandler(w, f)
		h.Colored(false)
		l.AddHandler(h)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.Info("TEST_TEST_TEST")
	}
}

func BenchmarkLogAsync5Handlers(b *testing.B) {
	initLockVisor()
	arr := make([]*emptyWriter, 5)
	var err error
	for i := range arr {
		arr[i], err = newEmptyWriter(b.N, i)
		if err != nil {
			b.Error(err)
		}
		defer arr[i].w.Close()
	}

	l := NewWithWriter("test", nil)
	for _, w := range arr {
		f := NewBaseFormatter(false)
		f.ParseFormat("{{}}")
		h := NewStreamHandler(w, f)
		h.Colored(false)
		l.AddHandler(h)
	}
	l.SetAsync(true)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.Info("TEST_TEST_TEST")
	}
}
