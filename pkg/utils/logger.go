package utils

// func SetupFolders(outputfolder string) error {
// 	err := os.Mkdir(outputfolder, 0755)
// 	if err != nil {
// 		if errors.Is(err, os.ErrExist) {
// 			w.Logger.Info(err)
// 		} else {
// 			w.Logger.Error(err)
// 		}
// 	}
// }

import (
	"time"

	nested "github.com/antonfisher/nested-logrus-formatter"

	log "github.com/sirupsen/logrus"
)

func NewLogger(field string, level log.Level) *log.Entry {
	l := log.New()
	l.SetLevel(level)
	l.SetFormatter(&nested.Formatter{
		HideKeys:        true,
		TrimMessages:    true,
		TimestampFormat: time.Stamp,
	})

	return l.WithFields(log.Fields{"": field})

}
