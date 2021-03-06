// Licensed under GPL, 2016
// Refer to LICENSE for more details
// Refer to README for structural information

package main

import (
	"database/sql"
	"fmt"
	"log"
	"regexp"
	"time"
)

type Solution struct {
	Times       uint
	Solves      *Task
	Solution    string
	First, Last time.Time
}

type Task struct {
	Id          uint
	Called      uint
	Created     time.Time
	Match       []string
	Dmatch      []string
	Solutions   []Solution
	Title       string
	Author      string
	Discription string
}

func (t Task) matches(sol string) bool {
	re, err := regexp.Compile(sol)
	if err != nil {
		return false
	}

	for _, w := range t.Match {
		if re.FindString(w) != w {
			return false
		}
	}

	for _, w := range t.Dmatch {
		if re.FindString(w) == w {
			return false
		}
	}

	return true
}

// Check if task is "acceptable"
// currently only checks if a word in M is in D too
func (t Task) isAcceptable() error {
	for _, w := range t.Match {
		for _, c := range t.Dmatch {
			if w == c && w != "" {
				return fmt.Errorf("Words to match and words not to match have duplicates (%q and %q)", w, c)
			}
		}
	}

	for {
		break
	}

	return nil
}

// writes a solution for a task into the database
func (t *Task) submit(sol string) {
	re := regexp.MustCompile(sol) // to simplify
	_, err := subm.Exec(re.String(), t.Id)
	if err != nil {
		log.Println(err)
	}
}

func (t Task) insert() (int, error) {

	tx, err := db.Begin()
	if err != nil {
		return 0, err
	}

	var id int
	err = create.QueryRow(
		t.Title,
		t.Author,
		t.Discription,
	).Scan(&id)

	if err != nil {
		return 0, err
	} else {
		for _, c := range t.Match {
			if c != "" {
				_, err = tx.Stmt(create_word).Exec(c, 1, id)
				if err != nil {
					log.Println(err)
				}
			}
		}
		for _, c := range t.Dmatch {
			if c != "" {
				_, err = tx.Stmt(create_word).Exec(c, 0, id)
				if err != nil {
					log.Println(err)
				}
			}
		}
		tx.Commit()
		return id, nil
	}
}

func (t Task) GetSolutionCount() (count uint) {
	q, err := getsolc.Query(t.Id)
	if err != nil {
		return
	}
	defer q.Close()

	q.Next()
	if err = q.Scan(&count); err != nil {
		return
	}
	return
}

func (t *Task) loadSolutions() (err error) {
	q, err := getsol.Query(t.Id)
	if err != nil {
		log.Panic(err)
		return
	}
	defer q.Close()

	t.Solutions = nil
	for q.Next() {
		var S Solution
		if err = q.Scan(
			&S.Times,
			&S.Solution,
			&S.First,
			&S.Last,
		); err != nil {
			return
		}
		t.Solutions = append(t.Solutions, S)
	}
	return
}

func (t *Task) loadWords(tx *sql.Tx) (err error) {
	q, err := tx.Stmt(getwrd).Query(t.Id)
	if err != nil {
		log.Panicln(err)
		return
	}
	defer q.Close()

	var word string
	var matches bool
	for q.Next() {
		if err = q.Scan(&word, &matches); err != nil {
			log.Println(err)
			return
		}
		if matches {
			t.Match = append(t.Match, word)
		} else {
			t.Dmatch = append(t.Dmatch, word)
		}
	}
	return
}
