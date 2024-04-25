package cmd

import (
	"github.com/primait/nuvola/pkg/io/logging"
	"github.com/spf13/cobra"
)

const (
	flagVerbose         = "verbose"
	flagDebug           = "debug"
	flagAWSProfile      = "aws-profile"
	flagAWSEndpointUrl  = "aws-endpoint-url"
	flagOutputDirectory = "output-dir"
	flagOutputFormat    = "output-format"
	flagDumpOnly        = "dump-only"
	flagImportFile      = "import"
	flagNoImport        = "no-import"
)

var (
	logger          logging.LogManager
	awsProfile      string
	awsEndpointUrl  string
	outputDirectory string
	outputFormat    string
	dumpOnly        bool
	importFile      string
	noImport        bool
	rootCmd         = &cobra.Command{
		Use:   "nuvola",
		Short: "A tool to dump and perform automatic and manual security analysis on AWS",
	}
)

func init() {
	logger = logging.GetLogManager()
	rootCmd.PersistentFlags().BoolP(flagVerbose, "v", false, "Verbose output")
	rootCmd.PersistentFlags().BoolP(flagDebug, "d", false, "Debug output")
	dumpCmd.PersistentFlags().StringVarP(&awsProfile, flagAWSProfile, "p", "default", "AWS Profile to use")
	dumpCmd.PersistentFlags().StringVarP(&outputDirectory, flagOutputDirectory, "o", "", "Output folder where the files will be saved (default: \".\")")
	dumpCmd.PersistentFlags().StringVarP(&outputFormat, flagOutputFormat, "f", "zip", "Output format: ZIP or json files")
	dumpCmd.PersistentFlags().BoolVarP(&dumpOnly, flagDumpOnly, "", false, "Flag to prevent loading data into Neo4j (default: \"false\")")

	assessCmd.PersistentFlags().StringVarP(&importFile, flagImportFile, "i", "", "Input ZIP file to load")
	assessCmd.PersistentFlags().BoolVarP(&noImport, flagNoImport, "", false, "Use stored data from Neo4j without import (default)")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		logger.Error("Error executing command", "err", err)
	}
}

func markAsRequired(flag string) {
	if err := rootCmd.MarkFlagRequired(flag); err != nil {
		logger.Error("Required flags not provided", "err", err, "flag", flag)
	}
}
