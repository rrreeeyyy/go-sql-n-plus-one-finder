package np1finder

import (
	"context"
	"database/sql/driver"
	"log/slog"
	"runtime"
	"strconv"
	"strings"
	"sync"

	"github.com/percona/go-mysql/query"
	proxy "github.com/shogo82148/go-sql-proxy"
)

type Config struct {
	Context   context.Context
	Logger    *slog.Logger
	Threshold int
}

type Finder struct {
	ctx        context.Context
	mutex      sync.RWMutex
	logger     *slog.Logger
	threashold int
	uri        string
	channel    chan Message
	queries    []string
	counter    map[string]int
	caller     map[string]*runtime.Frame
}

type Message struct {
	query string
	frame *runtime.Frame
}

func NewFinder(config Config) *Finder {
	return &Finder{
		ctx:        config.Context,
		logger:     config.Logger,
		threashold: config.Threshold,
		channel:    make(chan Message),
		queries:    []string{},
		counter:    make(map[string]int),
		caller:     make(map[string]*runtime.Frame),
	}
}

func (f *Finder) Scan(uri string) {
	f.uri = uri
	f.channel = make(chan Message)

	go func() {
		for msg := range f.channel {
			f.queries = append(f.queries, msg.query)
			f.counter[msg.query]++
			if _, ok := f.caller[msg.query]; !ok {
				f.caller[msg.query] = msg.frame
			}
		}
	}()
}

func (f *Finder) Finish() {
	for q, c := range f.counter {
		if c >= f.threashold {
			f.logger.Warn(
				"N+1 Query Detected",
				slog.String("query", q),
				slog.Int("count", c),
				slog.String("uri", f.uri),
				slog.String("caller", strings.Join([]string{f.caller[q].File, strconv.Itoa(f.caller[q].Line)}, ":")),
			)
		}
	}

	f.uri = ""
	f.counter = make(map[string]int)
	f.queries = []string{}
	f.caller = make(map[string]*runtime.Frame)

	close(f.channel)
}

func (f *Finder) NewHooksContext() *proxy.HooksContext {
	return &proxy.HooksContext{
		Query: func(_ context.Context, _ interface{}, stmt *proxy.Stmt, args []driver.NamedValue, rows driver.Rows) error {
			if stmt == nil {
				return nil
			}

			select {
			case f.channel <- Message{query: query.Fingerprint(stmt.QueryString), frame: findCaller()}:
			default:
			}

			return nil
		},
	}
}

func findCaller() *runtime.Frame {
	// XXX: this is a hacky way to get caller
	skip := 15
	for {
		var rpc [8]uintptr
		var i int
		n := runtime.Callers(skip, rpc[:])
		frames := runtime.CallersFrames(rpc[:])
		for i = 0; ; i++ {
			frame, more := frames.Next()
			if !more {
				break
			}
			name := frame.Function
			if name == "" || strings.HasPrefix(name, "runtime.") {
				continue
			}

			return &frame
		}
		if n < len(rpc) {
			break
		}
		skip += i
	}
	return nil
}
