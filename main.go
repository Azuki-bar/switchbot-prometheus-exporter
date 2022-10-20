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
		Name:      "tempareture",
		Namespace: namespace,
	})
	humidity = prometheus.NewGauge(prometheus.GaugeOpts{
		Name:      "humidity",
		Namespace: namespace,
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
	Humidity    float64
}

func (s switchBotData) String() string {
	return fmt.Sprintf("temp: %.2f, Hum: %.2f", s.Tempareture, s.Humidity)
}
func registerMetrics(ctx context.Context, cd <-chan switchBotData, errC chan<- error) {
	for {
		select {
		case d := <-cd:
			log.Printf("set %s\n", d.String())
			tempareture.Set(d.Tempareture)
			humidity.Set(d.Humidity)
		case <-ctx.Done():
			break
		}
	}
}

func fetchData(ctx context.Context, cd chan<- switchBotData, errC chan<- error) {
	botClient := switchbot.New(conf.Token)
	do := func() {
		ctxFetch, cancel1 := context.WithDeadline(ctx, time.Now().Add(5*time.Second))
		defer cancel1()
		status, err := botClient.Device().Status(ctxFetch, conf.DeviceID)
		if err != nil {
			errC <- err
			return
		}
		cd <- switchBotData{
			Tempareture: status.Temperature,
			Humidity:    float64(status.Humidity),
		}
	}
	ticker := time.NewTicker(conf.FetchInterval)
	defer ticker.Stop()
	do()
	for {
		select {
		case <-ticker.C:
			do()
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

	select {
	case <-ctx.Done():
		log.Print("trap err")
	case err := <-errC:
		log.Print(err)
		cancel()
	}

	if err := srv.Shutdown(ctx); err != nil {
		log.Panic(err)
	}
}
