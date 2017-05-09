package model

import "github.com/prometheus/prometheus/retrieval"

type JobTargets struct {
	Targets []*jobTarget
}

type jobTarget struct {
	Name      string
	Endpoints []*jobEndpoint
}

type jobEndpoint struct {
	Endpoint string
	Health   string
}

func ToTargets(tps map[string][]*retrieval.Target) []*jobTarget {
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
