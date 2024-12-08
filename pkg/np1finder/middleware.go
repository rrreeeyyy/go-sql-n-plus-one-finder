package np1finder

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func (f *Finder) HTTPHandlerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f.mutex.Lock()
		defer f.mutex.Unlock()

		f.Scan(r.RequestURI)
		next.ServeHTTP(w, r)
		f.Finish()
	})
}

func (f *Finder) EchoMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			f.mutex.Lock()
			defer f.mutex.Unlock()

			f.Scan(c.Request().RequestURI)
			err := next(c)
			f.Finish()

			return err
		}
	}
}

func (f *Finder) ChiMiddleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			f.mutex.Lock()
			defer f.mutex.Unlock()

			f.Scan(r.RequestURI)
			next.ServeHTTP(w, r)
			f.Finish()
		}

		return http.HandlerFunc(fn)
	}
}
