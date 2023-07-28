// Copyright 2022 bytetrade
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"app-store-server/pkg/apiserver"
	"flag"

	"github.com/golang/glog"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func main() {
	cmd := newAPPServerCommand()
	flag.Parse()
	defer glog.Flush()

	if err := cmd.Execute(); err != nil {
		glog.Fatalln(err)
	}
}

func newAPPServerCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "app-store-server",
		Short: "REST API for app-store",
		Long:  `The app-store-server provides REST API interfaces for the app-store`,
		Run: func(cmd *cobra.Command, args []string) {
			err := Run()
			if err != nil {
				glog.Fatalln(err)
			}
		},
	}

	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)

	return cmd
}

func Run() error {

	// new server
	s, err := apiserver.New()
	if err != nil {
		return err
	}

	if err = s.PrepareRun(); err != nil {
		return err
	}

	glog.Infof("Start listening on %s", s.Server.Addr)
	return s.Run()
}
