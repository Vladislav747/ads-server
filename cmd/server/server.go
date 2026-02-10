package main

import (
	"github.com/VyacheslavGoryunov/simple-ads-server/internal/ads"
	"github.com/VyacheslavGoryunov/simple-ads-server/internal/metrics"
	"github.com/VyacheslavGoryunov/simple-ads-server/internal/stats"
	"github.com/VyacheslavGoryunov/simple-ads-server/internal/stats/clickhouse"
	"github.com/oschwald/geoip2-golang"
	"log"
	"time"
)

const (
	user = "default_new"
	pass = "qwerty123"
)

func main() {
	reader, err := geoip2.Open("GeoLite2-Country.mmdb")
	if err != nil {
		log.Fatal(err)
	}

	cw, err := clickhouse.NewClickhouseWriter("127.0.0.1", 18123, "rotator", "statistics",
		user, pass)
	if err != nil {
		log.Fatal(err)
	}

	statsManager := stats.NewManager(cw, 10*time.Second)
	// Начинаем запись статистики в ClickHouse
	statsManager.Start()

	go func() {
		_ = metrics.Listen("127.0.0.1:8082")
	}()

	s := ads.NewServer(reader, statsManager)

	if err := s.Listen(); err != nil {
		log.Fatal(err)
	}
}
