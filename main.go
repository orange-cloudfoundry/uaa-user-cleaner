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
	"github.com/prometheus/common/version"
	"gopkg.in/alecthomas/kingpin.v2"
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

func boot() error {
	kingpin.Version(version.Print("uaa-user-cleaner"))
	kingpin.HelpFlag.Short('h')
	usage := strings.ReplaceAll(kingpin.DefaultUsageTemplate, "usage: ", "usage: CLOUD_FILE=config.yml ")
	kingpin.UsageTemplate(usage)
	kingpin.Parse()

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

	cleaner := newCleaner()
	r := mux.NewRouter()

	r.Handle("/metrics", promhttp.Handler())
	r.HandleFunc("/v1/listInvalidUsers", cleaner.ListInvalidUsers)

	port := gautocloud.GetAppInfo().Port
	listen := gConfig.Web.Listen
	if port != 0 {
		listen = fmt.Sprintf(":%d", port)
	}

	go cleaner.Run(gConfig.Interval)

	if (gConfig.Web.SSLCert != "") && (gConfig.Web.SSLKey != "") {
		log.WithField("listen", listen).Infof("serving https")
		return http.ListenAndServeTLS(listen, gConfig.Web.SSLCert, gConfig.Web.SSLKey, r)
	}

	log.WithField("listen", listen).Infof("serving http")
	return http.ListenAndServe(listen, r)
}
