package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/codingconcepts/dgs/pkg/commands"
	"github.com/codingconcepts/dgs/pkg/model"
)

var (
	logger zerolog.Logger

	// Application version (populated via ldflags).
	version string

	// Shared flags.
	url   string
	debug bool

	// Gen data flags.
	config  string
	batch   int
	workers int

	// Gen config flags.
	schema string
)

func main() {
	rootCmd := &cobra.Command{
		Use:              "dgs",
		Short:            "dgs is a streaming data generator",
		PersistentPreRun: initialize,
	}

	genCmd := &cobra.Command{
		Use:   "gen",
		Short: "Generate things using dgs",
	}

	genCmd.PersistentFlags().StringVar(&url, "url", "", "connection string")
	genCmd.PersistentFlags().BoolVar(&debug, "debug", false, "enhable debug logging")
	genCmd.MarkPersistentFlagRequired("url")

	genDataCmd := &cobra.Command{
		Use:   "data",
		Short: "Generate relational data",
		Run:   genData,
	}

	genDataCmd.Flags().StringVar(&config, "config", "", "absolute or relative path to the config file")
	genDataCmd.Flags().IntVar(&batch, "batch", 10000, "query and insert batch size")
	genDataCmd.Flags().IntVar(&workers, "workers", 4, "number of workers to run concurrently")
	genDataCmd.MarkFlagRequired("config")

	genConfigCmd := &cobra.Command{
		Use:   "config",
		Short: "Generate the config file for a given database schema",
		Run:   genConfig,
	}

	genConfigCmd.Flags().StringVar(&schema, "schema", "public", "name of the schema to create a config for")
	genConfigCmd.MarkFlagRequired("schema")

	genCmd.AddCommand(genDataCmd, genConfigCmd)

	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Show application version",
		Run:   showVersion,
	}

	rootCmd.AddCommand(genCmd, versionCmd)

	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("error running dgs: %v", err)
	}
}

func initialize(cmd *cobra.Command, args []string) {
	logger = zerolog.New(zerolog.ConsoleWriter{
		Out: os.Stderr,
		PartsExclude: []string{
			zerolog.TimestampFieldName,
		},
	}).Level(lo.Ternary(debug, zerolog.DebugLevel, zerolog.InfoLevel))
}

func genData(cmd *cobra.Command, args []string) {
	logger.Debug().Msgf("reading file: %s", config)
	file, err := os.ReadFile(config)
	if err != nil {
		logger.Fatal().Msgf("error reading config file: %v", err)
	}

	logger.Debug().Msgf("parsing file: %s", config)
	c, err := model.ParseConfig(string(file), logger)
	if err != nil {
		logger.Fatal().Msgf("error parsing config file: %v", err)
	}

	db, err := pgxpool.New(context.Background(), url)
	if err != nil {
		logger.Fatal().Msgf("error connecting to database: %v", err)
	}
	defer db.Close()

	g := commands.NewDataGenerator(db, logger, c, workers, batch)

	logger.Debug().Msg("generating data")
	if err = g.Generate(); err != nil {
		logger.Fatal().Msgf("error generating data: %v", err)
	}

	logger.Info().Msg("done")
}

func genConfig(cmd *cobra.Command, args []string) {
	db, err := pgxpool.New(context.Background(), url)
	if err != nil {
		logger.Fatal().Msgf("error connecting to database: %v", err)
	}
	defer db.Close()

	config, err := commands.GenerateConfig(db, schema)
	if err != nil {
		logger.Fatal().Msgf("error generating config: %v", err)
	}

	if err = yaml.NewEncoder(os.Stdout).Encode(config); err != nil {
		logger.Fatal().Msgf("error printing config: %v", err)
	}
}

func showVersion(cmd *cobra.Command, args []string) {
	fmt.Println(version)
}
