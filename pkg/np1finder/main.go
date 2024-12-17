package np1finder

import (
	"context"
	"database/sql/driver"
	"log/slog"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"

	"github.com/percona/go-mysql/query"
	proxy "github.com/shogo82148/go-sql-proxy"
)

type Config struct {
	Context       context.Context
	Logger        *slog.Logger
	Threshold     int
	PackageFilter []string
}

type Finder struct {
	ctx           context.Context
	logger        *slog.Logger
	threashold    int
	uri           string
	channel       chan Message
	channelClosed bool
	queries       []string
	counter       sync.Map
	caller        sync.Map
	filter        []string
	mu            sync.Mutex
}

type Message struct {
	query string
	frame *runtime.Frame
}

func NewFinder(config Config) *Finder {
	if config.Context == nil {
		config.Context = context.TODO()
	}

	if config.Logger == nil {
		config.Logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{}))
	}

	if config.Threshold == 0 {
		config.Threshold = 2
	}

	if config.PackageFilter == nil {
		config.PackageFilter = DefaultPackageFilter()
	}

	return &Finder{
		ctx:        config.Context,
		logger:     config.Logger,
		threashold: config.Threshold,
		channel:    make(chan Message),
		queries:    []string{},
		filter:     config.PackageFilter,
	}
}

func DefaultPackageFilter() []string {
	return []string{
		"runtime",
		"database/sql",
		"github.com/rrreeeyyy/go-sql-n-plus-one-finder/pkg/np1finder",
		"github.com/shogo82148/go-sql-proxy",
		"github.com/jmoiron/sqlx",
	}
}

func (f *Finder) Scan(uri string) {
	f.uri = uri
	f.channel = make(chan Message)
	f.channelClosed = false

	go func() {
		for msg := range f.channel {
			f.queries = append(f.queries, msg.query)
			count, _ := f.counter.LoadOrStore(msg.query, 0)
			f.counter.Store(msg.query, count.(int)+1)
			if _, ok := f.caller.Load(msg.query); !ok {
				f.caller.Store(msg.query, msg.frame)
			}
		}
	}()
}

func (f *Finder) Finish() {
	f.counter.Range(func(key, value interface{}) bool {
		q := key.(string)
		c := value.(int)
		if c >= f.threashold {
			frame, _ := f.caller.Load(q)
			f.logger.Warn(
				"N+1 Query Detected",
				slog.String("query", q),
				slog.Int("count", c),
				slog.String("uri", f.uri),
				slog.String("caller", strings.Join([]string{frame.(*runtime.Frame).File, strconv.Itoa(frame.(*runtime.Frame).Line)}, ":")),
			)
		}
		return true
	})

	f.uri = ""
	f.counter = sync.Map{}
	f.queries = []string{}
	f.caller = sync.Map{}

	f.mu.Lock()
	defer f.mu.Unlock()
	if !f.channelClosed {
		close(f.channel)
		f.channelClosed = true
	}
}

func (f *Finder) NewHooksContext() *proxy.HooksContext {
	return &proxy.HooksContext{
		Query: func(_ context.Context, _ interface{}, stmt *proxy.Stmt, args []driver.NamedValue, rows driver.Rows) error {
			if stmt == nil {
				return nil
			}

			f.mu.Lock()
			defer f.mu.Unlock()
			if !f.channelClosed {
				select {
				case f.channel <- Message{query: query.Fingerprint(stmt.QueryString), frame: f.findCaller()}:
				default:
				}
			}

			return nil
		},
	}
}

func (f *Finder) findCaller() *runtime.Frame {
	// skip 3 frames to get the caller of the function calling this function
	// 0: runtime.Callers, 1: findCaller, 2: NewHooksContext
	skip := 3
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

			if f.callerFilter(name) {
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

func (f *Finder) callerFilter(name string) bool {
	for _, filter := range f.filter {
		if strings.HasPrefix(name, filter) {
			return true
		}
	}
	return false
}
