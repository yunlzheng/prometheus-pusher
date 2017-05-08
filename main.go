package main

import (
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

type JobTargets struct {
	Targets []*JobTarget
}

type JobTarget struct {
	Name      string
	Endpoints []*JobEndpoint
}

type JobEndpoint struct {
	Endpoint string
	Health   string
}

func ToTargets(tps map[string][]*retrieval.Target) []*JobTarget {
	targets := []*JobTarget{}
	for job, pool := range tps {
		targets = append(targets, &JobTarget{
			Name: job,
			Endpoints: covertToEndpoints(pool),
		})
	}
	return targets
}

func covertToEndpoints(targets []*retrieval.Target) []*JobEndpoint {
	endpoints := []*JobEndpoint{}
	for _, endpoint := range targets {
		endpoints = append(endpoints, &JobEndpoint{
			Endpoint: endpoint.URL().String(),
			Health: string(endpoint.Health()),
		})
	}
	return endpoints
}

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

		c.JSON(200, ToTargets(tps))

	})
	r.Run()

}

