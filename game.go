// Licensed under GPL, 2016
// Refer to LICENSE for more details
// Refer to README for structural information

package main

import (
	"database/sql"
	"fmt"
	"io"
	"log"
	"net"
	"time"
)

func gameServer() {
	listener, err := net.Listen("tcp", ":25921")
	if err != nil {
		log.Fatal(err)
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal(err)
			continue
		}

		go func(c io.ReadWriteCloser) {
			fmt.Fprintln(c, "@ [RGS] REDB game server")
			defer c.Close()

			var (
				lvl  uint
				task Task
				sol  string
				err  error
			)

			for {
				// read level
				fmt.Fprint(c, ": ")
				_, err = fmt.Fscanf(c, "%d", &lvl)
				if err == io.EOF {
					fmt.Fprintln(c, "\n@ session quit")
					break
				} else if err != nil {
					fmt.Fprintf(c, "! %s\n", err.Error())
					break
				}

				// select random task (according to `getrnd` query)
				if err := getrnd.QueryRow(lvl).Scan(&task.Id); err != nil {
					if err == sql.ErrNoRows {
						fmt.Fprintf(c, "! %s\n", err.Error())
						continue
					}
					fmt.Fprintf(c, "! %s\n", err.Error())
					break
				}

				// generate transaction and load words
				if tx, err := db.Begin(); err != nil {
					fmt.Fprintf(c, "! %s\n", err.Error())
				} else if err := task.loadWords(tx); err != nil {
					fmt.Fprintf(c, "! %s\n", err.Error())
				} else {
					tx.Stmt(inccou).Exec(task.Id)
					tx.Commit()
				}

				// print information
				task.calcLevel()
				fmt.Fprintf(c, "@ id: %x; lvl: %d\n", task.Id, task.Level)

				// print information
				for _, w := range task.Match {
					fmt.Fprintf(c, "+ %s\n", w)
				}
				for _, w := range task.Dmatch {
					fmt.Fprintf(c, "- %s\n", w)
				}

				// read suggestion
				fmt.Fprint(c, "> ")
				_, err = fmt.Fscanf(c, "%s", &sol)
				if err != nil {
					fmt.Fprintf(c, "! %s\n", err.Error())
					break
				}

				// respond adequately (wait if wrong)
				if task.matches(sol) {
					fmt.Fprintln(c, "@ correct")
					task.submit(sol)
				} else {
					fmt.Fprintln(c, "@ invalid")
					time.Sleep(time.Second * time.Duration(len(sol)))
				}

				// clear task object
				task = Task{}
			}
		}(conn)
	}
}
