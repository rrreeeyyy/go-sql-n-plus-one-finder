# go-sql-n-plus-one-finder (np1finder)

**np1finder** is a simple and efficient tool to detect N+1 query problems in Go applications that use `database/sql`. It helps identify query patterns that can lead to performance bottlenecks, particularly in scenarios where multiple database queries are repeatedly made in loops, causing unnecessary database load.

## Installation

To install **np1finder**, run:

```bash
go install github.com/your-username/go-sql-n-plus-one-finder@latest
```

This will download and install the latest version of **np1finder**.

## Usage

**np1finder** supports integration with both `http.Handler` and `echo` to detect N+1 queries in Go web applications. Here’s how to set it up:

### Basic Setup

1. Import **np1finder** and **go-sql-proxy**.
2. Initialize **np1finder** with your preferred configuration.
3. Register the database driver with **np1finder** hooks.

Example:

```go
import (
	"context"
	"database/sql"
	"os"

	"github.com/rrreeeyyy/go-sql-n-plus-one-finder/pkg/np1finder"
	proxy "github.com/shogo82148/go-sql-proxy"
	"github.com/jmoiron/sqlx"
	"github.com/go-sql-driver/mysql"
	"golang.org/x/exp/slog"
)

func main() {
	ctx := context.Background()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{})) // JSON logger setup

	// Configure np1finder with desired settings
	finder := np1finder.NewFinder(np1finder.Config{
		Context:   ctx,
		Logger:    logger,
		Threshold: 2, // Set the threshold for potential N+1 query detection
	})

	// Register the MySQL driver with np1finder hooks
	sql.Register("mysql:np1finder", proxy.NewProxyContext(&mysql.MySQLDriver{}, finder.Newn-plusOneFinderHooksContext(ctx)))

	// Open the database connection with the np1finder-enabled driver
	dsn := "your-database-dsn-here"
	db, err := sqlx.Open("mysql:np1finder", dsn)
	if err != nil {
		// Handle error
	}
	defer db.Close()

	// Use db for database operations
}
```

### Using np1finder with `http.Handler`

To use **np1finder** as middleware with `http.Handler`, wrap your handler function as follows:

```go
http.Handle("/", finder.HTTPHandlerNP1FinderMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	// Your handler logic here
})))
```

In this setup, **np1finder** will monitor HTTP requests for potential N+1 queries, logging any detected issues based on the configured threshold.

### Using np1finder with `echo`

To use **np1finder** with the `echo` web framework, apply the middleware using `finder.EchoNP1FinderMiddleware()`:

```go
import (
	"github.com/labstack/echo/v4"
)

func main() {
	e := echo.New()

	// Use np1finder middleware with echo
	e.Use(finder.EchoNP1FinderMiddleware())

	// Define routes and start server
}
```

With `echo`, **np1finder** middleware will automatically monitor all routes for N+1 query patterns.

### Log Output

When **np1finder** detects N+1 queries, it logs warnings with relevant details, including the query, count, URI, and source location in the code. Below is an example of the JSON-formatted log output:

```json
{
  "time": "2024-10-31T16:58:14.759932+09:00",
  "level": "WARN",
  "msg": "N+1 Query Detected",
  "query": "select * from posts where user_id = ?",
  "count": 2,
  "uri": "/",
  "caller": "/Users/rrreeeyyy/src/github.com/rrreeeyyy/go-sql-n-plus-one-finder/example/echo/main.go:68"
}
{
  "time": "2024-10-31T16:58:14.760404+09:00",
  "level": "WARN",
  "msg": "N+1 Query Detected",
  "query": "select * from comments where post_id = ?",
  "count": 3,
  "uri": "/",
  "caller": "/Users/rrreeeyyy/src/github.com/rrreeeyyy/go-sql-n-plus-one-finder/example/echo/main.go:76"
}
```

In these examples:
- `"query"` shows the detected N+1 query pattern.
- `"count"` is the number of times this query was repeated in the same request.
- `"uri"` is the request path associated with the query.
- `"caller"` is the location in the source code where the query was executed.

**Note**: The `caller` field is currently obtained using a workaround, which may occasionally lead to inaccurate results. Please use this information as a general guide, and verify manually if the location appears incorrect.

## License

This project is licensed under the MIT License. See [LICENSE](./LICENSE) for more details.
