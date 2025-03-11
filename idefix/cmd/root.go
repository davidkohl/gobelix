// cmd/root.go
package cmd

import (
	"github.com/spf13/cobra"
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
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Enable verbose output")

	// Version flag
	rootCmd.Flags().BoolP("version", "V", false, "Print version information")
	rootCmd.SetVersionTemplate("Idefix v{{.Version}} - ASTERIX decoder companion\n")
	rootCmd.Version = "0.1.0" // Set your version here
}
