package logger

import (
	"github.com/pelletier/go-toml"
	"github.com/sirupsen/logrus"
	"log"
	"os"
)

var L = logrus.New()

func init() {
	conf, err := toml.LoadFile("./conf.toml")
	if err != nil {
		log.Fatalf("Read Config File Fail %e", err)
	}
	level := conf.Get("log.Level").(string)
	res, err := logrus.ParseLevel(level)
	if err != nil {
		log.Fatalf("Parse Config Log Level fail %e", err)
	}
	L.SetLevel(res)
	L.SetOutput(os.Stdout)
}
