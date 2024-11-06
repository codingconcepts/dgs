package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pkg/profile"
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
	url        string
	debug      bool
	cpuProfile string

	// Gen data flags.
	config     string
	batch      int
	workers    int
	insertMode string

	// Gen config flags.
	schema    string
	rowCounts []string
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
	genCmd.PersistentFlags().StringVar(&cpuProfile, "cpu-profile", "", "run cpu profiling (output to cpuprof)")
	genCmd.MarkPersistentFlagRequired("url")

	genDataCmd := &cobra.Command{
		Use:   "data",
		Short: "Generate relational data",
		Run:   genData,
	}

	genDataCmd.Flags().StringVar(&config, "config", "", "absolute or relative path to the config file")
	genDataCmd.Flags().IntVar(&batch, "batch", 1000, "query and insert batch size")
	genDataCmd.Flags().IntVar(&workers, "workers", 4, "number of workers to run concurrently")
	genDataCmd.Flags().StringVar(&insertMode, "insert-mode", "upsert", "type of insert to run [insert | upsert]")
	genDataCmd.MarkFlagRequired("config")

	genConfigCmd := &cobra.Command{
		Use:   "config",
		Short: "Generate the config file for a given database schema",
		Run:   genConfig,
	}

	genConfigCmd.Flags().StringVar(&schema, "schema", "public", "name of the schema to create a config for")
	genConfigCmd.Flags().StringSliceVar(&rowCounts, "row-count", nil, "row count per table as TABLE_NAME:ROW_COUNT (otherwise 100,000)")
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
	if cpuProfile != "" {
		defer profile.Start(profile.ProfilePath(".")).Stop()
	}

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

	db := mustConnect(url)
	defer db.Close()

	parsedInsertMode := model.ParseInsertMode(insertMode)
	if parsedInsertMode == model.InsertModeInvalid {
		logger.Fatal().Msgf("%s is not a valid insert-mode", insertMode)
	}

	g := commands.NewDataGenerator(db, logger, c, workers, batch, parsedInsertMode)

	logger.Debug().Msg("generating data")
	if err = g.Generate(); err != nil {
		logger.Fatal().Msgf("error generating data: %v", err)
	}

	logger.Info().Msg("done")
}

func genConfig(cmd *cobra.Command, args []string) {
	if cpuProfile != "" {
		defer profile.Start(profile.ProfilePath(".")).Stop()
	}

	db := mustConnect(url)
	defer db.Close()

	config, err := commands.GenerateConfig(db, schema, rowCounts)
	if err != nil {
		logger.Fatal().Msgf("error generating config: %v", err)
	}

	if err = yaml.NewEncoder(os.Stdout).Encode(config); err != nil {
		logger.Fatal().Msgf("error printing config: %v", err)
	}
}

func mustConnect(url string) *pgxpool.Pool {
	cfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		log.Fatalf("error parsing connection string: %v", err)
	}
	cfg.MaxConns = int32(workers)

	db, err := pgxpool.NewWithConfig(context.Background(), cfg)
	if err != nil {
		logger.Fatal().Msgf("error connecting to database: %v", err)
	}

	if err = db.Ping(context.Background()); err != nil {
		log.Fatalf("error pinging database: %v", err)
	}

	return db
}

func showVersion(cmd *cobra.Command, args []string) {
	fmt.Println(version)
}
