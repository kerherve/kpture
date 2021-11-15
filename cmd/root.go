/*
Copyright © 2021 Stephane Guillemot <kpture.git@gmail.com>
All rights reserved.

Redistribution and use in source and binary forms, with or without
modification, are permitted provided that the following conditions are met:

1. Redistributions of source code must retain the above copyright notice,
   this list of conditions and the following disclaimer.

2. Redistributions in binary form must reproduce the above copyright notice,
   this list of conditions and the following disclaimer in the documentation
   and/or other materials provided with the distribution.

3. Neither the name of the copyright holder nor the names of its contributors
   may be used to endorse or promote products derived from this software
   without specific prior written permission.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE
ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE
LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR
CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF
SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS
INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN
CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE)
ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE
POSSIBILITY OF SUCH DAMAGE.
*/
package cmd

import (
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"

	nested "github.com/antonfisher/nested-logrus-formatter"
	"github.com/kpture/kpture/pkg/config"
	"github.com/kpture/kpture/pkg/kpture"
	"github.com/kpture/kpture/pkg/tcpserver"
	"github.com/kpture/kpture/pkg/utils"
	"github.com/spf13/cobra"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

var (
	cfgFile      string
	Namespaces   []string
	Kubeconfig   string
	OutputFolder string
	Config       config.AppConfig
	Logger       *log.Entry
	TcpServer    bool
	Wireshark    bool
	HttpServer   bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "kpture",
	Short: "Packet capture for kubernetes",
	Long: `Kpture is a packet capture tool for kubernetes, it consist of a capture pod on each nodes (daemonset)
which samples packets on desired pods and send the captured informations back via TCP socket.
	`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		utils.PrintAppName()

		k := kpture.KptureCli{}
		t := &tcpserver.TcpServer{}

		//Init the tcpserver only if needed
		if Wireshark || TcpServer {
			t = tcpserver.NewTcpServer(log.TraceLevel)
		}

		if Wireshark {
			KubernetesHostFile, err := k.GetDnsHostFile()
			if err != nil {
				Logger.Fatal("Could not Create DNS File", err)
			}
			err = Config.Wireshark.Init(KubernetesHostFile, log.TraceLevel)
			if err != nil {
				Logger.Fatal(err)
			}
			//Start the capture
			Config.Wireshark.Open(t.Port)
		}

		//Setup And Start Capture
		k.Setup(Kubeconfig, OutputFolder, Namespaces, log.TraceLevel, t.Receiver)
		k.Start(HttpServer)

	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {

	t := fmt.Sprint(time.Now().Unix())

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.kpture.yaml)")
	rootCmd.PersistentFlags().StringArrayVarP(&Namespaces, "namespace", "n", []string{"default"}, "kubernetes namespace")

	home, err := homedir.Dir()
	cobra.CheckErr(err)
	rootCmd.PersistentFlags().BoolVarP(&TcpServer, "tcpserver", "t", false, "Create a local tcp server where wireshark/tshark can read from")
	rootCmd.PersistentFlags().BoolVarP(&Wireshark, "wireshark", "w", false, "Generate wireshark profile and run it on the local tcp server")
	rootCmd.PersistentFlags().BoolVarP(&HttpServer, "metricserver", "m", false, "Enalbe capture stats on :8090/metrics")

	rootCmd.PersistentFlags().StringVarP(&Kubeconfig, "kubeconfig", "k", home+"/.kube/config", "kubernetes configfile")
	rootCmd.Flags().StringVarP(&OutputFolder, "output", "o", t, "Output folder for pcap files")
	cobra.OnInitialize(initConfig)

}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".kpture" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".kpture")
		viper.SetConfigType("yaml") // REQUIRED if the config file does not have the extension in the name

		if err := viper.ReadInConfig(); err != nil {
			fmt.Println(err)
			return
		}

		if err := viper.Unmarshal(&Config); err != nil {
			fmt.Println(err)
			return
		}

	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Errorf("Using config file:", viper.ConfigFileUsed())
	}

	// viper.WriteConfig()

	l := log.New()
	l.SetLevel(log.TraceLevel)
	l.SetFormatter(&nested.Formatter{
		HideKeys:        true,
		TrimMessages:    true,
		TimestampFormat: time.Stamp,
	})

	Logger = l.WithFields(log.Fields{"": "cli"})
}
