package logging

import (
	"log"
	"os"
	"strings"
	"sync"
	"time"
)

var L Logger

type Logger struct {
	l  *log.Logger
	mu *sync.Mutex
}

func Init() {
	if L.l != nil {
		L.Warn("Multiple calls to L.Init()")
		return
	}

	l := log.New(os.Stdout, "", 0)
	L = Logger{l, &sync.Mutex{}}

	// ===========================================
	// NO LOGGING ABOVE THIS LINE
	// ===========================================

	L.Info("Logging initialized...")
}

func (L Logger) Info(msg ...string) {
	if len(msg) == 0 {
		return
	}

	L.print("[ERROR]", msg...)
}
func (L Logger) Warn(msg ...string) {
	if len(msg) == 0 {
		return
	}

	L.print("[ERROR]", msg...)
}
func (L Logger) Error(msg ...string) {
	if len(msg) == 0 {
		return
	}

	L.print("[ERROR]", msg...)
}

func (L Logger) print(lvl string, msg ...string) {
	L.mu.Lock()
	defer L.mu.Unlock()

	if L.l == nil {
		panic("uninitialized logger")
	}

	timeNow := time.Now().Format(time.RFC822Z)
	sb := strings.Builder{}

	for _, m := range msg {
		_, err := sb.WriteString(m)
		if err != nil {
			L.l.Printf("%s\t[ERROR]\t%s%s\n", timeNow, "failed logging: ", m)
			return
		}
		_, err = sb.WriteString(" ")
		if err != nil {
			L.l.Printf("%s\t[ERROR]\t%s\n", timeNow, "failed logging: ' '")
			return
		}
	}

	L.l.Printf("%s\t%s\t%s\n", timeNow, lvl, sb.String())
}
