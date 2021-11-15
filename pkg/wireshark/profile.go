package wireshark

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"regexp"
	"time"

	nested "github.com/antonfisher/nested-logrus-formatter"
	"github.com/kpture/kpture/pkg/utils"

	log "github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
)

type Wireshark struct {
	ExecutablePath   string   `mapstructure:"ExecutablePath"`
	ProfilePath      string   `mapstructure:"ProfilePath"`
	ProfileName      string   `mapstructure:"ProfileName"`
	OtherArgs        []string `mapstructure:"OtherArgs"`
	AdditionnalHosts []string `mapstructure:"AdditionnalHosts"`
	Logger           *log.Entry
}

const (
	defautName = "k8s"
)

var semverRegex = `(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(?:-((?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\+([0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?`

func (w *Wireshark) Init(hostFile []byte, loglevel log.Level) error {

	WiresharkLogger := log.New()
	WiresharkLogger.SetLevel(loglevel)
	WiresharkLogger.SetFormatter(&nested.Formatter{
		HideKeys:        true,
		TrimMessages:    true,
		TimestampFormat: time.Stamp,
	})

	w.Logger = WiresharkLogger.WithFields(log.Fields{
		"": "wireshark",
	})

	if w.ProfileName == "" {
		w.ProfileName = defautName
	}

	if err := w.CheckProfilePath(); err != nil {
		return err
	}
	if err := w.CheckExecutablePath(); err != nil {
		return err
	}
	if err := w.CheckVersion(); err != nil {
		return err
	}
	w.CreateProfile(hostFile)

	// 	err = wireshark.Run()
	// 	kubernetes.GetLogs(client, Namespace, pods, podLogOpts, OutputFolder)
	// 	os.Exit(0)

	return nil
}

func (w *Wireshark) CheckVersion() error {
	w.Logger.WithFields(log.Fields{"executablepath": path.Base(w.ExecutablePath) + " --version"}).Trace("Checking Version")
	WirsharkVersion := exec.Command(w.ExecutablePath, "--version")
	var outb bytes.Buffer
	WirsharkVersion.Stdout = &outb

	if err := WirsharkVersion.Run(); err != nil {
		fmt.Println(err)
		if exitError, ok := err.(*exec.ExitError); ok {
			log.Debug(exitError.ExitCode())
			return err
		}
	}

	re := regexp.MustCompile(semverRegex)
	find := re.Find(outb.Bytes())

	if find == nil {
		return errors.New("could not get version")
	}
	w.Logger.Info("Detected version ", string(find))

	return nil
}

func (w *Wireshark) CheckProfilePath() error {
	w.Logger.WithFields(log.Fields{"ProfilePath": w.ProfilePath}).Trace("Checking Profile Path")

	pp := (path.Dir(w.ProfilePath + "/"))
	if pp == "." {
		return errors.New("ProfilePath is not valid")
	}
	w.ProfilePath = pp
	return nil
}

func (w *Wireshark) CheckExecutablePath() error {
	w.Logger.WithFields(log.Fields{"ExecutablePath": w.ExecutablePath}).Trace("Checking Executable Path")
	if _, err := os.Stat(w.ExecutablePath); errors.Is(err, os.ErrNotExist) {
		return errors.New("ExecutablePath is not valid")
	}
	return nil
}

func (w *Wireshark) CreateProfile(DnsConfig []byte) {
	w.Logger.WithFields(log.Fields{"Profile": path.Join(w.ProfilePath, w.ProfileName)}).Trace("Create Profile")
	err := os.Mkdir(path.Join(w.ProfilePath, w.ProfileName), 0755)
	if err != nil {
		if errors.Is(err, os.ErrExist) {
			w.Logger.Info(err)
		} else {
			w.Logger.Error(err)
		}
	}
	w.createHosts(DnsConfig)
	w.createPreferences()
}

//Create host configuration file
func (w *Wireshark) createHosts(DnsConfig []byte) {
	w.Logger.Trace("Create Dns hosts")
	f, err := os.Create(path.Join(w.ProfilePath, w.ProfileName) + "/hosts")
	if err != nil {
		if errors.Is(err, os.ErrExist) {
			w.Logger.Info(err)
		} else {
			w.Logger.Error(err)
		}
	}
	f.Write(DnsConfig)
}

//Create preferences files to use host dns
func (w *Wireshark) createPreferences() {
	w.Logger.Trace("Creating preferences file...")
	f, err := os.Create(path.Join(w.ProfilePath, w.ProfileName) + "/preferences")
	if err != nil {
		fmt.Println(err)
		cobra.CheckErr(err)
	}
	f.Write([]byte(conf))
}

func (w *Wireshark) Open(port int) {
	ip := utils.GetOutboundIP()
	go func() {
		w.Logger.Info("Opening Wireshark")
		wireshark := exec.Command(w.ExecutablePath, "-k", "-i", "TCP@"+ip.String()+":0"+fmt.Sprint(port), "-C", w.ProfileName)
		wireshark.Start()
	}()
}
