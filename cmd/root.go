package cmd

import (
	"io/ioutil"
	"os"
	"strings"

	"github.com/appuio/seiso/cfg"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var (
	// rootCmd represents the base command when called without any subcommands
	rootCmd = &cobra.Command{
		Use:              "seiso",
		Short:            "Keeps your Kubernetes projects clean",
		PersistentPreRun: parseConfig,
	}
	config = cfg.NewDefaultConfig()
)

// Execute is the main entrypoint of the CLI, it executes child commands as given by the user-defined flags and arguments.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringP("namespace", "n", config.Namespace, "Cluster namespace of current context")
	rootCmd.PersistentFlags().String("log.level", config.Log.LogLevel, "Log level, one of [debug info warn error fatal]")
	rootCmd.PersistentFlags().BoolP("log.verbose", "v", config.Log.Verbose, "Shorthand for --log.level debug")
	rootCmd.PersistentFlags().BoolP("log.batch", "b", config.Log.Batch, "Use Batch mode (disables logging, prints deleted images only)")
	cobra.OnInitialize(initRootConfig)
}

func initRootConfig() {
	bindFlags(rootCmd.Flags())
}

// parseConfig reads the flags and ENV vars
func parseConfig(cmd *cobra.Command, args []string) {
	bindFlags(cmd.PersistentFlags())
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	viper.AutomaticEnv()

	if err := viper.Unmarshal(&config); err != nil {
		log.WithError(err).Fatal("Could not read config")
	}

	log.SetFormatter(&log.TextFormatter{
		DisableTimestamp: true,
	})

	if config.Log.Batch {
		log.SetOutput(ioutil.Discard)
	} else {
		log.SetOutput(os.Stderr)
	}
	if config.Log.Verbose {
		config.Log.LogLevel = "debug"
	}
	level, err := log.ParseLevel(config.Log.LogLevel)
	if err != nil {
		log.WithError(err).Warn("Could not parse log level, fallback to info level")
		log.SetLevel(log.InfoLevel)
	} else {
		log.SetLevel(level)
	}

	if config.Force {
		log.Warn("--force is deprecated and will be removed by June 30, 2020. Please use --delete instead.")
		log.Warn("Continuing with --delete")
		config.Delete = true
	}
}

func bindFlags(flagSet *pflag.FlagSet) {
	if err := viper.BindPFlags(flagSet); err != nil {
		log.WithError(err).Fatal("Could not bind flags")
	}
}

// SetVersion sets the version string in the help messages
func SetVersion(version string) {
	rootCmd.Version = version
}
