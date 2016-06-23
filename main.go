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
	var collectionIn, collectionOut, fromUrl, toUrl string

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "in",
			Value:       "input",
			Usage:       "collection input to migrate",
			Destination: &collectionIn,
		},
		cli.StringFlag{
			Name:        "out",
			Value:       "output",
			Usage:       "collection output name to be migrated",
			Destination: &collectionOut,
		},
		cli.StringFlag{
			Name:        "from",
			Value:       "mongodb://localhost:27017/example",
			Usage:       "mongo url from origin where is the collection to migrate",
			Destination: &fromUrl,
		},
		cli.StringFlag{
			Name:        "to",
			Value:       "mongodb://localhost:27017/example2",
			Usage:       "mongo url destination where will be collection to migrate",
			Destination: &toUrl,
		},
	}
	app.Action = func(c *cli.Context) {

		fmt.Println("get session from url: ", fromUrl)
		fromSession, err := getSession(fromUrl)
		if err != nil {
			fmt.Println("ops!: ", err)
			panic(err)
		}
		fmt.Println("get session to url: ", toUrl)
		toSession, err := getSession(toUrl)

		if err != nil {
			fmt.Println("ops!: ", err)
			panic(err)
		}
		fmt.Println("o")

		defer fromSession.Close()

		toSession.SetMode(mgo.Monotonic, true)
		fromSession.SetMode(mgo.Monotonic, true)

		from := InstanceInfo{Session: fromSession, CollectionName: collectionIn}
		to := InstanceInfo{Session: toSession, CollectionName: collectionOut}

		err = termui.Init()
		if err != nil {
			panic(err)
		}
		defer termui.Close()

		termui.Render(keyboardShortcuts())
		started := false

		handleMigration := HandleMigration{false, false, false}

		setupKeyboardHandle(handleMigration, started, from, to)

		termui.Loop()
	}

	app.Run(os.Args)
}
func getSession(uri string) (*mgo.Session, error) {

	dialInfo, err := mgo.ParseURL(uri)

	if err != nil {
		return nil, err
	}

	dialInfo.Timeout = 300 * time.Millisecond

	session, err := mgo.DialWithInfo(dialInfo)
	if err != nil {
		return nil, err
	}

	return session, err
}
