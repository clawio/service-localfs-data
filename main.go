package main

import (
	"fmt"
	"github.com/rs/xaccess"
	"github.com/rs/xhandler"
	"github.com/rs/xlog"
	"log"
	"net/http"
	"os"
	"strconv"
)

const (
	serviceID         = "CLAWIO_LOCALSTOREDATA"
	dataDirEnvar      = serviceID + "_DATADIR"
	tmpDirEnvar       = serviceID + "_TMPDIR"
	portEnvar         = serviceID + "_PORT"
	propEnvar         = serviceID + "_PROP"
	sharedSecretEnvar = "CLAWIO_SHAREDSECRET"

	endPoint = "/"
)

type environ struct {
	dataDir      string
	tmpDir       string
	port         int
	prop         string
	sharedSecret string
}

func getEnviron() (*environ, error) {
	e := &environ{}
	e.dataDir = os.Getenv(dataDirEnvar)
	e.tmpDir = os.Getenv(tmpDirEnvar)
	port, err := strconv.Atoi(os.Getenv(portEnvar))
	if err != nil {
		return nil, err
	}
	e.port = port
	e.sharedSecret = os.Getenv(sharedSecretEnvar)
	e.prop = os.Getenv(propEnvar)
	return e, nil
}

func printEnviron(e *environ) {
	log.Printf("%s=%s", dataDirEnvar, e.dataDir)
	log.Printf("%s=%s", tmpDirEnvar, e.tmpDir)
	log.Printf("%s=%d", portEnvar, e.port)
	log.Printf("%s=%s", propEnvar, e.prop)
	log.Printf("%s=%s", sharedSecretEnvar, "******")
}

func setUpLog() {

}

func main() {

	host, _ := os.Hostname()
	conf := xlog.Config{
		// Log info level and higher
		Level: xlog.LevelDebug,
		// Set some global env fields
		Fields: xlog.F{
			"svc":  serviceID,
			"host": host,
		},
		// Output everything on console
		Output: xlog.NewOutputChannel(xlog.NewConsoleOutput()),
	}

	xl := xlog.New(conf)

	c := xhandler.Chain{}
	c.UseC(xlog.NewHandler(conf))
	c.UseC(xhandler.CloseHandler)
	c.UseC(xaccess.NewHandler())

	// Plug the xlog handler's input to Go's default logger
	log.SetOutput(xl)

	log.Printf("Service %s started", serviceID)

	env, err := getEnviron()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	printEnviron(env)

	p := &newServerParams{}
	p.dataDir = env.dataDir
	p.tmpDir = env.tmpDir
	p.prop = env.prop
	p.sharedSecret = env.sharedSecret

	srv := newServer(p)

	http.Handle(endPoint, c.Handler(srv))

	fmt.Fprintln(os.Stderr, http.ListenAndServe(fmt.Sprintf(":%d", env.port), nil))
}
