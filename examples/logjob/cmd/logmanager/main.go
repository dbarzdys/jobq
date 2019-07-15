package main

import (
	"flag"
	"fmt"

	"github.com/dbarzdys/jobq"

	"github.com/dbarzdys/jobq/examples/logjob"
)

func main() {
	var (
		dbPort     = flag.String("db-port", "5432", "DB port")
		dbHost     = flag.String("db-host", "localhost", "DB host")
		dbUser     = flag.String("db-user", "postgres", "DB user")
		dbPassword = flag.String("db-password", "postgres", "DB password")
		dbName     = flag.String("db-name", "postgres", "DB name")
	)
	conninfo := fmt.Sprintf(
		"host=%s port=%s user=%s dbname=%s sslmode=disable password=%s",
		*dbHost,
		*dbPort,
		*dbUser,
		*dbName,
		*dbPassword,
	)
	manager := jobq.NewManager(conninfo)
	if err := manager.Register(
		logjob.Name,
		logjob.New(),
		jobq.WithJobWorkerPoolSize(50),
	); err != nil {
		panic(err)
	}
	if err := manager.Run(); err != nil {
		panic(err)
	}
}
