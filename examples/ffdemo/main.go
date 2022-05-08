/*
 * Copyright © 2019 Hedzr Yeh.
 */

package main

import (
	"fmt"
	"os"
	"time"

	"github.com/hedzr/cmdr"
	"github.com/hedzr/cmdr/flag"
)

var (
	treatAsLongOpt = flag.TreatAsLongOpt(true)

	serv = flag.String("service", "hello_service", "service name",
		flag.WithAction(func(cmd *cmdr.Command, args []string) (err error) {
			return
		}), flag.WithDescription("", ""))
	host           = flag.String("host", "localhost", "listening host")
	port           = flag.Int("port", 50001, "listening port")
	reg            = flag.String("reg", "localhost:32379", "register etcd address")
	count          = flag.Int("count", 3, "instance's count")
	connectTimeout = flag.Duration("connect-timeout", 5*time.Second, "connect timeout")

	republish = flag.Bool("republish", false, "re-publish the service config or not")

	ver = flag.String("ver", "1.0.3", "the simulating version")

	// `{
	//     "location": {
	//         "latitude": 416802456,
	//         "longitude": -742370183
	//     },
	//     "name": "352 South Mountain Road, Wallkill, NY 12589, USA"
	// }`
	latitude, longitude = 41.6802456, -74.2370183
)

// var servers []*grpc.Server

func init() {
	// flag.treatAsLongOpt = true

	// app logger
	// logrus.SetLevel(logrus.TraceLevel)
	// logex.Enable()

	// grpc logger
	// grpclog.SetLoggerV2(grpclog.NewLoggerV2WithVerbosity(os.Stdout, os.Stderr, os.Stderr, 9))

	// registry backend
}

func initRegistry() func() {
	// etcd3.NewRegistryBackend(*reg, registry.DefaultPrefix, registry.DefaultTTL)
	// registry.Start(etcd3.SchemeEtcd3)
	//
	// // weighted: [index, key.addr, key.version, key.addr.hash]
	// if err := registry.PublishServiceConfig(*serv, serverConfig, *republish); err != nil {
	// 	panic(err)
	// }
	//
	// return registry.Stop
	return func() {}
}

func main() {
	flag.Parse()

	fmt.Printf(`

	server: %v
	host: %v
	port: %v
	reg: %v
	count: %v
	connectTimeout: %v
	republish: %v
	ver: %v
	latitude, longitude: %v,%v

`,
		*serv, *host, *port, *reg, *count, *connectTimeout, *republish, *ver,
		latitude, longitude)
}

func doServer() {
	defer initRegistry()

	for p := *port; p < *port+*count; p++ {
		// logrus.Debugf("checking port %v and starting...", p)
		go serverRun(p, *ver)
	}

	defer func() {
		// logrus.Println("\nEND")
	}()

	go cmdr.TrapSignals(func(s os.Signal) {
		// for _, s := range servers {
		// 	s.Stop()
		// }
		// logrus.Infof("done")
	})()
}

func serverRun(port int, version string) {
	// lis, err := net.Listen("tcp", net.JoinHostPort(*host, strconv.Itoa(port)))
	// if err != nil {
	// 	panic(err)
	// }
	//
	// err = registry.Register(*serv, *host, port, registry.Meta{
	// 	"version":   version,
	// 	"latitude":  latitude,
	// 	"longitude": longitude,
	// })
	// if err != nil {
	// 	panic(err)
	// }
	//
	// logrus.Infof(">> starting hello service at %v", port)
	// s := grpc.NewServer()
	// greeter.RegisterGreeterServer(s, &server{port: port, version: version})
	// servers = append(servers, s)
	// if err = s.Serve(lis); err != nil {
	// 	logrus.Error(err)
	// }
}

// // server is used to implement helloworld.GreeterServer.
// type server struct {
// 	port    int
// 	version string
// }
//
// // SayHello implements helloworld.GreeterServer
// func (s *server) SayHello(ctx context.Context, in *greeter.HelloRequest) (*greeter.HelloReply, error) {
// 	logrus.Infof("%v: Receive is %s\n", time.Now(), in.Name)
// 	return &greeter.HelloReply{Message: "Hello " + in.Name + " [v" + s.version + "] " + net.JoinHostPort(*host, strconv.Itoa(s.port))}, nil
// }
//
// const (
// 	serverConfig = `{
//  "loadBalancingPolicy": "weighted_round_robin",
//  "loadBalancingConfig": [{
//   "weighted_round_robin": {
// 	"healthCheck": true,
// 	"method": "key.version",
// 	"methods": {
// 	  "index": {
// 		"weights": { "-1":1, "1":59,"2":30,"3":10 }
// 	  },
// 	  "key.version": {
// 		"weights": { "<1.2.0":90,"~1.2.x":10 }
// 	  },
// 	  "key.addr.hash": {
// 		"weights": { "hash": 100 }
// 	  }
// 	}
//   }
//  }]
// }
// `
// )
