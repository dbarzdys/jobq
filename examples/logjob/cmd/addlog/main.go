package main

import (
	"database/sql"
	"flag"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/dbarzdys/jobq"

	"github.com/dbarzdys/jobq/examples/logjob"
)

func main() {
	fs := flag.NewFlagSet("addlog", flag.ExitOnError)
	var (
		dbPort     = fs.String("db-port", "5432", "postgresql port")
		dbHost     = fs.String("db-host", "localhost", "postgresql host")
		dbUser     = fs.String("db-user", "postgres", "postgresql user")
		dbPassword = fs.String("db-password", "postgres", "postgresql password")
		dbName     = fs.String("db-name", "postgres", "postgresql name")
		n          = fs.Uint("n", 1, "number of tasks to queue")
	)
	fs.Usage = usageFor(fs, os.Args[0]+" [flags] message")
	fs.Parse(os.Args[1:])
	if len(fs.Args()) == 0 {
		fs.Usage()
		os.Exit(1)
	}
	conninfo := fmt.Sprintf(
		"host=%s port=%s user=%s dbname=%s sslmode=disable password=%s",
		*dbHost,
		*dbPort,
		*dbUser,
		*dbName,
		*dbPassword,
	)
	// connect to postgres
	db, err := sql.Open("postgres", conninfo)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	// add tasks
	for i := uint(0); i < *n; i++ {
		task := jobq.NewTask(logjob.Name, &logjob.TaskBody{
			Message: strings.Join(fs.Args(), " "),
		})
		err = task.Queue(db)
		if err != nil {
			panic(err)
		}
	}
	fmt.Printf("%d tasks queued\n", *n)
}

func usageFor(fs *flag.FlagSet, short string) func() {
	return func() {
		fmt.Fprintf(os.Stderr, "USAGE\n")
		fmt.Fprintf(os.Stderr, "  %s\n", short)
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "FLAGS\n")
		w := tabwriter.NewWriter(os.Stderr, 0, 2, 2, ' ', 0)
		fs.VisitAll(func(f *flag.Flag) {
			fmt.Fprintf(w, "\t-%s %s\t%s\n", f.Name, f.DefValue, f.Usage)
		})
		w.Flush()
		fmt.Fprintf(os.Stderr, "\n")
	}
}
