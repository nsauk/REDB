// Licensed under GPL, 2016
// Refer to LICENSE for more details
// Refer to README for structural information

package main

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"os"
)

const entries = 14 // entries per page, should be even

var (
	db                         *sql.DB
	create, getsol, getone     *sql.Stmt
	subm, getsolc, create_word *sql.Stmt
	findn, findp, findd        *sql.Stmt
	findnr, findpr, finddr     *sql.Stmt
	getan, getap, getad        *sql.Stmt
	getrnd, getwrd, inccou     *sql.Stmt
)

func init() {
        log.SetFlags(log.LstdFlags | log.Lshortfile)
	if os.Getenv("DBCONN") == "" {
		log.Fatal("\"$DBCONN\" not defined.")
	}

	var err error
	db, err = sql.Open("sqlite3", os.Getenv("DBCONN"))
	if err != nil {
		log.Fatal(err)
	}

	mustExec := func(stmt string) {
		_, err := db.Exec(stmt)
		if err != nil {
			log.Fatal(err)
		}
	}

	mustPrepare := func(stmt string) *sql.Stmt {
		r, err := db.Prepare(stmt)
		if err != nil {
			log.Fatal(err)
		}
		return r
	}

	genGetall := func(sort string) *sql.Stmt {
		c := `SELECT id, title, called, created
                        FROM tasks ORDER BY %s LIMIT %d OFFSET $1`
		stmt, err := db.Prepare(fmt.Sprintf(c, sort, entries))
		if err != nil {
			log.Fatal(err)
		}
		return stmt
	}

	genFind := func(sort string) *sql.Stmt {
		c := `SELECT id, title, called, created
                        FROM tasks WHERE title LIKE '%%$1%%'
                    ORDER BY %s LIMIT %d OFFSET $2`
		stmt, err := db.Prepare(fmt.Sprintf(c, sort, entries))
		if err != nil {
			log.Fatal(err)
		}
		return stmt
	}

	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}

	mustExec(`CREATE TABLE IF NOT EXISTS tasks (
                         id       INTEGER PRIMARY KEY AUTOINCREMENT,
                         title    TEXT NOT NULL,
                         author   TEXT,
                         discrip  TEXT,
                         called   INTEGER DEFAULT 0,
                         created  TIMESTAMP DEFAULT CURRENT_TIMESTAMP)`)

	mustExec(`CREATE TABLE IF NOT EXISTS solutions (
		         suggested   INTEGER DEFAULT 1,
                         solves      INTEGER NOT NULL REFERENCES tasks (id) ON DELETE CASCADE,
                         solution    TEXT NOT NULL,
                         first       TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                         last        TIMESTAMP DEFAULT CURRENT_TIMESTAMP)`)
//,                         PRIMARY KEY (solution, solves))`)

	mustExec(`CREATE TABLE IF NOT EXISTS words (
                         word    VARCHAR(75) NOT NULL,
                         matches BOOL NOT NULL,
                         task    INTEGER NOT NULL REFERENCES tasks(id) ON DELETE CASCADE)`)

	getad = genGetall("called")
	getan = genGetall("created DESC")
	getap = genGetall("(SELECT COUNT(1) FROM solutions WHERE solves = id) DESC")

	findd = genFind("called")
	findn = genFind("created DESC")
	findp = genFind("(SELECT COUNT(1) FROM solutions WHERE solves = id) DESC")

	finddr = genFind("called DESC")
	findnr = genFind("created")
	findpr = genFind("(SELECT COUNT(1) FROM solutions WHERE solves = id)")

	create = mustPrepare(`INSERT INTO tasks (title, author, discrip)
                                   VALUES ($1, $2, $3);
                                   SELECT last_insert_rowid()`)

	create_word = mustPrepare(`INSERT INTO words (word, matches, task)
                                        VALUES ($1, $2, $3)`)

	subm = mustPrepare(`INSERT INTO solutions (solution, solves)
                                 VALUES ($1, $2)`)
//                                  WHERE true
//                            ON CONFLICT (solution, solves)
//                          DO UPDATE SET suggested = solutions.suggested + 1,
//                                             last = CURRENT_TIMESTAMP`)

	getsolc = mustPrepare(`SELECT COUNT(1)
	                       FROM solutions
                               WHERE solves = $1`)

	getsol = mustPrepare(`SELECT suggested, solution, first, last
                                FROM solutions
                               WHERE solves = $1
                            ORDER BY -suggested`)

	getwrd = mustPrepare(`SELECT word, matches
                                FROM words
                               WHERE task = $1`)

	getone = mustPrepare(`SELECT title, author, discrip, called,created
                                FROM tasks
                               WHERE id = $1`)

	getrnd = mustPrepare(`SELECT id
                                FROM tasks
                            ORDER BY called / strftime('%s', created) DESC
                               LIMIT 1
                              OFFSET RANDOM() * (SELECT COUNT(1) - 1 FROM tasks)`)

	inccou = mustPrepare(`UPDATE tasks
                                 SET called = called + 1
                               WHERE id = $1`)
}
