package log

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/labstack/gommon/color"
	"github.com/mattn/go-colorable"
	"github.com/mattn/go-isatty"
	"github.com/valyala/fasttemplate"
)

type (
	Logger struct {
		prefix     string
		level      int
		output     io.Writer
		template   *fasttemplate.Template
		levels     []string
		color      *color.Color
		filename   string // filename
		backups    int    // max backup
		size       int    // current size
		maxsize    int    // maxsize per file
		bufferPool sync.Pool
		mutex      sync.Mutex
	}
)

const (
	DEBUG = iota
	INFO
	WARN
	ERROR
	FATAL
	OFF
)

var (
	global    = New("", INFO, 0, 0)
	timeLocal = "2006-01-02 15:04:05.999"
	//defaultFormat = "time=${time_rfc3339}, level=${level}, prefix=${prefix}, file=${short_file}, " +
	//	"line=${line}, message=${message}\n"
	defaultFormat = "${prefix}${time_local} ${level}:${pid}:${mid_file}:${line}: ${message}\n"
	pid           = ""
	megabyte      = 1024 * 1024
)

func init() {
	pid = strconv.Itoa(os.Getpid())
}

func New(filename string, level, maxsize, backups int) (l *Logger) {
	l = &Logger{
		level:    level,
		prefix:   "",
		filename: filename,
		maxsize:  maxsize * megabyte,
		backups:  backups,
		template: l.newTemplate(defaultFormat),
		color:    color.New(),
		bufferPool: sync.Pool{
			New: func() interface{} {
				return bytes.NewBuffer(make([]byte, 256))
			},
		},
	}
	l.initLevels()
	l.DisableColor()
	if l.filename != "" {
		l.open()
	} else {
		l.SetOutput(colorable.NewColorableStdout())
	}
	return
}

func SetLogger(l *Logger) {
	global = l
}

func (l *Logger) open() {
	f, err := os.OpenFile(l.filename, os.O_APPEND|os.O_WRONLY|os.O_CREATE, os.ModePerm)
	if err != nil {
		l.Error(err)
		return
	}
	fi, err := os.Stat(l.filename)
	if err != nil {
		l.Error(err)
		return
	}
	l.size = int(fi.Size())
	l.SetOutput(f)
}

func (l *Logger) initLevels() {
	l.levels = []string{
		l.color.Blue("DEBUG"),
		l.color.Green("INFO"),
		l.color.Yellow("WARN"),
		l.color.Red("ERROR"),
		l.color.RedBg("FATAL"),
	}
}

func (l *Logger) newTemplate(format string) *fasttemplate.Template {
	return fasttemplate.New(format, "${", "}")
}

func (l *Logger) DisableColor() {
	l.color.Disable()
	l.initLevels()
}

func (l *Logger) EnableColor() {
	l.color.Enable()
	l.initLevels()
}

func (l *Logger) Prefix() string {
	return l.prefix
}

func (l *Logger) SetPrefix(p string) {
	l.prefix = p
}

func (l *Logger) Level() int {
	return l.level
}

func (l *Logger) SetLevel(v int) {
	l.level = v
}

func (l *Logger) Output() io.Writer {
	return l.output
}

func (l *Logger) SetFormat(f string) {
	l.template = l.newTemplate(f)
}

func (l *Logger) SetOutput(w io.Writer) {
	l.output = w
	if w, ok := w.(*os.File); !ok || !isatty.IsTerminal(w.Fd()) {
		l.DisableColor()
	}
}

func (l *Logger) Print(i ...interface{}) {
	fmt.Fprintln(l.output, i...)
}

func (l *Logger) Printf(format string, args ...interface{}) {
	f := fmt.Sprintf("%s\n", format)
	fmt.Fprintf(l.output, f, args...)
}

func (l *Logger) Debug(i ...interface{}) {
	l.log(DEBUG, "", i...)
}

func (l *Logger) Debugf(format string, args ...interface{}) {
	l.log(DEBUG, format, args...)
}

func (l *Logger) Info(i ...interface{}) {
	l.log(INFO, "", i...)
}

func (l *Logger) Infof(format string, args ...interface{}) {
	l.log(INFO, format, args...)
}

func (l *Logger) Warn(i ...interface{}) {
	l.log(WARN, "", i...)
}

func (l *Logger) Warnf(format string, args ...interface{}) {
	l.log(WARN, format, args...)
}

func (l *Logger) Error(i ...interface{}) {
	l.log(ERROR, "", i...)
}

func (l *Logger) Errorf(format string, args ...interface{}) {
	l.log(ERROR, format, args...)
}

func (l *Logger) Fatal(i ...interface{}) {
	l.log(FATAL, "", i...)
	os.Exit(1)
}

func (l *Logger) Fatalf(format string, args ...interface{}) {
	l.log(FATAL, format, args...)
	os.Exit(1)
}

func DisableColor() {
	global.DisableColor()
}

func EnableColor() {
	global.EnableColor()
}

func Prefix() string {
	return global.Prefix()
}

func SetPrefix(p string) {
	global.SetPrefix(p)
}

func Level() int {
	return global.Level()
}

func SetLevel(v int) {
	global.SetLevel(v)
}

func Output() io.Writer {
	return global.Output()
}

func SetOutput(w io.Writer) {
	global.SetOutput(w)
}

func SetFormat(f string) {
	global.SetFormat(f)
}

func Print(i ...interface{}) {
	global.Print(i...)
}

func Printf(format string, args ...interface{}) {
	global.Printf(format, args...)
}

func Debug(i ...interface{}) {
	global.Debug(i...)
}

func Debugf(format string, args ...interface{}) {
	global.Debugf(format, args...)
}

func Info(i ...interface{}) {
	global.Info(i...)
}

func Infof(format string, args ...interface{}) {
	global.Infof(format, args...)
}

func Warn(i ...interface{}) {
	global.Warn(i...)
}

func Warnf(format string, args ...interface{}) {
	global.Warnf(format, args...)
}

func Error(i ...interface{}) {
	global.Error(i...)
}

func Errorf(format string, args ...interface{}) {
	global.Errorf(format, args...)
}

func Fatal(i ...interface{}) {
	global.Fatal(i...)
}

func Fatalf(format string, args ...interface{}) {
	global.Fatalf(format, args...)
}

func (l *Logger) log(v int, format string, args ...interface{}) {
	if v < l.level {
		return
	}

	l.mutex.Lock()
	defer l.mutex.Unlock()
	buf := l.bufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer l.bufferPool.Put(buf)
	_, file, line, _ := runtime.Caller(3)

	message := ""
	if format == "" {
		message = fmt.Sprint(args...)
	} else {
		message = fmt.Sprintf(format, args...)
	}
	if v == FATAL {
		stack := make([]byte, 4<<10)
		length := runtime.Stack(stack, true)
		message = message + "\n" + string(stack[:length])
	}

	_, err := l.template.ExecuteFunc(buf, func(w io.Writer, tag string) (int, error) {
		switch tag {
		case "time_local":
			return w.Write([]byte(time.Now().Format(timeLocal)))
		case "time_rfc3339":
			return w.Write([]byte(time.Now().Format(time.RFC3339)))
		case "level":
			return w.Write([]byte(l.levels[v]))
		case "pid":
			return w.Write([]byte(pid))
		case "prefix":
			return w.Write([]byte(l.prefix))
		case "long_file":
			return w.Write([]byte(file))
		case "short_file":
			return w.Write([]byte(path.Base(file)))
		case "mid_file":
			return w.Write([]byte(filepath.Base(filepath.Dir(file)) + "/" + filepath.Base(file)))
		case "line":
			return w.Write([]byte(strconv.Itoa(line)))
		case "message":
			return w.Write([]byte(message))
		default:
			return w.Write([]byte(fmt.Sprintf("[unknown tag %s]", tag)))
		}
	})

	if err != nil {
		return
	}
	l.output.Write(buf.Bytes())
	if l.filename != "" {
		l.size += len(buf.Bytes())
		if l.size >= l.maxsize {
			l.rotate()
		}
	}
}

func (l *Logger) rotate() {
	backupFile := fmt.Sprintf("%s.tmp", l.filename)
	os.Remove(backupFile)
	if err := os.Rename(l.filename, backupFile); err != nil {
		l.Error(err)
		return
	}

	l.open()

	go func() {
		dir := filepath.Dir(l.filename)
		base := filepath.Base(l.filename)
		list, err := ioutil.ReadDir(dir)
		if err != nil {
			l.Error(err)
			return
		}

		var archives []int
		for _, file := range list {
			if file.IsDir() || !strings.HasPrefix(file.Name(), base) {
				continue
			}

			idxStr := strings.TrimPrefix(file.Name(), base+".")
			idx, _ := strconv.Atoi(idxStr)
			if idx != 0 {
				archives = append(archives, idx)
			}
		}

		sort.Sort(sort.Reverse(sort.IntSlice(archives)))
		for _, i := range archives {
			filename := fmt.Sprintf("%s.%d", l.filename, i)
			if i+1 >= l.backups {
				os.Remove(filename)
				continue
			}

			newFile := fmt.Sprintf("%s.%d", l.filename, i+1)
			os.Rename(filename, newFile)
		}

		newFile := fmt.Sprintf("%s.%d", l.filename, 1)
		os.Rename(backupFile, newFile)
	}()
}
