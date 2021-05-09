package adapter

import (
	"context"
	"crypto/tls"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/go-logr/logr"
	"github.com/gogo/protobuf/proto"
	"github.com/golang/snappy"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/prometheus/prompb"
	"github.com/phin1x/prometheus-cloudwatch-adapter/pkg/handlers"
)

func New(config *Configuration, logger logr.Logger) (*adapter, error) {
	sess, err := session.NewSession()
	if err != nil {
		return nil, err
	}

	svc := cloudwatch.New(sess)
	svc.Handlers.Build.PushBackNamed(handlers.NewRequestCompressionHandler(logger))

	return &adapter{
		cw:     svc,
		config: config,
		logger: logger,
	}, nil
}

type adapter struct {
	cw     *cloudwatch.CloudWatch
	config *Configuration
	logger logr.Logger
}

func (a *adapter) Run() error {
	if a.config.Debug {
		a.logger.Info("cloudwatch namespace: " + a.config.CloudwatchNamespace)
	}

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("/health", func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/write", a.write)

	server := &http.Server{
		Addr:    a.config.ListenAddress,
		Handler: mux,
		TLSConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
		},
	}

	done := make(chan bool)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-quit
		a.logger.Info("server is shutting down...")

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		server.SetKeepAlivesEnabled(false)
		if err := server.Shutdown(ctx); err != nil {
			a.logger.Error(err, "could not gracefully shutdown the server")
		}

		close(done)
	}()

	a.logger.Info("server is ready to handle requests at " + a.config.ListenAddress)

	var err error
	if a.config.TLSCert != "" && a.config.TLSKey != "" {
		err = server.ListenAndServeTLS(a.config.TLSCert, a.config.TLSKey)
	} else {
		err = server.ListenAndServe()
	}
	if err != nil && err != http.ErrServerClosed {
		a.logger.Error(err, "could not listen on "+a.config.ListenAddress)
		return err
	}

	<-done
	a.logger.Info("server stopped")

	return nil
}

func (a *adapter) write(w http.ResponseWriter, r *http.Request) {
	compressed, err := ioutil.ReadAll(r.Body)
	if err != nil {
		a.logger.Error(err, "")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	reqBuf, err := snappy.Decode(nil, compressed)
	if err != nil {
		a.logger.Error(err, "Decode error")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var promReq prompb.WriteRequest
	if err := proto.Unmarshal(reqBuf, &promReq); err != nil {
		a.logger.Error(err, "Unmarshal error")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := a.processWriteRequest(&promReq); err != nil {
		a.logger.Error(err, "failed to write data to cloudwatch")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
