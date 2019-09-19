package cmd

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type configCmd struct {
	savesPath  string
	backupPath string
	force      bool
	keepCount  int
}

var configCmdLong = ``

func newConfigCmd() *cobra.Command {

	cfg := &configCmd{}

	cmd := &cobra.Command{
		Use:   "config",
		Short: "Generate configuration",
		Long:  configCmdLong,
		Run: func(cmd *cobra.Command, args []string) {
			cfg.run()
		},
	}

	f := cmd.Flags()
	f.StringVarP(&cfg.savesPath, "saves-path", "s", "", "The directory where minecraft writes world save data")
	f.StringVarP(&cfg.backupPath, "target-path", "t", "", "Target path to write backup archives, recommend Dropbox or similar cloud storage")
	f.IntVarP(&cfg.keepCount, "keep-count", "k", 10, "Number of backups (per world) to keep, default 10")
	f.BoolVarP(&cfg.force, "force", "f", false, "Force overwriting of existing config file")

	viper.BindPFlag("saves-path", f.Lookup("saves-path"))
	viper.BindPFlag("target-path", f.Lookup("target-path"))
	viper.BindPFlag("keep-count", f.Lookup("keep-count"))

	return cmd
}

func init() {
	root.AddCommand(newConfigCmd())
}

func (c *configCmd) run() {

	var err error

	if c.force {
		err = viper.WriteConfig()
	} else {
		err = viper.SafeWriteConfig()
	}

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Wrote config file")
}

var (
	defaultConfigPathWin    = `%s\.mcbackup`
	defaultConfigPath       = `%s/.mcbackup`
	defaultPathWin          = `%s\AppData\.minecraft\saves`
	defaultPathDarwin       = `%s/Library/Application Support/minecraft/saves`
	defaultPathLinux        = `%s/.minecraft/saves`
	defaultTargetPathWin    = `%s\AppData\.minecraft\backups`
	defaultTargetPathDarwin = `%s/Library/Application Support/minecraft/backups`
	defaultTargetPathLinux  = `%s/.minecraft/backups`
)

func LoadConfig() {

	viper.SetConfigName("config")
	viper.AddConfigPath("$HOME/.mcbackup")
	viper.AddConfigPath(".")

	viper.SetEnvPrefix("mcbackup")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()

	switch runtime.GOOS {
	case "windows":
		viper.AddConfigPath(expand(defaultConfigPathWin))
		viper.SetDefault("saves-path", expand(defaultPathWin))
		viper.SetDefault("target-path", expand(defaultTargetPathWin))
	case "linux":
		viper.AddConfigPath(expand(defaultConfigPath))
		viper.SetDefault("saves-path", expand(defaultPathLinux))
		viper.SetDefault("target-path", expand(defaultTargetPathLinux))
	default:
		// Default to Darwin
		viper.AddConfigPath(expand(defaultConfigPath))
		viper.SetDefault("saves-path", expand(defaultPathDarwin))
		viper.SetDefault("target-path", expand(defaultTargetPathDarwin))
	}

	err := viper.ReadInConfig()

	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			// 	trimTrailingSlash()
			// 	cfg := &configCmd{
			// 		savesPath:  viper.GetString("saves-path"),
			// 		backupPath: viper.GetString("target-path"),
			// 	}
			// 	cfg.run()
			log.Println("Unable to load config file: ", err)
		}

		// }
	}

	trimTrailingSlash()
}

func trimTrailingSlash() {
	if strings.HasSuffix(viper.GetString("saves-path"), "/") {
		viper.Set("saves-path", strings.TrimSuffix(viper.GetString("saves-path"), "/"))
	}
	if strings.HasSuffix(viper.GetString("target-path"), "/") {
		viper.Set("target-path", strings.TrimSuffix(viper.GetString("target-path"), "/"))
	}
}

func expand(path string) string {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	return fmt.Sprintf(path, home)
}
