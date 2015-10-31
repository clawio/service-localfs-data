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
	sharedSecretEnvar = "CLAWIO_SHAREDSECRET"

	endPoint = "/"
)

type environ struct {
	dataDir      string
	tmpDir       string
	port         int
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
	return e, nil
}

func printEnviron(e *environ) {
	log.Printf("%s=%s", dataDirEnvar, e.dataDir)
	log.Printf("%s=%s", tmpDirEnvar, e.tmpDir)
	log.Printf("%s=%d", portEnvar, e.port)
	log.Printf("%s=%s", sharedSecretEnvar, "******")
}

func main() {

	c := xhandler.Chain{}

	// Add close notifier handler so context is cancelled when the client closes
	// the connection
	c.UseC(xhandler.CloseHandler)

	// Install the logger handler with default output on the console
	lh := xlog.NewHandler(xlog.LevelDebug)

	// Set some global env fields
	host, _ := os.Hostname()
	lh.SetFields(xlog.F{
		"svc":  serviceID,
		"host": host,
	})

	c.UseC(lh.HandlerC)

	// Plug the xlog handler's input to Go's default logger
	log.SetOutput(lh.NewLogger())

	c.UseC(lh.HandlerC)

	c.UseC(xaccess.NewHandler())

	log.Printf("Service %s started", serviceID)

	env, err := getEnviron()
	printEnviron(env)

	if err != nil {
		log.Fatal(err)
	}

	p := &newServerParams{}
	p.dataDir = env.dataDir
	p.tmpDir = env.tmpDir
	p.sharedSecret = env.sharedSecret

	srv := newServer(p)

	http.Handle(endPoint, c.Handler(srv))

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", env.port), nil))
}
