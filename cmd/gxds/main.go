// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2017 Canonical Ltd
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"bitbucket.org/tonyespy/gxds/daemon"
)

var flags struct {
	configPath *string
}

func init() {
	fmt.Fprintf(os.Stdout, "Init called\n")

	flags.configPath = flag.String("config", "./daemon-config.json", "daemon configuration file")
}

func main() {

	flag.Parse()

	if err := startDaemon(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func startDaemon() error {

	d, err := daemon.New()
	if err != nil {
		return err
	}

	// TODO: create a ProtocolHandler implementation
	// and pass to d.Init()
	if err := d.Init(flags.configPath, nil); err != nil {
		return err
	}

	d.Version = "0.1"

	d.Start()

	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	select {
	case sig := <-ch:
		fmt.Fprintf(os.Stderr, "Exiting on %s signal.\n", sig)
	}

	return d.Stop()
}
