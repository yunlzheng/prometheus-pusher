package main

import (
	"github.com/yunlzheng/prometheus-pusher/scrape"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/prometheus/retrieval"
	"github.com/prometheus/prometheus/storage"
	"github.com/prometheus/prometheus/config"
	"fmt"
	"flag"
	"strings"
)

var cfg = struct {
	configFile        string
	customLabels      string
	customLabelValues string
}{}

var (
	labels, values []string
)

func init() {
	flag.StringVar(
		&cfg.configFile, "config.file", "prometheus_pusher.yml",
		"Prometheus configuration file name.",
	)
	flag.StringVar(
		&cfg.customLabels, "config.customLabels", "", "custom metrics labels",
	)
	flag.StringVar(
		&cfg.customLabelValues, "config.customLabelValues", "", "custom mertics label values",
	)

	if cfg.customLabels=="" {
		labels = []string{}
		values = []string{}
	} else {
		labels = strings.Split(cfg.customLabels, ",")
		values = strings.Split(cfg.customLabelValues, ",")
	}



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
		scrapeManager = scrape.NewExporterScrape(jobTargets, labels, values)
	)

	fmt.Println("Loading prometheus config file: " + cfg.configFile)
	fmt.Println("Custom labels: " + cfg.customLabels + "\t Custom label values: " + cfg.customLabelValues)
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

