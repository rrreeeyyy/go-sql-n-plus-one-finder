package main

import (
	"context"
	"database/sql"
	"log/slog"
	"os"

	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	echo "github.com/labstack/echo/v4"
	"github.com/rrreeeyyy/go-sql-n-plus-one-finder/pkg/np1finder"
	proxy "github.com/shogo82148/go-sql-proxy"
)

type UserModel struct {
	ID    int    `db:"id"`
	Name  string `db:"name"`
	Posts []PostModel
}

type PostModel struct {
	ID          int    `db:"id"`
	UserID      int    `db:"user_id"`
	Description string `db:"description"`
	Comments    []CommentModel
}

type CommentModel struct {
	ID          int    `db:"id"`
	PostID      int    `db:"post_id"`
	Description string `db:"description"`
}

func main() {
	ctx := context.Background()

	dsn := "root@tcp(localhost:3306)/np1finder"

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{}))

	finder := np1finder.NewFinder(np1finder.Config{Context: ctx, Logger: logger, Threshold: 2})
	sql.Register("mysql:np1finder", proxy.NewProxyContext(&mysql.MySQLDriver{}, finder.NewNPlusOneFinderHooksContext(ctx)))
	db, err := sqlx.Open("mysql:np1finder", dsn)

	// db, err := sqlx.Open("mysql", dsn)

	if err != nil {
		panic(err)
	}

	e := echo.New()

	e.Use(finder.EchoNP1FinderMiddleware())

	e.GET("/", func(c echo.Context) error {
		resp := []UserModel{}

		users := []UserModel{}
		err := db.SelectContext(ctx, &users, "SELECT * FROM users")
		if err != nil {
			return err
		}
		resp = append(resp, users...)

		for i, user := range users {
			posts := []PostModel{}
			err := db.SelectContext(ctx, &posts, "SELECT * FROM posts WHERE user_id = ?", user.ID)
			if err != nil {
				return err
			}
			resp[i].Posts = append(resp[i].Posts, posts...)

			for j, post := range posts {
				comments := []CommentModel{}
				err := db.SelectContext(ctx, &comments, "SELECT * FROM comments WHERE post_id = ?", post.ID)
				if err != nil {
					return err
				}
				resp[i].Posts[j].Comments = append(resp[i].Posts[j].Comments, comments...)
			}
		}

		return c.JSON(200, resp)
	})
	e.Logger.Fatal(e.Start(":1323"))
}
