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
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcapgo"
	"github.com/kpture/kpture/pkg/kubernetes"
	"github.com/kpture/kpture/pkg/socket"
	"github.com/spf13/cobra"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

var cfgFile string

//Namespace represent the kubernetes Namespace provided in configuration
var Namespace string

//Kubeconfig represent the kubernetes configuration file
var Kubeconfig string

//OutputFolder represent the kubernetes configuration file
var OutputFolder string

//Logs
var Logs bool

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
		client, err := kubernetes.LoadClient(Kubeconfig)
		cobra.CheckErr(err)
		config, err := kubernetes.LoadConfig(Kubeconfig)
		cobra.CheckErr(err)
		pods, dial := kubernetes.SelectPod(client, Namespace, config)

		if len(pods) == 0 {
			return
		}

		err = os.Mkdir(OutputFolder, os.ModePerm)
		if err != nil {
			cobra.CheckErr(err)
		}

		f, err := os.Create(OutputFolder + "/merged.pcap")
		if err != nil {
			cobra.CheckErr(err)
		}
		wf := pcapgo.NewWriter(f)
		wf.WriteFileHeader(1024, layers.LinkTypeEthernet)

		podLogOpts := v1.PodLogOptions{}
		if Logs {
			podLogOpts = v1.PodLogOptions{SinceTime: &metav1.Time{Time: time.Now()}}
			c := make(chan os.Signal)
			signal.Notify(c, os.Interrupt, syscall.SIGTERM)
			go func() {
				<-c
				kubernetes.GetLogs(client, Namespace, pods, podLogOpts, OutputFolder)
				os.Exit(1)
			}()
		}
		for _, pod := range pods {
			err = os.Mkdir(OutputFolder+"/"+pod, os.ModePerm)
			if err != nil {
				cobra.CheckErr(err)
			}
			socket.StartCapture(socket.Capture{ContainerName: pod, ContainerNamespace: Namespace, Interface: "eth0", FileName: OutputFolder + "/" + pod + "/" + pod + ".pcap"}, dial, wf)
		}
		for {

		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {

	t := fmt.Sprint(time.Now().Unix())
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.kpture.yaml)")
	rootCmd.PersistentFlags().StringVarP(&Namespace, "namespace", "n", "default", "kubernetes namespace")

	rootCmd.Flags().BoolVarP(&Logs, "logs", "l", false, "fetch container logs as well")

	home, err := homedir.Dir()
	cobra.CheckErr(err)
	rootCmd.PersistentFlags().StringVarP(&Kubeconfig, "kubeconfig", "k", home+"/.kube/config", "kubernetes configfile")
	rootCmd.Flags().StringVarP(&OutputFolder, "output", "o", t, "Output folder for pcap files")
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
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}
