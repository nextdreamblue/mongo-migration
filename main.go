package main

import (
	"fmt"
	"os"
	"time"

	"github.com/codegangsta/cli"
	"github.com/gizak/termui"
	"gopkg.in/mgo.v2"
)

func main() {
	app := cli.NewApp()
	app.Name = "mongo-migration"
	app.Usage = "mongo-migration --from <mongo-host-origin> --collection-in <collection-to-migrate> --to <mongo-target> --collection-out <collection-destinity>"
	var collectionIn, collectionOut, fromDb, toDb, fromUrl, toUrl string

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "collection-in",
			Value:       "timeline-events",
			Usage:       "collection to migrate",
			Destination: &collectionIn,
		},
		cli.StringFlag{
			Name:        "collection-out",
			Value:       "timeline-events-out",
			Usage:       "collection destination name to migrate",
			Destination: &collectionOut,
		},
		cli.StringFlag{
			Name:        "from",
			Value:       "localhost",
			Usage:       "mongo url from origin where is the collection to migrate",
			Destination: &fromUrl,
		},
		cli.StringFlag{
			Name:        "from-db",
			Value:       "rdstation_development",
			Usage:       "mongo database from origin where is the collection to migrate",
			Destination: &fromDb,
		},
		cli.StringFlag{
			Name:        "to",
			Value:       "localhost",
			Usage:       "mongo url destination where will be collection to migrate",
			Destination: &toUrl,
		},
		cli.StringFlag{
			Name:        "to-db",
			Value:       "rdstation_development",
			Usage:       "mongo database destination where will be the collection to migrate",
			Destination: &toDb,
		},
	}
	app.Action = func(c *cli.Context) {

		fromSession, err := mgo.Dial(fromUrl)
		toSession, err := mgo.Dial(toUrl)

		if err != nil {
		}

		defer toSession.Close()

		fromSession.SetMode(mgo.Monotonic, true)

		from := InstanceInfo{Session: fromSession, Database: fromDb, CollectionName: collectionIn}
		to := InstanceInfo{Session: toSession, Database: toDb, CollectionName: collectionOut}

		err = termui.Init()
		if err != nil {
			panic(err)
		}
		defer termui.Close()

		strs := []string{
			"[q] [quit](fg-red)",
			"[s] [start](fg-white,bg-green)"}

		ls := termui.NewList()
		ls.Items = strs
		ls.ItemFgColor = termui.ColorYellow
		ls.BorderLabel = "Press key to action"
		ls.Height = 7
		ls.Width = 25
		ls.Y = 0
		termui.Render(ls)
		started := false

		handleMigration := HandleMigration{false, false, false}

		termui.Handle("/sys/kbd/q", func(termui.Event) {
			handleMigration.Stop = true
			for !handleMigration.Stopped {
				fmt.Print(".")
				time.Sleep(1000 * time.Millisecond)
			}
			termui.StopLoop()
		})

		termui.Handle("/sys/kbd/s", func(termui.Event) {
			if !started {
				ImportCollection(&from, &to, &handleMigration)
				started = true
			}
		})
		termui.Handle("/sys/kbd/d", func(termui.Event) {
			handleMigration.LogMode = !handleMigration.LogMode
		})

		termui.Loop()
	}

	app.Run(os.Args)
}
