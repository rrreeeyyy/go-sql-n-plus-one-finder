package np1finder

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func (f *Finder) HTTPHandlerNP1FinderMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f.mutex.Lock()
		defer f.mutex.Unlock()

		f.Scan(r.RequestURI)
		next.ServeHTTP(w, r)
		f.Finish()
	})
}

func (f *Finder) EchoNP1FinderMiddleware() echo.MiddlewareFunc {
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
