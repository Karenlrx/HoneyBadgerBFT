package hbbft

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/axiomesh/axiom-kit/log"
	"github.com/sirupsen/logrus"
)

const (
	honeyBadger = "honey_badger"
	bba         = "bba"
	rbc         = "rbc"
	acs         = "acs"
)

type LoggerWrapper struct {
	loggers map[string]*logrus.Entry
}

var w = &LoggerWrapper{
	loggers: map[string]*logrus.Entry{
		honeyBadger: log.NewWithModule(honeyBadger),
		bba:         log.NewWithModule(bba),
		rbc:         log.NewWithModule(rbc),
		acs:         log.NewWithModule(acs),
	},
}

func Initialize() error {
	path, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("Failed to get current working directory: %v\n", err)
	}

	if err = log.Initialize(
		log.WithCtx(context.Background()),
		log.WithFileName("hb"),
		log.WithFilePath(path),
		log.WithEnableColor(true),
		log.WithMaxAge(7),
		log.WithMaxSize(128),
		log.WithRotationTime(24*time.Hour),
		log.WithPersist(true),
		log.WithDisableTimestamp(false),
	); err != nil {
		return err
	}
	m := make(map[string]*logrus.Entry)
	m[honeyBadger] = log.NewWithModule(honeyBadger)
	m[honeyBadger].Logger.SetLevel(logrus.InfoLevel)
	m[bba] = log.NewWithModule(bba)
	m[bba].Logger.SetLevel(logrus.InfoLevel)
	m[rbc] = log.NewWithModule(rbc)
	m[rbc].Logger.SetLevel(logrus.InfoLevel)
	m[acs] = log.NewWithModule(acs)
	m[acs].Logger.SetLevel(logrus.InfoLevel)
	w.loggers = m
	return nil
}

func Logger(name string) logrus.FieldLogger {
	return w.loggers[name]
}
