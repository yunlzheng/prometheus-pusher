package main

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/prometheus/retrieval"
	"github.com/prometheus/prometheus/storage"
	"github.com/prometheus/prometheus/config"
	"fmt"
	"sort"
	"github.com/prometheus/common/model"
)

var cfg = struct {
	configFile string
}{}

type TargetPools struct {
	pools map[string][]*retrieval.Target
}

func init() {
}

func main() {

	var (
		sampleAppender = storage.Fanout{}
	)

	var (
		targetManager = retrieval.NewTargetManager(sampleAppender)
	)

	cfg.configFile = "prometheus.yml"

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
		for _, t := range targetManager.Targets() {
			job := string(t.Labels()[model.JobLabel])
			//fmt.Println(job, "<<<<<<<<", t)
			tps[job] = append(tps[job], t)
		}

		for _, targets := range tps {
			//fmt.Println(targets, "<<<< targets")
			sort.Slice(targets, func(i, j int) bool {
				return targets[i].Labels()[model.InstanceLabel] < targets[j].Labels()[model.InstanceLabel]
			})
		}

		pools := &TargetPools{
			pools: tps,
		}

		for job, pool := range pools.pools {
			fmt.Println("job %s", job)
			for _, endpoint := range pool {
				fmt.Println(endpoint, "<<<endpoint")
			}
		}

		c.JSON(200, gin.H{
			"pools":pools,
		})

	})
	r.Run()

}

