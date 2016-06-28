package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gizak/termui"
	"gopkg.in/mgo.v2"
)

type InstanceInfo struct {
	Session        *mgo.Session
	Database       string
	CollectionName string
}
type Migration struct {
	Started          time.Time
	RemoteCollection *mgo.Collection
	LocalCollection  *mgo.Collection
	Total            int
	Missing          int
	Throughtput      *termui.LineChart
	Percentage       *termui.Gauge
	InsertRemote     chan []interface{}
	RemoveOrigin     chan interface{}
}

type LogDocs struct {
	node InstanceInfo
}
type HandleMigration struct {
	LogMode bool
	Stop    bool
	Stopped bool
}

func lineChartWithLabel(label string) *termui.LineChart {
	lineChart := termui.NewLineChart()
	lineChart.BorderLabel = label
	lineChart.Data = make([]float64, 101)
	lineChart.Width = 100
	lineChart.Height = 10
	lineChart.X = 0
	lineChart.Y = 0
	lineChart.AxesColor = termui.ColorGreen
	lineChart.LineColor = termui.ColorGreen | termui.AttrBold
	return lineChart
}

func gaugeWithLabel(label string) *termui.Gauge {
	gauge := termui.NewGauge()
	gauge.Percent = 0
	gauge.Width = 50
	gauge.Height = 3
	gauge.X = 55
	gauge.BorderLabel = label
	gauge.LabelAlign = termui.AlignRight
	gauge.BarColor = termui.ColorGreen
	return gauge
}

func ImportCollection(LocalInstance *InstanceInfo, RemoteInstance *InstanceInfo, HandleMigration *HandleMigration) {

	LocalInstance.Session.SetBatch(100000)
	LocalInstance.Session.SetPrefetch(0.1)

	originDB := LocalInstance.Session.DB("") // database name is used from DialInfo
	destinationDB := RemoteInstance.Session.DB("")
	localCollection := originDB.C(LocalInstance.CollectionName)

	totalToImport, _ := localCollection.Count()
	migration := Migration{
		time.Now(),
		destinationDB.C(RemoteInstance.CollectionName),
		localCollection, totalToImport, totalToImport,
		lineChartWithLabel("Imports per minute"),
		gaugeWithLabel("Migration Status"),
		make(chan []interface{}),
		make(chan interface{}),
	}

	f, errFile := os.OpenFile(LocalInstance.CollectionName+".log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if errFile != nil {
		panic("error opening log file")
	}
	defer f.Close()

	log.SetOutput(f)

	log.Println(fmt.Sprintf("\nStarted at %s\n", time.Now().Local()))

	termui.Body.AddRows(
		termui.NewRow(termui.NewCol(6, 0, migration.Percentage),
			termui.NewCol(6, 0, migration.Throughtput)))

	termui.Body.Align()
	termui.Render(termui.Body)

	var i = 0
	var mode = 0

	go migration.keepRemovingOrigin()
	go migration.keepInsertingRemote()

	const BATCH_SIZE = 1000
	var values []interface{} = make([]interface{}, BATCH_SIZE)
	var i_value = 0

	iter := localCollection.Find(nil).Iter()

	for {
		var v map[string]interface{}
		iter.Next(&v)
		if v != nil {
			values[i_value] = v
			i_value++
		}
		mode = i_value % BATCH_SIZE

		if i > 0 && mode == 0 || v == nil || HandleMigration.Stop {
			migration.Percentage.BarColor = termui.ColorRed
			termui.Render(termui.Body)
			if mode != 0 {
				compact := make([]interface{}, mode)
				for j, value := range values {
					if value == nil {
						break
					}
					compact[j] = value
				}
				migration.InsertRemote <- compact
				break
			} else {
				migration.InsertRemote <- values
				if HandleMigration.Stop {
					break
				}
				values = make([]interface{}, BATCH_SIZE)
				i_value = 0
			}
			i++
			migration.Percentage.BarColor = termui.ColorGreen
			termui.Render(termui.Body)
		}

		if i%100 == 0 {
			migration.refreshStatistics(i)

			if i%100000 == 0 {
				log.Println(fmt.Sprintf("%s status %d", time.Now().Local(), i))
			}
			if HandleMigration.LogMode {
				log.Println(fmt.Sprintf("%s status %d - %v", time.Now().Local(), i, v))
			}
		}
		i++
	}

	migration.refreshStatistics(i)
	HandleMigration.Stopped = true

	label := "finished"
	if HandleMigration.Stop {
		label = "stopped"
	}
	info(fmt.Sprintf("\nUpload %s", label))
}
func (migration Migration) totalImported() int {
	return migration.Total - migration.Missing
}
func (migration Migration) upToNow() float64 {
	return float64(time.Since(migration.Started).Seconds() / 60)
}
func (migration Migration) percent() float64 {
	return (float64(migration.totalImported()) / float64(migration.Total)) * 100.0
}
func (migration Migration) rpm() int {
	return int(float64(migration.totalImported()) / migration.upToNow())
}
func (migration Migration) refreshStatistics(i int) {
	missing, _ := migration.LocalCollection.Count()
	migration.Missing = missing
	percent := migration.percent()
	rpm := migration.rpm()

	if percent > 0 && percent <= 100.0 {
		migration.Throughtput.Data[int(percent)] = float64(rpm)
	}

	migration.Throughtput.BorderLabel = fmt.Sprintf("  [ %s ] RPM: %d - Current: %d    ", time.Now().Format(time.Kitchen), rpm, i)
	migration.Percentage.Percent = int(percent)
	status := fmt.Sprintf("Collection: %q Status: %.3f", migration.LocalCollection.Name, percent) + "%   "
	migration.Percentage.BorderLabel = status
	migration.Percentage.Label = fmt.Sprintf("%d from %d. Missing: %d. (%.3f", migration.totalImported(), migration.Total, missing, percent) + "%) "
	migration.Percentage.BarColor = termui.ColorGreen
	termui.Render(termui.Body)
}
func (migration Migration) keepInsertingRemote() {
	values := <-migration.InsertRemote
	if err := migration.RemoteCollection.Insert(values...); err != nil {
		info(fmt.Sprintf("Error inserting batch: %v", err))
		migration.Percentage.BarColor = termui.ColorMagenta
		termui.Render(termui.Body)
		for _, value := range values {
			if value != nil {
				if err = migration.RemoteCollection.Insert(value); err != nil {
					info(fmt.Sprintf("Error: %v - inserting value: %v", err, value))
				} else {
					migration.RemoveOrigin <- value
				}
			}
		}
		migration.Percentage.BarColor = termui.ColorGreen
		termui.Render(termui.Body)
	} else {
		migration.Percentage.BarColor = termui.ColorCyan
		termui.Render(termui.Body)
		for _, value := range values {
			if value != nil {
				migration.RemoveOrigin <- value
			}
		}
		migration.Percentage.BarColor = termui.ColorGreen
		termui.Render(termui.Body)
	}
	migration.keepInsertingRemote()
}

func info(message string) {
	log.Println(fmt.Sprintf("%s: %s", time.Now().Local(), message))
}

func (migration Migration) keepRemovingOrigin() {
	value := <-migration.RemoveOrigin
	id := value.(map[string]interface{})["_id"]
	if id != nil {
		err := migration.LocalCollection.RemoveId(id)
		if err != nil {
			log.Println(fmt.Sprintf("Error removing from origin: %v - value: %v", err, value))
		}
	}
	migration.keepRemovingOrigin()
}
