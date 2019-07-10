package main

import (
	"fmt"
	"os"

	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/jeremy-miller/standup-reporter/internal/asana"
	"github.com/jeremy-miller/standup-reporter/internal/configuration"
)

// set by release process
var (
	version string //nolint:gochecknoglobals
	commit  string //nolint:gochecknoglobals
	date    string //nolint:gochecknoglobals
)

func main() {
	var (
		app        = kingpin.New("standup-reporter", "Command-line application to gather daily standup reports.")
		days       = app.Flag("days", "Number of days to go back to collect completed tasks. Default 1 day (or 3 days on Monday).").Short('d').PlaceHolder("N").Int() //nolint:lll
		asanaToken = app.Flag("asana", "Asana Personal Access Token").Short('a').Required().PlaceHolder("TOKEN").String()
	)
	app.HelpFlag.Short('h')
	app.Version(fmt.Sprintf("Version: %s\nCommit: %s\nBuild Date: %s", version, commit, date))
	kingpin.MustParse(app.Parse(os.Args[1:]))
	fmt.Println("Running standup reporter")
	config := configuration.Get(*days, *asanaToken)
	if err := asana.Report(config); err != nil {
		fmt.Printf("\n%v\n", err)
	}
}
