package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"time"

	"codereliant.io/cless/admin"
	"codereliant.io/cless/container"
	"codereliant.io/cless/db"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

var containerManager container.ContainerManager
var svcDefinitionManager *admin.ServiceDefinitionManager
var gormDbInstance *gorm.DB
var err error
var srv = &http.Server{
	Addr: ":80",
}

func main() {
	// logging
	debug := flag.Bool("debug", false, "sets log level to debug")
	flag.Parse()
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if *debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	// sqlite db instance
	gormDbInstance, err = db.NewSqliteDB()
	if err != nil {
		log.Error().Err(err).Msg("Failed to create sqlite db")
		panic(err)
	}

	// admin service/server
	repo := admin.NewSqliteServiceDefinitionRepository(gormDbInstance)
	svcDefinitionManager = admin.NewServiceDefinitionManager(repo)
	go admin.StartAdminServer(svcDefinitionManager)

	// container manager
	containerManager, err = container.NewDockerContainerManager(svcDefinitionManager)
	if err != nil {
		fmt.Printf("Failed to create container manager: %s\n", err)
		return
	}

	// setup http server
	http.HandleFunc("/", handler)
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal().Err(err).Msg("Failed to start http server")
		}
	}()

	// gracefull shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal().Err(err).Msg("Failed to start http server")
	}
	errList := containerManager.StopAndRemoveAllContainers()
	if len(errList) > 0 {
		log.Error().Errs("errors", errList).Msg("Failed to stop and remove containers")
	} else {
		log.Info().Msg("Stopped and removed all containers")
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	// handle admin requests
	if r.Host == admin.AdminHost {
		proxyToURL(w, r, fmt.Sprintf("%s:%d", "localhost", admin.AdminPort))
		return
	}
	svc, err := svcDefinitionManager.GetServiceDefinitionByHost(r.Host)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get service definition")
		w.Write([]byte("Failed to get service definition"))
		return
	}

	svcVersion := svc.ChooseVersion()
	log.Debug().Str("host", r.Host).Uint("service version", svcVersion).Msg("choosing service version")

	svcLocalHost, err := containerManager.GetRunningServiceForHost(r.Host, svcVersion)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get running service")
		w.Write([]byte("Failed to get running service"))
		return
	}
	log.Debug().Str("host", r.Host).Str("service localhost", *svcLocalHost).Msg("proxying request")
	proxy := httputil.NewSingleHostReverseProxy(&url.URL{
		Scheme: "http",
		Host:   *svcLocalHost,
	})
	proxy.ServeHTTP(w, r)
}

func proxyToURL(w http.ResponseWriter, r *http.Request, pURL string) {
	proxy := httputil.NewSingleHostReverseProxy(&url.URL{
		Scheme: "http",
		Host:   pURL,
	})
	proxy.ServeHTTP(w, r)
}
