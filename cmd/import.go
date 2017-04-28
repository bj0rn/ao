// Copyright © 2017 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"

	"github.com/skatteetaten/aoc/pkg/setup"
	"github.com/spf13/cobra"
	"log"
	"os"
)

// importCmd represents the import command
var importCmd = &cobra.Command{
	Use:   "import folder | file",
	Short: "Imports a set of configuration files to the central store.",
	Long: `When used with a .json file as an argument, it will import the application referred to in the
file merged with about.json in the same folder, and about.json and aos-features.json in the parent folder.`,
	Run: func(cmd *cobra.Command, args []string) {
		// TODO: Work your own magic here
		var setupObject setup.SetupClass
		output, err := setupObject.ExecuteSetupImport(args,
			nil, &persistentOptions, localDryRun, false)
		if err != nil {
			l := log.New(os.Stderr, "", 0)
			l.Println(err.Error())
			os.Exit(-1)
		} else {
			if output != "" {
				fmt.Println(output)
			}
		}
	},
}

func init() {
	RootCmd.AddCommand(importCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// importCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// importCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	importCmd.Flags().BoolVarP(&localDryRun, "localdryrun",
		"z", false, "Does not initiate API, just prints collected files")
}