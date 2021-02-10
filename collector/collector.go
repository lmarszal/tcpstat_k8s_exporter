package collector

import (
	"context"
	"fmt"
	"github.com/lmarszal/tcpstat_k8s_exporter/docker"
	"github.com/prometheus/client_golang/prometheus"
	"log"
	"path"
	"strconv"
	"time"
)

func New(docker *docker.ClientWithCache, namespace string) *Collector {
	return &Collector{
		docker:    docker,
		namespace: namespace,
	}
}

type Collector struct {
	docker    *docker.ClientWithCache
	namespace string
}

var connectionStatesDesc = prometheus.NewDesc(
	prometheus.BuildFQName("pod_exporter", "tcp", "connection_states"),
	"Number of connections by state, pod and namespace",
	[]string{"pod", "namespace", "state"},
	nil,
)

func (c *Collector) Describe(d chan<- *prometheus.Desc) {
	d <- connectionStatesDesc
}

func (c *Collector) Collect(m chan<- prometheus.Metric) {
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	sandboxes, err := c.docker.ListPodSandboxes(ctx, c.namespace)
	if err != nil {
		log.Printf("Error listing pod sandboxes: %+v", err)
		return
	}

	for _, s := range sandboxes {
		err = c.update(s, m)
		if err != nil {
			log.Printf("Error updating metrics values for pod %s: %+v", s.PodName, err)
			return
		}
	}
}

func (c *Collector) update(sandbox docker.PodSandbox, m chan<- prometheus.Metric) error {
	statsFile := path.Join("/proc", strconv.Itoa(sandbox.Pid), "net", "tcp")
	tcpStats, err := getTCPStats(statsFile)
	if err != nil {
		return fmt.Errorf("couldn't get tcpstats: %w", err)
	}

	for st, value := range tcpStats {
		m <- prometheus.MustNewConstMetric(connectionStatesDesc, prometheus.GaugeValue, value, sandbox.PodName, sandbox.Namespace, st.String())
	}

	return nil
}
