package db

import (
	"database/sql"
	"fmt"
	"strings"
)

type table struct {
	name   string
	schema string
}

func (t table) Schema() string {
	return fmt.Sprintf(t.schema, t.name)
}

var tables = []table{
	{
		"users",
		`
CREATE TABLE IF NOT EXISTS %s (
	id        serial                NOT NULL,
	firstname character varying(255) NOT NULL,
	lastname  character varying(255) NOT NULL,
	CONSTRAINT pk_users  PRIMARY KEY(id)
);
`,
	},
	{
		"pages",
		`
CREATE TABLE IF NOT EXISTS %s (
	slug     character varying(1024) NOT NULL,
	author   integer                 NOT NULL,
	content  text                    NOT NULL,
	CONSTRAINT pk_pages  PRIMARY KEY(slug),
	CONSTRAINT fk_author FOREIGN KEY(author) REFERENCES users(id)
);
`,
	},
}

type User struct {
	ID                  uint32
	Firstname, Lastname string
}

type Page struct {
	Slug    string
	Author  User
	Content string
}

func PurgeSchema(db *sql.DB) error {
	for i := len(tables) - 1; i >= 0; i-- {
		_, err := db.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s;", tables[i].name))
		if err != nil {
			return err
		}
	}

	return nil
}

func CreateSchema(db *sql.DB) error {
	for _, s := range tables {
		err := Tx(db, func(tx *sql.Tx) error {
			_, err := tx.Exec(s.Schema())
			return err
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func CreateDummyData(db *sql.DB) error {
	err := Tx(
		db,
		func(tx *sql.Tx) error {
			return Chain(tx).
				Exec(`INSERT INTO users (firstname, lastname) VALUES('Kobe',     'Lipkens');`).
				Exec(`INSERT INTO users (firstname, lastname) VALUES('Svetlana', 'Pak');`).
				Exec(`INSERT INTO users (firstname, lastname) VALUES('Warre',    'Janssens');`).
				Exec(`INSERT INTO users (firstname, lastname) VALUES('Julienne', 'Hiemeleers');`).
				Err()
		},
	)

	if err != nil {
		return err
	}

	res, err := db.Query("SELECT * FROM users;")
	if err != nil {
		return err
	}
	defer res.Close()

	users := make(map[string]uint32, 4)
	for res.Next() {
		u := User{}
		err = res.Scan(&u.ID, &u.Firstname, &u.Lastname)
		if err != nil {
			return err
		}
		users[strings.ToLower(u.Firstname+u.Lastname)] = u.ID
	}

	return Tx(
		db,
		func(tx *sql.Tx) error {
			return Chain(tx).
				Exec(
					"INSERT INTO pages VALUES('homepage', $1, 'Welcome to this website')",
					users["kobelipkens"],
				).
				Exec(
					"INSERT INTO pages VALUES('ru/homepage', $1, 'Здравствуйте')",
					users["svetlanapak"],
				).
				Exec(
					"INSERT INTO pages VALUES('kobe/about', $1, 'All about me ^^')",
					users["kobelipkens"],
				).
				Exec(
					"INSERT INTO pages VALUES('svetlana/portfolio', $1, 'Welcome to my portfolio!')",
					users["svetlanapak"],
				).
				Err()
		},
	)
}
