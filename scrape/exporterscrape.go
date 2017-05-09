package scrape

import (
	"time"
	"github.com/prometheus/prometheus/retrieval"
	"github.com/prometheus/prometheus/config"
	"github.com/prometheus/common/model"
	"sort"
	"fmt"
	"net/http"
	"github.com/prometheus/common/version"
	"github.com/prometheus/common/expfmt"
	"io"
	"bytes"
	"strings"
	"io/ioutil"
	"os"
)

var pushGateway string

func init() {
	pushGateway = getOr("PUSH_GATEWAY", "http://pushgateway.example.org:9091")
}

type JobTargets struct {
	tm *retrieval.TargetManager
}

func NewJobTargets(tm *retrieval.TargetManager) *JobTargets {
	return &JobTargets{
		tm: tm,
	}
}

func (jt *JobTargets) Targets() []*jobTarget {
	tps := map[string][]*retrieval.Target{}
	for _, t := range jt.tm.Targets() {
		job := string(t.Labels()[model.JobLabel])
		tps[job] = append(tps[job], t)
	}

	for _, targets := range tps {
		sort.Slice(targets, func(i, j int) bool {
			return targets[i].Labels()[model.InstanceLabel] < targets[j].Labels()[model.InstanceLabel]
		})
	}

	targets := []*jobTarget{}
	for job, pool := range tps {
		targets = append(targets, &jobTarget{
			Name: job,
			Endpoints: covertToEndpoints(pool),
		})
	}
	return targets
}

func covertToEndpoints(targets []*retrieval.Target) []*jobEndpoint {
	endpoints := []*jobEndpoint{}
	for _, endpoint := range targets {
		endpoints = append(endpoints, &jobEndpoint{
			Endpoint: endpoint.URL().String(),
			Health: string(endpoint.Health()),
		})
	}
	return endpoints
}

type jobTarget struct {
	Name      string
	Endpoints []*jobEndpoint
}

func (jt *jobTarget) Scrape() {
	for _, endpoint := range jt.Endpoints {
		go func(endpoint *jobEndpoint) {
			endpoint.scrape(jt.Name)
		}(endpoint)
	}
}

const acceptHeader = `application/vnd.google.protobuf;proto=io.prometheus.client.MetricFamily;encoding=delimited;q=0.7,text/plain;version=0.0.4;q=0.3,*/*;q=0.1`

var userAgentHeader = fmt.Sprintf("Prometheus/%s", version.Version)

func (endpoint *jobEndpoint) scrape(jobName string) error {

	if jobName!="HostsMetrics" {
		return nil
	}

	req, err := http.NewRequest("GET", endpoint.Endpoint, nil)
	if err != nil {
		return err
	}
	req.Header.Add("Accept", acceptHeader)
	req.Header.Set("User-Agent", userAgentHeader)
	req.Header.Set("X-Prometheus-Scrape-Timeout-Seconds", fmt.Sprintf("%f", 15))
	//
	client := &http.Client{}
	//

	resp, err := client.Do(req)

	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if (resp.StatusCode != http.StatusOK) {
		return fmt.Errorf("server returned HTTP status %s", resp.Status)

	}

	var (
		allSamples = make(model.Samples, 0, 200)
		decSamples = make(model.Vector, 0, 50)
	)

	sdec := expfmt.SampleDecoder{
		Dec: expfmt.NewDecoder(resp.Body, expfmt.ResponseFormat(resp.Header)),
		Opts: &expfmt.DecodeOptions{
			Timestamp: model.TimeFromUnixNano(time.Now().UnixNano()),
		},
	}

	for {
		if err = sdec.Decode(&decSamples); err != nil {
			break
		}
		allSamples = append(allSamples, decSamples...)
		decSamples = decSamples[:0]
	}

	if err == io.EOF {
		// Set err to nil since it is used in the scrape health recording.
		err = nil
	}

	var buffer bytes.Buffer

	for _, sample := range allSamples {
		metric := fmt.Sprintf("%s %s\n", sample.Metric, sample.Value)
		if !strings.Contains(metric, "go_") && !strings.Contains(metric, "http_") {
			buffer.WriteString(metric)
		}
	}

	url := fmt.Sprintf("%s/metrics/job/%s", pushGateway, jobName)
	fmt.Println("send data to pushgateway :" + url)

	post, err := http.NewRequest("POST", url, strings.NewReader(buffer.String()))

	if err != nil {
		fmt.Println(err.Error(), "post data to push gateway error!!!!!!!!!!")
		return err
	}

	response2, err := client.Do(post)
	if err != nil {
		fmt.Println(err.Error(), "post data to push gateway error!!!!!!!!!!")
	} else {
		body, _ := ioutil.ReadAll(response2.Body)
		fmt.Println(string(body))
	}

	return nil

}

type jobEndpoint struct {
	Endpoint string
	Health   string
}

type ExporterScrape struct {
	jt             *JobTargets
	ticker         *time.Ticker
	quit           chan struct{}
	ScrapeInterval model.Duration
}

func NewExporterScrape(jt *JobTargets) *ExporterScrape {
	return &ExporterScrape{
		jt: jt,
		ticker: time.NewTicker(time.Second * 15),
		quit: make(chan struct{}),
	}
}

func (es *ExporterScrape) Run() {
	go func() {
		for {
			select {
			case <-es.ticker.C:
				for _, jt := range es.jt.Targets() {
					jt.Scrape()
				}
			case <-es.quit:
				es.ticker.Stop()
				return
			}
		}
	}()
}

func (es *ExporterScrape) Stop() {
	es.ticker.Stop()
}

func (es *ExporterScrape) AppConfig(conf *config.Config) {
	es.ScrapeInterval = conf.GlobalConfig.ScrapeInterval
}

func getOr(env string, value string) string {
	envValue := os.Getenv(env)
	if envValue == "" {
		return value
	}
	return envValue
}



