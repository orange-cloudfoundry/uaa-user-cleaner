package main

import (
	"fmt"
	"github.com/cloudfoundry-community/gautocloud"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
	"strings"
	"time"
)

func init() {
	if gautocloud.IsInACloudEnv() && gautocloud.CurrentCloudEnv().Name() != "localcloud" {
		log.SetFormatter(&log.JSONFormatter{})
	}
}

func main() {
	panic(boot())
}

func initLogs(c Config) {
	log.SetOutput(os.Stderr)
	if c.Log.JSON {
		log.SetFormatter(&log.JSONFormatter{})
	} else {
		log.SetFormatter(&log.TextFormatter{
			DisableColors: c.Log.NoColor,
		})
	}

	log.SetLevel(log.ErrorLevel)
	switch strings.ToUpper(c.Log.Level) {
	case "ERROR":
		log.SetLevel(log.ErrorLevel)
	case "WARN":
		log.SetLevel(log.WarnLevel)
	case "DEBUG":
		log.SetLevel(log.DebugLevel)
	case "INFO":
		log.SetLevel(log.InfoLevel)
	case "PANIC":
		log.SetLevel(log.PanicLevel)
	case "FATAL":
		log.SetLevel(log.FatalLevel)
	}
}

var gConfig Config

func process() {
	duration, _ := time.ParseDuration(gConfig.Interval)
	for {
		cleaner, err := newCleaner()
		if err == nil {
			cleaner.Run()
			cleaner.Close()
		}
		time.Sleep(duration)
	}
}

func boot() error {
	gautocloud.Inject(&gConfig)
	if err := gConfig.Validate(); err != nil {
		return err
	}

	initLogs(gConfig)
	log.WithFields(log.Fields{
		"name":     os.Args[0],
		"dry_run":  gConfig.DryRun,
		"interval": gConfig.Interval,
	}).Infof("starting")

	r := mux.NewRouter()
	r.Handle("/metrics", promhttp.Handler())

	port := gautocloud.GetAppInfo().Port
	listen := gConfig.Web.Listen
	if port != 0 {
		listen = fmt.Sprintf(":%d", port)
	}

	go process()

	if (gConfig.Web.SSLCert != "") && (gConfig.Web.SSLKey != "") {
		log.WithField("listen", listen).Debugf("serving https")
		return http.ListenAndServeTLS(listen, gConfig.Web.SSLCert, gConfig.Web.SSLKey, r)
	}

	log.WithField("listen", listen).Debugf("serving http")
	return http.ListenAndServe(listen, r)
}
