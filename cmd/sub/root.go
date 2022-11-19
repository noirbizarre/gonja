/*
Copyright © 2022 Christian Kniep <christian@qnib.org>
*/
package sub

import (
	"os"
	"strings"

	"github.com/noirbizarre/gonja"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var (
	rootCmd = &cobra.Command{
		Use:   "gonja-cli <template> <dst>",
		Short: "Command-line tool to render gonja templates",
		Long:  `Command-line tool to render gonja templates`,
		Run:   rCmd,
	}
)

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {

}

func rCmd(cmd *cobra.Command, args []string) {
	if len(args) < 2 {
		cmd.Help()
		os.Exit(1)
	}
	tpl, err := gonja.FromFile(args[0])
	if err != nil {
		panic(err)
	}
	// Now you can render the template with the given
	// gonja.Context how often you want to.
	ctx := gonja.Context{}
	for _, element := range os.Environ() {
		variable := strings.Split(element, "=")
		ctx[variable[0]] = variable[1]
	}
	out, err := tpl.Execute(ctx)
	if err != nil {
		panic(err)
	}
	dst := args[1]
	if dst == "-" {
		os.Stdout.WriteString(out)
	} else {
		f, err := os.Create(dst)
		if err != nil {
			panic(err)
		}
		defer f.Close()
		f.WriteString(out + "\n")
	}
}
