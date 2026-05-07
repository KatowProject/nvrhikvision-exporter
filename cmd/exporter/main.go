package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/KatowProject/nvrhikvision-exporter/internal/client"
	"github.com/KatowProject/nvrhikvision-exporter/internal/collector"
	"github.com/KatowProject/nvrhikvision-exporter/internal/config"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	configFile := flag.String("config", "config.yaml", "Path to config file")
	nvrIP := flag.String("ip", "", "IP Address of Hikvision NVR (legacy mode)")
	user := flag.String("user", "admin", "NVR Username (legacy mode)")
	pass := flag.String("pass", "cctv12345", "NVR Password (legacy mode)")
	port := flag.String("port", "", "Port to expose metrics")
	flag.Parse()

	var cfg config.Config

	if fileConfig, err := config.LoadConfigFromFile(*configFile); err == nil {
		cfg = *fileConfig
	} else {
		if *nvrIP == "" {
			log.Fatalf("No configuration found. Use one of:\n" +
				"  1. Config file: -config=config.yaml\n" +
				"  2. CLI flags: -ip=10.47.10.3 -user=admin -pass=password")
		}
		cfg = *config.LoadConfigFromFlags(*nvrIP, *user, *pass, "9102")
		log.Printf("Running in legacy mode for single NVR: %s", *nvrIP)
	}

	// Override the port if it is specified via a flag.
	if *port != "" {
		cfg.Server.Port = *port
	}

	if len(cfg.NVRs) == 0 {
		log.Fatalf("No NVRs configured")
	}

	log.Printf("Starting Hikvision Exporter for %d NVR(s)", len(cfg.NVRs))

	var targets []*collector.NVRTarget
	for _, nvrConfig := range cfg.NVRs {
		c := client.NewClient(nvrConfig.IP, nvrConfig.Username, nvrConfig.Password)
		target := &collector.NVRTarget{
			Client: c,
			IP:     nvrConfig.IP,
			Name:   nvrConfig.Name,
		}
		targets = append(targets, target)
		log.Printf("Added NVR target: %s (%s)", nvrConfig.Name, nvrConfig.IP)
	}

	col := collector.NewCollector(targets)
	prometheus.MustRegister(col)
	log.Printf("Registered collector for %d NVR(s)", len(targets))

	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
			<head><title>Hikvision Exporter</title></head>
			<body>
			<h1>Hikvision NVR Exporter</h1>
			<p><a href="/metrics">Metrics</a></p>
			</body>
			</html>`))
	})

	log.Printf("Listening on port :%s", cfg.Server.Port)
	if err := http.ListenAndServe(":"+cfg.Server.Port, nil); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
