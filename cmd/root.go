/*
Copyright © 2020 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"log"

	"github.com/Jeffail/tunny"
	"github.com/lampnick/doctron/worker"

	"github.com/lampnick/doctron/app"
	"github.com/lampnick/doctron/conf"
	"github.com/spf13/cobra"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "doctron",
	Short: "doctron is use for document convert.",
	Long:  `doctron is use for document convert. It's support html convert to pdf, html convert to image(such as png,jpeg), add watermark on pdf document, pdf document convert to images'`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		initDoctronWorker()
		doctron := app.NewDoctron()
		err := doctron.Listen(conf.LoadedConfig.Doctron.Domain)
		if err != nil {
			doctron.Logger().Fatal("start doctron failed. %v", err)
		}
	},
}

func initDoctronWorker() {
	worker.Pool = tunny.NewFunc(conf.LoadedConfig.Doctron.MaxConvertWorker, worker.DoctronHandler)
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("[root Execute] err: %v", err)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.doctron.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			log.Fatalf("[initConfig homedir] err: %v", err)
		}

		// Search config in home directory with name "default" (without extension).
		viper.AddConfigPath(home + "/go/Example/doctron/conf")
		viper.SetConfigName("default")

	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	conf.LoadedConfig = conf.NewConfig()
	if err := viper.ReadInConfig(); err == nil {
		log.Println("Using config file:", viper.ConfigFileUsed())
		err := viper.UnmarshalKey("doctron", &conf.LoadedConfig.Doctron)
		if err != nil {
			log.Fatalf("[read config UnmarshalKey doctron] err: %v", err)
		}
		err = viper.UnmarshalKey("oss", &conf.LoadedConfig.Oss)
		if err != nil {
			log.Fatalf("[read config UnmarshalKey oss] err: %v", err)
		}
	} else {
		log.Fatalf("[read config ReadInConfig] err: %v", err)
	}
	log.Printf("[loaded config] \r\n%s\r\n", conf.LoadedConfig)
	initOssConfig(conf.LoadedConfig)
}

func initOssConfig(config *conf.Config) {
	conf.OssConfig.Endpoint = config.Oss.Endpoint
	conf.OssConfig.AccessKeyId = config.Oss.AccessKeyId
	conf.OssConfig.AccessKeySecret = config.Oss.AccessKeySecret
	conf.OssConfig.BucketName = config.Oss.BucketName
	conf.OssConfig.PrivateServerDomain = config.Oss.PrivateServerDomain
}
