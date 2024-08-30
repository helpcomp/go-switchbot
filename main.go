package main

import (
	"github.com/alecthomas/kong"
	"github.com/nasa9084/go-switchbot/v3/prom"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"net/http"
	"os"
)

var cli struct {
	MetricsPath     string `env:"EXPORTER_METRICS_PATH" help:"${env} - Path under which to expose metrics" default:"/metrics"`
	DefaultEndpoint string `env:"DEFAULT_ENDPOINT" help:"${env} - Switchbot API Endpoint" default:"https://api.switch-bot.com"`
	ListenAddress   string `env:"EXPORTER_LISTEN_ADDRESS"  help:"${env} - Address to listen on for web interface and telemetry" default:":9617"`
	Token           string `env:"SWITCHBOT_TOKEN" help:"${env} - Switchbot Developer Token" required:""`
	Key             string `env:"SWITCHBOT_KEY" help:"${env} - Switchbot Developer Key" required:""`
}

func main() {
	//region Initialization
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	kong.Parse(&cli)

	// Set up Switchbot, and refresh device data
	prom.New(cli.Token, cli.Key)

	prometheus.MustRegister(prom.NewExporter())
	http.Handle(cli.MetricsPath, promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte(`<html>
					<head><title>B2 Exporter</title></head>
		            <body>
		            <h1>B2 Exporter</h1>
		            <p><a href='` + cli.MetricsPath + `'>Metrics</a></p>
		            </body>
		            </html>`))
		if err != nil {
			return
		}
	})

	log.Info().Msgf("⚡ Starting HTTP server http://127.0.0.1%s%s on listen address %s and metric path %s", cli.ListenAddress, cli.MetricsPath, cli.ListenAddress, cli.MetricsPath)

	if err := http.ListenAndServe(cli.ListenAddress, nil); err != nil {
		log.Fatal().Err(err).Msgf("⛔️ %v", err)
	}
}
