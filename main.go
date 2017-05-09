package main

import (
	"github.com/yunlzheng/prometheus-pusher/scrape"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/prometheus/retrieval"
	"github.com/prometheus/prometheus/storage"
	"github.com/prometheus/prometheus/config"
	"fmt"
	"flag"
)

var cfg = struct {
	configFile string
}{}

func init() {
	flag.StringVar(
		&cfg.configFile, "config.file", "prometheus.yml",
		"Prometheus configuration file name.",
	)
}

func main() {

	flag.Parse()

	var (
		sampleAppender = storage.Fanout{}
	)

	var (
		targetManager = retrieval.NewTargetManager(sampleAppender)
	)

	var (
		jobTargets = scrape.NewJobTargets(targetManager)
	)

	var (
		scrapeManager = scrape.NewExporterScrape(jobTargets)
	)


	fmt.Println("Loading prometheus config file: " + cfg.configFile)
	conf, err := config.LoadFile(cfg.configFile)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	targetManager.ApplyConfig(conf)

	go targetManager.Run()
	defer targetManager.Stop()

	scrapeManager.AppConfig(conf)

	go scrapeManager.Run()
	defer scrapeManager.Stop()

	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	r.GET("/targets", func(c *gin.Context) {
		c.JSON(200, jobTargets.Targets())
	})
	r.Run()

}

