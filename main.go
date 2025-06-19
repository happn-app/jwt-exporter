package main

import (
	"flag"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"gopkg.in/yaml.v2"
	jwtexporter "happn.io/jwt-exporter/src"
)

func ExpandPath(path string) (string, error) {
	// First expand env vars like $HOME
	path = os.ExpandEnv(path)

	// Then expand ~
	if strings.HasPrefix(path, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		path = filepath.Join(home, strings.TrimPrefix(path, "~"))
	}

	return path, nil
}

type Config struct {
	ListenAddress   string        `yaml:"address"`
	LabelSelectors  []string      `yaml:"label_selectors"`
	PollingInterval time.Duration `yaml:"polling_interval"`
	KubeconfigPath  string        `yaml:"kubeconfig_path"`
	AnnotationKey   string        `yaml:"annotation_key"`
}

var (
	// Define command line flags
	configPath string
)

func init() {
	flag.StringVar(&configPath, "config", "/config/config.yaml", "Path to the configuration file")
	zap.LevelFlag("level", zap.InfoLevel, "Log level (debug, info, warn, error, dpanic, panic, fatal)")
}

func main() {
	flag.Parse()
	baseLogger, _ := zap.NewProduction()
	logger := baseLogger.With(
		zap.String("type", "jwt-exporter"),
	)
	defer logger.Sync() // flushes buffer, if any
	// Load the file; returns []byte
	f, err := os.ReadFile(configPath)
	if err != nil {
		logger.Fatal(err.Error(), zap.String("configPath", configPath), zap.String("type", "jwt-exporter"))
	}
	c := Config{
		ListenAddress:   ":8080",
		LabelSelectors:  []string{"monitor.jwt.io/monitoring=true"},
		PollingInterval: time.Hour,
		KubeconfigPath:  "",
		AnnotationKey:   "jwt-exporter/secret-key",
	}

	// Unmarshal our input YAML file into empty Car (var c)
	if err := yaml.UnmarshalStrict(f, &c); err != nil {
		logger.Fatal(
			"Encountered error while unmarshalling configuration",
			zap.Error(err), zap.String("configPath", configPath),
		)
	}
	logger.Info("Loaded configuration", zap.Reflect("config", c))

	jwtexporter.Init()
	logger.Info("Starting JWT Exporter")
	kubeConfigPath, err := ExpandPath(c.KubeconfigPath)
	if err != nil {
		logger.Fatal("Error expanding kubeconfig path", zap.Error(err), zap.String("kubeconfigPath", c.KubeconfigPath))
	}
	checker := jwtexporter.GetChecker(
		c.PollingInterval,
		c.LabelSelectors,
		kubeConfigPath,
		jwtexporter.GetExporter(logger.With(zap.String("component", "exporter"))),
		c.AnnotationKey,
		logger.With(zap.String("component", "checker")),
	)
	go checker.StartChecking()

	// Start the HTTP server to expose metrics
	handler := promhttp.HandlerFor(prometheus.DefaultGatherer, promhttp.HandlerOpts{})
	http.Handle("/metrics", handler)
	logger.Fatal(
		"Encountered error while serving prometheus exporter",
		zap.Error(http.ListenAndServe(c.ListenAddress, nil)),
	)
}
