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
	"github.com/AlecAivazis/survey/v2"
	"github.com/kpture/kpture/pkg/install"
	"github.com/kpture/kpture/pkg/kubernetes"
	"github.com/spf13/cobra"
)

// installCmd represents the install command
var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install kpture tools on your kubernetes cluster",
	Long:  `This perform the installation of a daemonset, a proxy and a nodePort service`,
	Run: func(cmd *cobra.Command, args []string) {

		socketpath := ""
		prompt := &survey.Input{
			Message: "Containerd socket location",
		}
		err := survey.AskOne(prompt, &socketpath)
		cobra.CheckErr(err)

		ctrnamespace := ""
		prompt = &survey.Input{
			Message: "Containerd namespace",
		}
		err = survey.AskOne(prompt, &ctrnamespace)

		cobra.CheckErr(err)

		client, err := kubernetes.LoadClient(Kubeconfig)
		config, err := kubernetes.LoadConfig(Kubeconfig)
		cobra.CheckErr(err)
		// install.InstallDaemonset(client, "kpture", "moby", "/run/containerd/containerd.sock")
		install.InstallDaemonset(client, "kpture", ctrnamespace, socketpath)
		install.InstallProxy(client, "kpture")
		install.Installservice(client, "kpture")
		install.InstallRole("kpture", config)
		install.InstallRoleBinding("kpture", config)
	},
}

func init() {
	rootCmd.AddCommand(installCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// installCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// installCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
