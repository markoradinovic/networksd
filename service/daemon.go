package service

import (
	"context"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/julienschmidt/httprouter"
	"github.com/spf13/viper"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

func StartDaemon(conf Conf) {

	log.SetFormatter(&log.TextFormatter{})
	if viper.GetBool("debug") {
		log.SetLevel(log.DebugLevel)
		log.Info("Enabled Debug log level")
	}
	log.Info("Using config file: ", viper.ConfigFileUsed())

	ctx := context.Background()

	svc := NewNetworkAddressService(conf)

	router := httprouter.New()
	initTransport(ctx, router, svc)

	errs := make(chan error)
	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errs <- fmt.Errorf("%s", <-c)
	}()

	server := &http.Server{Handler: router}
	unixSocket := viper.GetString("unix-socket")

	go func() {
		//debugPrint("Listening and serving HTTP on unix:/%s", file)
		//defer func() { debugPrintError(err) }()

		os.Remove(unixSocket)
		listener, err := net.Listen("unix", unixSocket)
		if err != nil {
			errs <- err
			return
		}
		defer listener.Close()
		log.Info("Listening on unix://", unixSocket)

		errs <- server.Serve(listener)
		//errs <- http.Serve(listener, router)
	}()

	httpAddr := ":" + strconv.Itoa(viper.GetInt("port"))
	go func() {
		log.Info("Listening on http://0.0.0.0", httpAddr)
		errs <- http.ListenAndServe(httpAddr, router)
	}()

	err := <-errs
	timeoutContext, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(timeoutContext); err != nil {
		log.Fatal("Daemon Shutdown: ", err)
	}
	os.Remove(unixSocket)
	log.Info("Daemon Shutdown. [", err, "]")

}

func initTransport(ctx context.Context, router *httprouter.Router, svc NetworkAddressService) {
	router.GET("/", Index(ctx, svc))
	router.POST("/api/bridge", CreateBridgeNetwork(ctx, svc))
	router.POST("/api/bridge/:name", CreateBridgeNetwork(ctx, svc))
	router.POST("/api/overlay", CreateOverlayNetwork(ctx, svc))
	router.POST("/api/overlay/:name", CreateOverlayNetwork(ctx, svc))
}
