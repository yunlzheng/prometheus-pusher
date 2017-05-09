package main

import (
	pusher "github.com/yunlzheng/prometheus-pusher/model"
	"github.com/prometheus/common/model"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/prometheus/retrieval"
	"github.com/prometheus/prometheus/storage"
	"github.com/prometheus/prometheus/config"
	"fmt"
	"sort"
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

	fmt.Println("Loading prometheus config file: " + cfg.configFile)
	conf, err := config.LoadFile(cfg.configFile)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	targetManager.ApplyConfig(conf)

	go targetManager.Run()
	defer targetManager.Stop()

	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	r.GET("/targets", func(c *gin.Context) {

		tps := map[string][]*retrieval.Target{}
		type AA struct {
			Name string
		}
		for _, t := range targetManager.Targets() {
			job := string(t.Labels()[model.JobLabel])
			tps[job] = append(tps[job], t)
		}

		for _, targets := range tps {
			sort.Slice(targets, func(i, j int) bool {
				return targets[i].Labels()[model.InstanceLabel] < targets[j].Labels()[model.InstanceLabel]
			})
		}

		for job, pool := range tps {
			for _, endpoint := range pool {
				fmt.Print(fmt.Sprintf("job: %s endpoint: %s health: %s LastScrape: %s\n", job, endpoint.URL(), endpoint.Health(), endpoint.LastScrape()))
			}
		}

		c.JSON(200, pusher.ToTargets(tps))

	})
	r.Run()

}

