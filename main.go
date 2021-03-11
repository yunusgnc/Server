package main

import (
	"flag"

	"github.com/sirupsen/logrus"
)

//Server code borrowed from: https://dev.to/koddr/let-s-write-config-for-your-golang-web-app-on-right-way-yaml-5ggp

var configPath = flag.String("config", "./config.yaml", "path to config file")

const AppName = "cerci-platform"

func main() {
	flag.Parse()
	logrus.SetFormatter(&logrus.JSONFormatter{})

	var conf *Config
	var err error
	if *configPath != "" {
		conf, err = NewConfig(*configPath)
		if err != nil {
			logrus.WithError(err).Fatal("Failed to read configuration file")
		}
	}
	logrus.WithField("configFile", *configPath).Info("Read config file")
	conf.AppName = AppName
	app := NewApp(conf)
	err = app.Run()
	if err != nil {
		logrus.WithError(err).Fatal("Failed to start application")
	}

}
