/*
Copyright © 2019 Stephen Hallett

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/rjeczalik/notify"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type watchCmd struct {
}

var watchCmdLong = ``

func newWatchCmd() *cobra.Command {
	watch := &watchCmd{}

	cmd := &cobra.Command{
		Use:   "watch",
		Short: "Watch the configured save path for changes",
		Long:  watchCmdLong,
		Run: func(cmd *cobra.Command, args []string) {
			watch.run()
		},
	}

	return cmd
}

func init() {
	root.AddCommand(newWatchCmd())
}

func (c *watchCmd) run() {
	minecraftDir := strings.TrimSuffix(viper.GetString("saves-path"), "/")

	ch := make(chan notify.EventInfo, 1)

	if err := notify.Watch(minecraftDir+"/...", ch, notify.Create, notify.Write, notify.Remove); err != nil {
		log.Fatal(err)
	}
	defer notify.Stop(ch)

	for {

		var eis []notify.EventInfo
	Listen:
		for {
			select {
			case ei := <-ch:
				log.Println("Got event:", ei)
				eis = append(eis, ei)
			case <-time.After(1 * time.Second):
				break Listen
			}
		}

		if len(eis) > 0 {
			fmt.Println("Got events:", eis)
		}
	Eis:
		for _, e := range eis {
			if e.Event() == notify.Write {
				path := strings.TrimPrefix(e.Path(), minecraftDir)
				path = strings.TrimPrefix(path, "/")
				fmt.Println("p:", path)
				pc := strings.Split(path, "/")
				fmt.Println("pc:", pc)
				var dir string
				for i, comp := range pc {
					fmt.Println("c:", comp)
					if i == 0 {
						// Assuming we're given the save dir this will be the world dir
						dir = comp
						fmt.Println("Assuming the directory is:", dir)
						continue
					}
					if i == len(pc)-1 {
						// This will be the file
						fs := strings.Split(comp, ".")
						fmt.Println("fs:", fs)
						if len(fs) > 0 {
							ext := fs[len(fs)-1]
							fmt.Println("Extension is:", ext)
							if ext == "mca" {
								err := runBackup(dir)
								if err != nil {
									log.Fatal(err)
								}
								break Eis
							}
						}
						continue
					}

				}
			}
		}

	}

}
