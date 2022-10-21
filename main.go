package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/caarlos0/env/v6"
	"github.com/nasa9084/go-switchbot"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	namespace   = "switchbot"
	tempareture = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "tempareture",
	})
	humidity = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "humidity",
	})
	conf = config{}
)

type config struct {
	FetchInterval time.Duration `env:"FETCH_INTERVAL" envDefault:"5m"`
	Port          int           `env:"PORT" envDefault:"8080"`
	Token         string        `env:"TOKEN,required"`
	DeviceID      string        `env:"DEVICE_ID,required"`
}
type switchBotData struct {
	Tempareture float64
	Humidity    int
}

func (s switchBotData) String() string {
	return fmt.Sprintf("Temp: %.1f, Hum: %d", s.Tempareture, s.Humidity)
}
func registerMetrics(ctx context.Context, cd <-chan switchBotData, errC chan<- error) {
LOOP:
	for {
		select {
		case d := <-cd:
			log.Printf("set %s\n", d)
			tempareture.Set(d.Tempareture)
			humidity.Set(float64(d.Humidity))
		case <-ctx.Done():
			break LOOP
		}
	}
}

func fetchData(ctx context.Context, cd chan<- switchBotData, errC chan<- error) {
	botClient := switchbot.New(conf.Token)
	do := func(ctx context.Context) {
		ctxFetch, cancel1 := context.WithTimeout(ctx, 5*time.Second)
		defer cancel1()
		status, err := botClient.Device().Status(ctxFetch, conf.DeviceID)
		if err != nil {
			errC <- err
			return
		}
		cd <- switchBotData{
			Tempareture: status.Temperature,
			Humidity:    status.Humidity,
		}
	}
	ticker := time.NewTicker(conf.FetchInterval)
	defer ticker.Stop()
	do(ctx)
	for {
		select {
		case <-ticker.C:
			go do(ctx)
		case <-ctx.Done():
			break
		}
	}
}
func init() {
	prometheus.MustRegister(tempareture)
	prometheus.MustRegister(humidity)
	if err := env.Parse(&conf); err != nil {
		log.Fatal(err)
	}
}
func main() {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	srv := &http.Server{
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		Addr:              fmt.Sprintf(":%d", conf.Port),
	}
	ctx := context.Background()
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	cd := make(chan switchBotData)
	errC := make(chan error)
	go fetchData(ctx, cd, errC)
	go registerMetrics(ctx, cd, errC)
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			errC <- err
		}
	}()

LOOP:
	for {
		select {
		case <-ctx.Done():
			log.Print(ctx.Err())
			log.Print("main context is done!")
			break LOOP
		case err := <-errC:
			if err != nil {
				log.Print(err)
				cancel()
			}
		}
	}

	if err := srv.Shutdown(ctx); err != nil {
		log.Panic(err)
	}
}
