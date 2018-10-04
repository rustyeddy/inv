package main

import (
	"flag"
)

type ConfigLogger struct {
	Output string
	Level  string
	Format string
}

type Configuration struct {
	Log         ConfigLogger
	Addrport    string
	Pubdir      string // Where to serve the static files from
	Depth       int
	HttpAddr    string
	StaticAddr  string
	StartHttp   bool
	StartStatic bool
	Client      bool
}

var (
	Config Configuration
)

func init() {
	flag.StringVar(&Config.Log.Output, "output", "stdout", "Were to send log output")
	flag.StringVar(&Config.Log.Level, "level", "warn", "Log level to set")
	flag.StringVar(&Config.Log.Format, "format", "json", "Format to print log files")

	// use flags
	flag.StringVar(&Config.HttpAddr, "http-addr", ":4444", " an Daemon in the background")
	flag.StringVar(&Config.StaticAddr, "static-addr", ":5555", "Run an Daemon in the background")

	flag.BoolVar(&Config.StartStatic, "static", false, "Start the static server ")
	flag.BoolVar(&Config.StartHttp, "http", false, "Start the static server ")

	flag.IntVar(&Config.Depth, "depth", 1, "Max crawl depth")
	flag.BoolVar(&Config.Client, "cli", false, "Run a command line client")
	flag.StringVar(&Config.Pubdir, "dir", "./pub", "Run an Daemon in the background")
}

func GetConfiguration() *Configuration {
	return &Config
}
