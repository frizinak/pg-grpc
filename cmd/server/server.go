package main

import (
	"database/sql"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/frizinak/pg-grpc/db"
	"github.com/frizinak/pg-grpc/pb"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"
)

type app struct {
	db *sql.DB
	pb.UnimplementedAppServer
}

func (app *app) Pages(r *pb.PagesRequest, res pb.App_PagesServer) error {
	rows, err := app.db.Query(`
SELECT p.slug, u.firstname, u.lastname, p.content
FROM pages p
JOIN users u ON p.author = u.id;
`,
	)
	if err != nil {
		return err
	}

	defer rows.Close()
	for rows.Next() {
		r := &pb.Page{Author: &pb.Author{}}
		ex(rows.Scan(
			&r.Slug,
			&r.Author.Firstname,
			&r.Author.Lastname,
			&r.Content,
		))

		if err := res.Context().Err(); err != nil {
			return err
		}

		if err := res.Send(r); err != nil {
			return err
		}
	}

	return nil
}

func ex(err error) {
	if err == nil {
		return
	}
	fmt.Fprintln(os.Stderr, "-- SERVER:", err)
	os.Exit(1)
}

func main() {
	debug := true
	host := os.Args[1]
	ca := os.Args[2]

	pg, err := sql.Open(
		"postgres",
		fmt.Sprintf("user=db_user password=db_pass host=%s dbname=app sslmode=verify-full sslrootcert=%s", host, ca),
	)
	ex(err)

	if debug {
		ex(db.PurgeSchema(pg))
	}

	ex(db.CreateSchema(pg))

	if debug {
		ex(db.CreateDummyData(pg))
	}

	app := &app{db: pg}

	tcp, err := net.Listen("tcp", ":8080")
	ex(err)
	grpc := grpc.NewServer(grpc.ConnectionTimeout(time.Second * 3))
	pb.RegisterAppServer(grpc, app)
	ex(grpc.Serve(tcp))
}
