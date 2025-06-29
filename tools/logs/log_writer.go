package logs

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/fatih/color"
)

// print log to prettify it,
// buf is for the text payload
func PrintLog(log *Log, payload []byte) {

	defaultLevel := color.BgGreen

	switch log.Level {
	case slog.LevelDebug:
		defaultLevel = color.FgHiBlack
	case slog.LevelInfo:
		defaultLevel = color.FgBlue
	case slog.LevelError:
		defaultLevel = color.FgRed
	case slog.LevelWarn:
		defaultLevel = color.FgYellow
	}

	c := color.New(defaultLevel).SprintFunc()
	ct := color.New(color.FgGreen).SprintFunc()

	fmt.Printf("%s %s [%s] msg='%s' || %s  \n", ct(log.Time.Format(time.RFC3339)), c(log.Level), log.Source, log.Message, string(payload))

}
