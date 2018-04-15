package main

import (
	"path/filepath"
	"strings"

	"github.com/spacelavr/pandora/pkg/api"
	"github.com/spacelavr/pandora/pkg/core"
	"github.com/spacelavr/pandora/pkg/discovery"
	"github.com/spacelavr/pandora/pkg/log"
	"github.com/spacelavr/pandora/pkg/node"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	config string

	// CLI main command
	CLI = &cobra.Command{
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			abs, err := filepath.Abs(config)
			if err != nil {
				log.Fatal(err)
			}

			base := filepath.Base(abs)
			path := filepath.Dir(abs)

			viper.SetConfigName(strings.Split(base, ".")[0])
			viper.AddConfigPath(path)

			if err := viper.ReadInConfig(); err != nil {
				log.Fatal(err)
			}

			log.SetVerbose(viper.GetBool("verbose"))
		},

		Run: func(cmd *cobra.Command, args []string) {
			var (
				done    = make(chan bool)
				apps    = make(chan bool)
				wait    = 0
				daemons = map[string]func() bool{
					"api":       api.Daemon,
					"core":      core.Daemon,
					"discovery": discovery.Daemon,
					"node":      node.Daemon,
				}
			)

			components := []string{"api", "core", "discovery", "node"}

			if len(args) > 0 {
				components = args
			}

			for _, app := range components {
				go func(app string) {
					if _, ok := daemons[app]; ok {
						wait++
						apps <- daemons[app]()
					}
				}(app)
			}

			go func() {
				for {
					select {
					case <-apps:
						wait--
						if wait == 0 {
							done <- true
							return
						}
					}
				}
			}()

			<-done
		},
	}
)

func init() {
	CLI.Flags().StringVarP(&config, "config", "c", "./contrib/config.yml", "/path/to/config.yml")
}

func main() {
	if err := CLI.Execute(); err != nil {
		log.Fatal(err)
	}
}