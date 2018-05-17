package logger

import (
	"os"

	"github.com/sirupsen/logrus"
)

var gnetLogger *logrus.Logger

func init() {
	gnetLogger = logrus.New()

	//TODO logge config
	gnetLogger.SetLevel(logrus.DebugLevel)
	gnetLogger.Out = os.Stdout
}

func Debug(args ...interface{})   { gnetLogger.Debug(args) }
func Warning(args ...interface{}) { gnetLogger.Warning(args) }
func Error(args ...interface{})   { gnetLogger.Error(args) }
func Fatal(args ...interface{})   { gnetLogger.Fatal(args) }
func Info(args ...interface{})    { gnetLogger.Info(args) }

func Debugf(format string, args ...interface{})   { gnetLogger.Debugf(format, args) }
func Warningf(format string, args ...interface{}) { gnetLogger.Warningf(format, args) }
func Errorf(format string, args ...interface{})   { gnetLogger.Errorf(format, args) }
func Fatalf(format string, args ...interface{})   { gnetLogger.Fatalf(format, args) }
func Infof(format string, args ...interface{})    { gnetLogger.Infof(format, args) }

func Debugln(args ...interface{})   { gnetLogger.Debugln(args) }
func Warningln(args ...interface{}) { gnetLogger.Warningln(args) }
func Errorln(args ...interface{})   { gnetLogger.Errorln(args) }
func Fatalln(args ...interface{})   { gnetLogger.Fatalln(args) }
func Infoln(args ...interface{})    { gnetLogger.Infoln(args) }

func WithField(key string, val interface{}) *logrus.Entry { return gnetLogger.WithField(key, val) }
func WithFields(fields logrus.Fields) *logrus.Entry       { return gnetLogger.WithFields(fields) }
