/*
@version: 1.0
@author: allanpan
@license:  Apache Licence
@contact: panjf2000@gmail.com
@site:
@file: init.go
@time: 2018/2/11 16:40
@tag: 1,2,3
@todo: ...
*/
package config

import (
	"fmt"
	"log"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

//RuntimeViper runtime config
var RuntimeViper *viper.Viper

func init() {
	RuntimeViper = viper.New()
	RuntimeViper.SetConfigType("toml")
	RuntimeViper.SetConfigName("cfg")         // name of config file (without extension)
	RuntimeViper.AddConfigPath("/etc/proxy/") // path to look for the config file in
	RuntimeViper.AddConfigPath("./config/")   // optionally look for config in the working directory
	if err := RuntimeViper.ReadInConfig(); err != nil {
		panic(fmt.Errorf("fatal error config file: %s", err))
	}
	RuntimeViper.WatchConfig()
	RuntimeViper.OnConfigChange(func(e fsnotify.Event) {
		log.Printf("config file changed:%s", e.Name)
	})
}
