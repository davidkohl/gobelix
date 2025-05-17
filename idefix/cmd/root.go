// cmd/root.go
package cmd

import (
	"github.com/spf13/cobra"
)

// Global flags
var (
	Verbose  bool
	JsonLogs bool
)

var rootCmd = &cobra.Command{
	Use:   "idefix",
	Short: "ASTERIX message decoder and analyzer",
	Long: `
 ______        __             ______   __           
/      |      /  |           /      \ /  |          
$$$$$$/   ____$$ |  ______  /$$$$$$  |$$/  __    __ 
  $$ |   /    $$ | /      \ $$ |_ $$/ /  |/  \  /  |
  $$ |  /$$$$$$$ |/$$$$$$  |$$   |    $$ |$$  \/$$/ 
  $$ |  $$ |  $$ |$$    $$ |$$$$/     $$ | $$  $$<  
 _$$ |_ $$ \__$$ |$$$$$$$$/ $$ |      $$ | /$$$$  \ 
/ $$   |$$    $$ |$$       |$$ |      $$ |/$$/ $$  |
$$$$$$/  $$$$$$$/  $$$$$$$/ $$/       $$/ $$/   $$/ 
                                                                                                      	
Idefix is a CLI utility for capturing, decoding, and analyzing 
ASTERIX data from network traffic or files. It works with the Gobelix ASTERIX 
decoding library by David Kohl to provide human-readable output of radar data.
https://github.com/davidkohl/gobelix
`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().BoolVarP(&Verbose, "verbose", "v", false, "Enable verbose output")
	rootCmd.PersistentFlags().BoolVar(&JsonLogs, "json", false, "Log in JSON format")

	// Version flag
	rootCmd.Flags().BoolP("version", "V", false, "Print version information")
	rootCmd.SetVersionTemplate("Idefix v{{.Version}} - ASTERIX decoder companion\n")
	rootCmd.Version = "0.2.0" // Updated version number
}
