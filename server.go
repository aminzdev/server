package server

import (
	"fmt"
	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
	"net/http"
	"sync"
)

func RunGrpcServer(service, host string, grpcServer *grpc.Server, wg *sync.WaitGroup) {
	defer wg.Done()
	listener, err := net.Listen("tcp", host)
	if err != nil {
		log.Fatalf("failed to listen on %s, %s\n", host, err)
	}
	fmt.Printf("%s grpc serving on %s\n", service, host)
	err = grpcServer.Serve(listener)
	if err != nil {
		log.Fatalf("%s failed to serve on %s, %s\n", service, host, err)
	}
}

func RunGrpcWebServer(service, host string, grpcServer *grpc.Server, wg *sync.WaitGroup) {
	defer wg.Done()
	grpcWebServer := grpcweb.WrapServer(
		grpcServer,
		grpcweb.WithOriginFunc(func(origin string) bool { return true }), // Enable CORS
	)
	srv := &http.Server{
		Handler: grpcWebServer,
		Addr:    host,
	}
	fmt.Printf("%s http serving on %s\n", service, srv.Addr)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("%s failed to serve http on %s, %s\n", service, host, err)
	}
}

func RunServer(server interface{}, serviceName, grpcHost, grpcWebHost string, registerServer func(*grpc.Server, interface{})) {
	grpcServer := grpc.NewServer()
	registerServer(grpcServer, server)
	reflection.Register(grpcServer)

	wg := &sync.WaitGroup{}
	if grpcHost != "" {
		wg.Add(1)
		go RunGrpcServer(serviceName, grpcHost, grpcServer, wg)
	}
	if grpcWebHost != "" {
		wg.Add(1)
		go RunGrpcWebServer(serviceName, grpcWebHost, grpcServer, wg)
	}
	wg.Wait()
}
