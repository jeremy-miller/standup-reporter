package main

import (
	"os"

	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/jeremy-miller/standup-reporter/internal/asana"
	"github.com/jeremy-miller/standup-reporter/internal/configuration"
)

func main() {
	var (
		app        = kingpin.New("standup-reporter", "Command-line application to gather daily standup reports.")
		days       = app.Flag("days", "Number of days to go back to collect completed tasks. Default 1 day (or 3 days on Monday).").Short('d').PlaceHolder("N").Int()
		asanaToken = app.Flag("asana", "Asana Personal Access Token").Short('a').Required().PlaceHolder("TOKEN").String()
	)
	app.HelpFlag.Short('h')
	kingpin.MustParse(app.Parse(os.Args[1:]))
	config := configuration.Get(*days, *asanaToken)
	asana.Report(config)
}
