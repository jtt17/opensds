package logs

import (
//	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
	"sync"
)

type configuration struct {
	Path          string `json:"Path"`
	MaxSize       uint64 `json:"MaxSize"`
	Level         int32  `json:"Level"`
	Ldate         bool   `json:"Ldate"`
	Ltime         bool   `json:"Ltime"`
	Lmicroseconds bool   `json:"Lmicroseconds"`
	LUTC          bool   `json:"LUTC"`
	LogToFile     bool   `json:"LogToFile"`
	LogToStdErr   bool   `json:"LogToStdErr"`
	Total         int64
}

const (
	info int32 = iota
	warn
	erro
	fata
)
var (
	program = filepath.Base(os.Args[0]) // program name
	mu		 sync.Mutex
	conf     configuration
	flog     *log.Logger
	now      time.Time
	curFile  *os.File
	fileinfo *os.FileInfo
	logmsg   []string
	lastTime time.Time
)
func loadConf() {
	conf.Path = filepath.Join("/var/log/opensds", strings.Split(program, ".")[0])
	conf.Level = -1
	conf.Ldate = true
	conf.Lmicroseconds = true
	conf.LogToFile = true
	conf.LogToStdErr = true
	conf.MaxSize = 1024*1024 *1
}

func exits(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}
func mkdir() {
	if exits(conf.Path) {
		return
	}
	e := os.MkdirAll(conf.Path, 0700)
	if e != nil {
		Fatal(e)
	}
}
func init() {
	loadConf()
	mkdir()
	var tmp int = 0
	if conf.Ldate {
		tmp |= log.Ldate
	}
	if conf.Ltime {
		tmp |= log.Ltime
	}
	if conf.Lmicroseconds {
		tmp |= log.Lmicroseconds
	}
	if conf.LUTC {
		tmp |= log.LUTC
	}
	log.SetFlags(tmp)
	curFile, fileinfo = initFile(conf.Path)
	logmsg = []string{}
	lastTime = time.Now()
}

func initFile(path string) (*os.File, *os.FileInfo) {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		Error(err)
		return nil, nil
	}
	l := len(files)
	fmt.Println("Got Log File List :",files)
	if l == 0 {
		return nil, nil
	}
	if uint64(files[l-1].Size()) >= conf.MaxSize {
		return nil, nil
	}
	tmpfile, e1 := os.OpenFile(filepath.Join(conf.Path, files[l-1].Name()), os.O_WRONLY|os.O_APPEND, 0666)
	if e1 != nil {
		return nil, nil
	}
	tmpinfo, e2 := os.Stat(tmpfile.Name())
	if e2 != nil {
		return nil,nil
	}
	fmt.Println("INIT: Now should USE FILE: ",tmpfile.Name())
	return tmpfile, &tmpinfo
}

func open()(*os.File,*os.FileInfo) {
	if fileinfo == nil {
		return nil,nil
	}
	file, err := os.OpenFile(filepath.Join(conf.Path, (*fileinfo).Name()), os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, nil
	}
	tmpinfo, _ := os.Stat(file.Name())
	return file, &tmpinfo
}
func create()(*os.File, *os.FileInfo)  {
	name := fmt.Sprintf("%s %04d-%02d-%02d %02d.%02d.%02d.log",
		filepath.Join(conf.Path, program),
		time.Now().Year(),
		time.Now().Month(),
		time.Now().Day(),
		time.Now().Hour(),
		time.Now().Minute(),
		time.Now().Second())
	file, err := os.OpenFile(name, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, nil
	}
	file.WriteString(fmt.Sprintf("Log file Create at: %v\n\n", time.Now().Format("2018-01-02 15:04:05.000000")))
	tmpinfo ,e := os.Stat(name)
	if e != nil {
		fmt.Println("get fileinfo failed ",e)
		return nil, nil
	}
	return file, &tmpinfo
}

//
//
func doPrint(s string) {
	logmsg = append(logmsg, s)
	if time.Now().Sub(lastTime).Seconds() < 30 { //seconds
		return
	}
	curFile, fileinfo = open()
	if fileinfo == nil || (uint64)((*fileinfo).Size()) >= conf.MaxSize {
		curFile, fileinfo = create()
	}
	flog = log.New(curFile, "", log.Flags())

	for _, x := range logmsg {
		flog.Println(x)
	}
	curFile.Close()
	logmsg = []string{}
	lastTime = time.Now()
}
func doInfo(v string) {
	if info < conf.Level {
		return
	}
	if conf.LogToStdErr {
		log.Println(v)
	}
	if conf.LogToFile {
		doPrint(v)
	}
}
func Info(v ...interface{}) {
	_, file, link, _ := runtime.Caller(1)
	s := fmt.Sprint("[Info]: ", file[strings.LastIndex(file,"opensds"):], " ", link, v)
	doInfo(s)
}
func Infof(format string, v ...interface{}) {
	_, file, link, _ := runtime.Caller(1)
	s := fmt.Sprint("[Info]: ", file[strings.LastIndex(file,"opensds"):], " ", link, v)
	doInfo(s)
}
func Infoln(v ...interface{}) {
	_, file, link, _ := runtime.Caller(1)
	s := fmt.Sprintln("[Info]:", file[strings.LastIndex(file,"opensds"):], link, v)
	doInfo(s)
}
func doWarn(v string) {
	if warn < conf.Level {
		return
	}
	if conf.LogToStdErr {
		log.Println(v)
	}
	if conf.LogToFile {
		doPrint(v)
	}
}

func Warning(v ...interface{}) {
	_, file, link, _ := runtime.Caller(1)
	s := fmt.Sprint("[Warn]: ", file[strings.LastIndex(file,"opensds"):], " ", link, v)
	doWarn(s)
}
func Warningf(format string, v ...interface{}) {
	_, file, link, _ := runtime.Caller(1)
	s := fmt.Sprint("[Warn]: ", file[strings.LastIndex(file,"opensds"):], " ", link, v)
	doWarn(s)
}
func Warningln(v ...interface{}) {
	_, file, link, _ := runtime.Caller(1)
	s := fmt.Sprintln("[Warn]:", file[strings.LastIndex(file,"opensds"):], link, v)
	doWarn(s)
}
func doError(s string) {
	if erro < conf.Level {
		return
	}
	if conf.LogToStdErr {
		log.Println(s)
	}
	if conf.LogToFile {
		doPrint(s)
	}
}

func Error(v ...interface{}) {
	_, file, link, _ := runtime.Caller(1)
	s := fmt.Sprint("[Erro]: ", file[strings.LastIndex(file,"opensds"):], " ", link, v)
	doError(s)
}
func Errorf(format string, v ...interface{}) {
	_, file, link, _ := runtime.Caller(1)
	s := fmt.Sprint("[Erro]: ", file[strings.LastIndex(file,"opensds"):], " ", link, v)
	doError(s)
}
func Errorln(v ...interface{}) {
	_, file, link, _ := runtime.Caller(1)
	s := fmt.Sprintln("[Erro]:", file[strings.LastIndex(file,"opensds"):], link, v)
	doError(s)
}
func doFatal(s string) {
	if fata < conf.Level {
		return
	}
	if conf.LogToStdErr {
		log.Println(s)
	}
	if conf.LogToFile {
		doPrint(s)
	}
	if curFile != nil {
		curFile.Close()
	}
	os.Exit(1)
}
func Fatal(v ...interface{}) {
	_, file, link, _ := runtime.Caller(1)
	s := fmt.Sprint("[Fata]: ", file[strings.LastIndex(file,"opensds"):], " ", link, v)
	doFatal(s)
}
func Fatalf(format string, v ...interface{}) {
	_, file, link, _ := runtime.Caller(1)
	s := fmt.Sprint("[Fata]: ", file[strings.LastIndex(file,"opensds"):], " ", link, v)
	doFatal(s)
}
func Fatalln(v ...interface{}) {
	_, file, link, _ := runtime.Caller(1)
	s := fmt.Sprintln("[Fata]:", file[strings.LastIndex(file,"opensds"):], link, v)
	doFatal(s)
}

func FlushLogs() {
	curFile, fileinfo = open()
	if fileinfo == nil || (uint64)((*fileinfo).Size()) >= conf.MaxSize {
		curFile, fileinfo = create()
	}
	flog = log.New(curFile, "", log.Flags())

	for _, x := range logmsg {
		flog.Println(x)
	}
	curFile.Close()
	logmsg = []string{}
	lastTime = time.Now()
}
func InitLogs(){
}
type Verbose bool

func V(level int32) Verbose {
	if level >= conf.Level {
		return Verbose(true)
	}
	return Verbose(false)
}
func (v Verbose) Info(args ...interface{}) {
	if v {
		_, file, link, _ := runtime.Caller(1)
		s := fmt.Sprint("[Info]: ", file[strings.LastIndex(file,"opensds"):], " ", link, args)
		doInfo(s)
	}
}
func (v Verbose) Infof(format string, args ...interface{}) {
	if v {
		_, file, link, _ := runtime.Caller(1)
		s := fmt.Sprint("[Info]: ", file[strings.LastIndex(file,"opensds"):], " ", link, args)
		doInfo(s)
	}
}
func (v Verbose) Infoln(args ...interface{}) {
	if v {
		_, file, link, _ := runtime.Caller(1)
		s := fmt.Sprintln("[Info]: ", file[strings.LastIndex(file,"opensds"):], " ", link, args)
		doInfo(s)
	}
}

