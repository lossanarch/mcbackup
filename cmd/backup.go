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
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"

	gosxnotifier "github.com/deckarep/gosx-notifier"
	"github.com/lossanarch/mcbackup/pkg/targz"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	root.AddCommand(newBackupCmd())
}

type backupCmd struct {
	keepCount int
	save      string
}

var backupCmdLong = ``

func newBackupCmd() *cobra.Command {

	bak := &backupCmd{}

	cmd := &cobra.Command{
		Use:   "backup <world name>",
		Short: "Take a backup of a world now",
		Long:  backupCmdLong,
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			bak.save = args[0]

			bak.run()
		},
	}

	f := cmd.Flags()
	f.IntVarP(&bak.keepCount, "keep-count", "k", 10, "Number of backups (per world) to keep, default 10")

	viper.BindPFlag("keep-count", f.Lookup("keep-count"))

	return cmd
}

func (c *backupCmd) run() {
	err := runBackup(c.save)
	if err != nil {
		log.Fatal(err)
	}
}

func runBackup(dir string) error {

	err := deleteExpired(dir)
	if err != nil {
		return err
	}

	// tar and gz dir
	ls, _ := targz.ReadDir(viper.GetString("saves-path") + "/" + dir)

	var fs []*os.File
	for _, file := range ls {
		f, err := os.Open(file)
		if err != nil {
			return err
		}
		defer f.Close()

		fs = append(fs, f)
	}

	dt, err := time.Now().Local().Round(time.Minute).MarshalText()

	if err != nil {
		return err
	}

	filename := "/Users/steve/Dropbox/minecraft/" + string(dt) + "_" + dir
	err = targz.BufferedReadWrite(viper.GetString("saves-path"), filename, fs)
	if err != nil {
		return err
	}

	// save somewhere

	pushNotification(dir)

	return nil
}

func deleteExpired(dir string) error {
	backupFiles, err := ioutil.ReadDir(viper.GetString("target-path"))
	if err != nil {
		log.Fatal(err)
	}
	// spew.Dump(backupFiles)

	r := regexp.MustCompile(`^\d{4}(?:-\d{2}){2}T\d{2}(?:\:\d{2}){2}\+\d{2}\:\d{2}_\w+\.tgz$`)

	backupsList := []os.FileInfo{}

	// Find relevant backups
	for _, file := range backupFiles {
		// fmt.Println("file", file.Name())
		if file.IsDir() {
			continue
		}
		if strings.Contains(file.Name(), dir) {
			// fmt.Println("contains", file.Name())
			if r.MatchString(file.Name()) {
				// fmt.Println("regexp", file.Name())
				backupsList = append(backupsList, file)
			}
		}
	}

	keepCount := viper.GetInt("keep-count")

	sort.Slice(backupsList, func(i, j int) bool {
		return backupsList[i].ModTime().After(backupsList[j].ModTime())
	})

	// spew.Dump(backupsList)

	// Expire oldest
	if len(backupsList) > keepCount {
		for i, b := range backupsList {
			if i >= keepCount {
				fmt.Println("Deleting", b.Name())
				err := os.Remove(viper.GetString("target-path") + "/" + b.Name())
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func pushNotification(worldName string) {
	note := gosxnotifier.NewNotification("Backup completed!")

	//Optionally, set a title
	note.Title = worldName

	//Optionally, set a subtitle
	note.Subtitle = "World Backup"

	//Optionally, set a group which ensures only one notification is ever shown replacing previous notification of same group id.
	note.Group = "com.lossanarch.mcbackup"

	//Optionally, set a sender (Notification will now use the Safari icon)
	note.Sender = "com.lossanarch.mcbackup"

	note.Sound = gosxnotifier.Submarine

	note.ContentImage = viper.GetString("saves-path") + "/" + worldName + "/" + "icon.png"

	// fmt.Println(note)

	note.Push()
}
