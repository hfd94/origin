package log

import (
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"io"
	"path/filepath"
	"strings"
	"time"
)

const (
	defaultFileExt = "log"
)

type WriterOptions struct {
	Path    string
	nodeId  string
	Level   Level
	MaxAge  time.Duration
	MaxSize int64
	CutRule CutRule
}

func NewWriter(opts WriterOptions) (io.Writer, error) {
	var (
		rotationTime time.Duration
		srcFileParts = make([]string, 0, 3)
		newFileParts = make([]string, 0, 4)
	)

	srcFileParts = append(srcFileParts, opts.nodeId)
	newFileParts = append(newFileParts, opts.nodeId)

	if opts.Level != 0 {
		srcFileParts = append(srcFileParts, strings.ToLower(opts.Level.String()))
		newFileParts = append(newFileParts, strings.ToLower(opts.Level.String()))
	}

	switch opts.CutRule {
	case CutByYear:
		newFileParts = append(newFileParts, "%Y")
		rotationTime = 365 * 24 * time.Hour
	case CutByMonth:
		newFileParts = append(newFileParts, "%Y_%m")
		rotationTime = 31 * 24 * time.Hour
	case CutByDay:
		newFileParts = append(newFileParts, "%Y_%m_%d")
		rotationTime = 24 * time.Hour
	case CutByHour:
		newFileParts = append(newFileParts, "%Y_%m_%d_%H")
		rotationTime = time.Hour
	case CutByMinute:
		newFileParts = append(newFileParts, "%Y_%m_%d_%H_%M")
		rotationTime = time.Minute
	case CutBySecond:
		newFileParts = append(newFileParts, "%Y_%m_%d_%H_%M_%S")
		rotationTime = time.Second
	}
	srcFileParts = append(srcFileParts, defaultFileExt)
	newFileParts = append(newFileParts, defaultFileExt)

	srcFileName := filepath.Join(opts.Path, strings.Join(srcFileParts, "."))
	newFileName := filepath.Join(opts.Path, strings.Join(newFileParts, "."))

	options := make([]rotatelogs.Option, 0, 4)
	options = append(options, rotatelogs.WithLinkName(srcFileName))
	if opts.MaxAge > 0 {
		options = append(options, rotatelogs.WithMaxAge(opts.MaxAge))
	}
	if opts.MaxSize > 0 {
		options = append(options, rotatelogs.WithRotationSize(opts.MaxSize))
	}
	if rotationTime > 0 {
		options = append(options, rotatelogs.WithRotationTime(rotationTime))
	}

	return rotatelogs.New(newFileName, options...)
}
