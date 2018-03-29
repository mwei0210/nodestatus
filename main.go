package main

import (
	"encoding/json"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

var (
	up = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "total_up_nodes",
		Help: "Current number of up node.",
	})
	down = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "total_down_nodes",
		Help: "Current number of down node.",
	})
	inactive = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "total_inactive_nodes",
		Help: "Current number of inactive node.",
	})
	url = "https://securenodes2.eu.zensystem.io/api/grid/nodes"
)

func init() {
	// Metrics have to be registered to be exposed:
	prometheus.MustRegister(up)
	prometheus.MustRegister(down)
	prometheus.MustRegister(inactive)
}

func main() {

	//CLI setup using cobra
	var duration int

	var rootCmd = &cobra.Command{
		Use: "app",
	}
	rootCmd.PersistentFlags().IntVarP(&duration, "duration", "d", 5, "Duration of interval between calls.")
	rootCmd.Execute()

	//loop in interval to get node status
	ticker := time.NewTicker(time.Second * time.Duration(duration))

	go func() {
		for range ticker.C {

			response, err := http.Get(url)

			if err != nil {
				fmt.Println(err)
				return
			}

			textBytes, err := ioutil.ReadAll(response.Body)

			if err != nil {
				fmt.Println(err)
				return
			}

			ns := nodeStatus{}
			if err := json.Unmarshal(textBytes, &ns); err != nil {
				panic(err)
			}
			up.Set(ns.Userdata.Up)
			down.Set(ns.Userdata.Down)
			inactive.Set(ns.Userdata.Inactive)

			fmt.Printf("UP: %.f\tDOWN: %.f\tINACTIVE: %.f\n", ns.Userdata.Up, ns.Userdata.Down, ns.Userdata.Inactive)
			fmt.Println("Nodes status updated.")

		}
	}()

	// The Handler function provides a default handler to expose metrics
	// via an HTTP server. "/metrics" is the usual endpoint for that.
	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(":8080", nil))
}

type nodeStatus struct {
	Userdata struct {
		Up       float64 `json:"up"`
		Down     float64 `json:"down"`
		Inactive float64 `json:"inactive"`
	} `json:"userdata"`
}
