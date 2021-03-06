package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"

	"github.com/rustyeddy/moni"
)

var (
	config moni.Configuration
)

func init() {

	// Logerr settings
	flag.StringVar(&config.Logfile, "logfile", "stdout", "Were to send log output")
	flag.StringVar(&config.LogLevel, "level", "warn", "Log level to set")
	flag.StringVar(&config.FormatString, "format", "json", "Format to print log files")

	// Crank up the verbosity a bit
	flag.BoolVar(&config.Debug, "debug", false, "Debug stuff with this program")

	// Address and Port the server to open and listen for connections
	flag.StringVar(&config.Addrport, "addr", ":8888", " an Daemon in the background")

	flag.StringVar(&config.ConfigFile, "cfg", "/srv/moni/config.json", "Use configuration file")

	// -crawl related
	flag.IntVar(&config.Depth, "crawl-depth", 1, "Max crawl depth")
	flag.StringVar(&config.Pubdir, "appdir", "pub", "Serve the site from this dir")
	flag.StringVar(&config.Storedir, "storedir", "/srv/moni/", "Directory for Store to use")
	flag.StringVar(&config.Tmpldir, "tmpldir", "../tmpl", "Basedir for templates")

	flag.BoolVar(&config.Profile, "prof", true, "Profile our http server (daemon)")
}

func main() {
	// Flags are mostly set in the moni.config.go package
	flag.Parse()

	// Gets the app, and saves the configuration file with the app
	cli := moni.NewClient(&config)
	cli.Init()
	cli.Start()

	// Wait for the server (and/or) client to end
	// ====================================================================
	// We'll accept graceful shutdowns when quit via SIGINT (Ctrl+C)
	// SIGKILL, SIGQUIT or SIGTERM (Ctrl+/) will not be caught.
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	// Block until we receive our signal.
	<-c

	// Create a deadline to wait for.
	ctx, cancel := context.WithTimeout(context.Background(), config.Wait)
	defer cancel()
	// Doesn't block if no connections, but will otherwise wait
	// until the timeout deadline.
	app.Shutdown(ctx)

	// Optionally, you could run srv.Shutdown in a goroutine and block on
	// <-ctx.Done() if your application should wait for other services
	// to finalize based on context cancellation.
	log.Println("shutting down")

	os.Exit(0)

}

func StartClient(app *moni.App) {

	// ========================================================================
	// Process commands from the command line
	nargs := len(flag.Args())
	if nargs > 0 {
		var resp *http.Response
		cmd := flag.Arg(0)

		// What address and port to connect to and something to
		// write the output to (io.Writer)
		cli := moni.NewClient(os.Stdout, config.Addrport)
		arg := flag.Arg(1)

		// Run a single command in the foreground
		switch cmd {
		case "GET", "POST", "PUT", "DELETE", "get", "post", "put", "delete":
			resp = cli.Do(cmd, arg)

			body := moni.GetBody(resp)
			fmt.Printf("Body %+v", body)
			/*
				case "crawl":
					cli.CrawlUrl(arg)

				case "crawlids":
					cli.CrawlList()

				case "view":
					cli.CrawlId(arg)
			*/
		}
	}
}
