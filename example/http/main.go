package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/rrreeeyyy/go-sql-n-plus-one-finder/pkg/np1finder"
	proxy "github.com/shogo82148/go-sql-proxy"
)

type UserModel struct {
	ID    int         `db:"id"`
	Name  string      `db:"name"`
	Posts []PostModel `json:"posts"`
}

type PostModel struct {
	ID          int            `db:"id" json:"id"`
	UserID      int            `db:"user_id" json:"user_id"`
	Description string         `db:"description" json:"description"`
	Comments    []CommentModel `json:"comments"`
}

type CommentModel struct {
	ID          int    `db:"id" json:"id"`
	PostID      int    `db:"post_id" json:"post_id"`
	Description string `db:"description" json:"description"`
}

func main() {
	ctx := context.Background()

	dsn := "root@tcp(localhost:3306)/np1finder"

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{}))

	finder := np1finder.NewFinder(np1finder.Config{Context: ctx, Logger: logger, Threshold: 2})
	sql.Register("mysql:np1finder", proxy.NewProxyContext(&mysql.MySQLDriver{}, finder.NewHooksContext()))
	db, err := sqlx.Open("mysql:np1finder", dsn)

	// db, err := sqlx.Open("mysql", dsn)

	if err != nil {
		panic(err)
	}

	http.Handle("/", finder.HTTPHandlerMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := []UserModel{}

		users := []UserModel{}
		err := db.SelectContext(ctx, &users, "SELECT * FROM users")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		resp = append(resp, users...)

		for i, user := range users {
			posts := []PostModel{}
			err := db.SelectContext(ctx, &posts, "SELECT * FROM posts WHERE user_id = ?", user.ID)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			resp[i].Posts = append(resp[i].Posts, posts...)

			for j, post := range posts {
				comments := []CommentModel{}
				err := db.SelectContext(ctx, &comments, "SELECT * FROM comments WHERE post_id = ?", post.ID)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
				}
				resp[i].Posts[j].Comments = append(resp[i].Posts[j].Comments, comments...)
			}
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			log.Fatal(err)
		}
	})))

	log.Fatal(http.ListenAndServe(":8080", nil))
}
