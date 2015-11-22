package main

import (
	"fmt"
	"github.com/rs/xhandler"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
	"strconv"
)

const (
	serviceID         = "CLAWIO_LOCALSTOREDATA"
	dataDirEnvar      = serviceID + "_DATADIR"
	tmpDirEnvar       = serviceID + "_TMPDIR"
	checksumEnvar     = serviceID + "_CHECKSUM"
	portEnvar         = serviceID + "_PORT"
	propEnvar         = serviceID + "_PROP"
	sharedSecretEnvar = "CLAWIO_SHAREDSECRET"

	endPoint = "/"
)

type environ struct {
	dataDir      string
	tmpDir       string
	checksum     string
	port         int
	prop         string
	sharedSecret string
}

func getEnviron() (*environ, error) {
	e := &environ{}
	e.dataDir = os.Getenv(dataDirEnvar)
	e.tmpDir = os.Getenv(tmpDirEnvar)
	e.checksum = os.Getenv(checksumEnvar)
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
	log.Infof("%s=%s\n", dataDirEnvar, e.dataDir)
	log.Infof("%s=%s\n", tmpDirEnvar, e.tmpDir)
	log.Infof("%s=%d\n", portEnvar, e.port)
	log.Infof("%s=%s\n", propEnvar, e.prop)
	log.Infof("%s=%s\n", sharedSecretEnvar, "******")
}

func main() {

	c := xhandler.Chain{}
	c.UseC(xhandler.CloseHandler)

	log.Info("Service %s started", serviceID)

	env, err := getEnviron()
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}

	printEnviron(env)

	p := &newServerParams{}
	p.dataDir = env.dataDir
	p.tmpDir = env.tmpDir
	p.checksum = env.checksum
	p.prop = env.prop
	p.sharedSecret = env.sharedSecret

	srv, err := newServer(p)
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}

	http.Handle(endPoint, c.Handler(srv))
	log.Error(http.ListenAndServe(fmt.Sprintf(":%d", env.port), nil))
}
