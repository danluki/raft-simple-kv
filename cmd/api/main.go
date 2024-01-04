package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	httpdelivery "github.com/danluki/kv-raft/internal/delivery/http"
	"github.com/danluki/kv-raft/internal/fsm"
	"github.com/danluki/kv-raft/internal/server"
	"github.com/dgraph-io/badger/v2"
	"github.com/hashicorp/raft"
	raftboltdb "github.com/hashicorp/raft-boltdb"
	"github.com/urfave/cli/v2"
	"golang.org/x/sync/errgroup"
)

func main() {
	app := cli.NewApp()
	app.Name = "Raft key value store"
	app.Usage = "One node of raft key value cluster"
	app.Version = "0.0.1"
	app.ArgsUsage = " "
	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:    "storage-path",
			Aliases: []string{"storage"},
			Usage:   "storage path",
			Value:   "tmp/storage",
		},
		&cli.StringFlag{
			Name:    "raft-port",
			Aliases: []string{"r-p"},
			Usage:   "port to use",
			Value:   "1111",
		},
		&cli.StringFlag{
			Name:    "port",
			Aliases: []string{"p"},
			Usage:   "port to use",
			Value:   "3000",
		},
		&cli.StringFlag{
			Name:    "node-id",
			Aliases: []string{"id", "node"},
			Usage:   "node id",
			Value:   "1",
		},
	}
	app.Commands = []*cli.Command{
		{
			Name:      "start",
			Usage:     "Start the server",
			ArgsUsage: " ",
			Before: func(c *cli.Context) error {
				if c.Args().Len() > 0 {
					return cli.Exit("Error: start takes no arguments", 1)
				}

				return nil
			},
			Action: func(c *cli.Context) error {
				storagePath := c.String("storage-path")
				raftPort := c.String("raft-port")
				nodeId := c.String("node-id")
				port := c.String("port")

				log.Println("Storage path", storagePath)
				log.Println("Port", port)
				log.Println("Node id", nodeId)
				log.Println("Raft port", raftPort)

				ctx, stop := signal.NotifyContext(
					context.Background(),
					os.Interrupt,
					syscall.SIGTERM,
					syscall.SIGQUIT,
				)
				defer stop()

				badgerOpt := badger.DefaultOptions(storagePath)
				badgerDB, err := badger.Open(badgerOpt)
				if err != nil {
					log.Fatal(err)
					return err
				}

				defer func() {
					if err := badgerDB.Close(); err != nil {
						_, _ = fmt.Fprintf(os.Stderr, "error close badger db %s\n", err.Error())
					}
				}()

				raftAddr := fmt.Sprintf("localhost:%s", raftPort)

				raftConf := raft.DefaultConfig()
				raftConf.LocalID = raft.ServerID(nodeId)
				raftConf.SnapshotThreshold = 1024

				fsmStore := fsm.NewBadger(badgerDB)

				store, err := raftboltdb.NewBoltStore(filepath.Join(storagePath, "raft.dataRepo"))
				if err != nil {
					log.Fatal(err)
					return err
				}

				cacheStore, err := raft.NewLogCache(512, store)
				if err != nil {
					log.Fatal(err)
					return err
				}

				snapshotStore, err := raft.NewFileSnapshotStore(storagePath, 2, os.Stdout)
				if err != nil {
					log.Fatal(err)
					return err
				}

				tcpAddr, err := net.ResolveTCPAddr("tcp", raftAddr)
				if err != nil {
					log.Fatal(err)
					return err
				}

				transport, err := raft.NewTCPTransport(
					raftAddr,
					tcpAddr,
					3,
					10*time.Second,
					os.Stdout,
				)
				if err != nil {
					log.Fatal(err)
					return err
				}

				raftServer, err := raft.NewRaft(
					raftConf,
					fsmStore,
					cacheStore,
					store,
					snapshotStore,
					transport,
				)
				if err != nil {
					log.Fatal(err)
					return err
				}

				configuration := raft.Configuration{
					Servers: []raft.Server{
						{
							ID:      raft.ServerID(nodeId),
							Address: transport.LocalAddr(),
						},
					},
				}

				raftServer.BootstrapCluster(configuration)
				handler := httpdelivery.NewHandler(httpdelivery.Opts{
					Raft: raftServer,
					DB:   badgerDB,
				})

				// Start the server
				server := server.NewServer(fmt.Sprintf(":%s", port), handler.Init())

				g, gCtx := errgroup.WithContext(ctx)
				g.Go(
					func() error {
						if err := server.Run(); err != nil && err != http.ErrServerClosed {
							server.E.Logger.Fatal("shutting down the server")
						}

						return nil
					},
				)
				g.Go(func() error {
					<-gCtx.Done()

					log.Println("Exiting server...")
					ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
					defer cancel()

					if err := server.E.Shutdown(ctx); err != nil {
						server.E.Logger.Fatal(err)
					}

					return nil
				})

				if err := g.Wait(); err != nil {
					panic(err)
				}

				return nil
			},
		},
	}

	_ = app.Run(os.Args)
}
