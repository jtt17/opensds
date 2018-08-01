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
)
/*
func loadConf() {
	lastpath, _ := os.Getwd()
	os.Chdir("/root/gopath/src/github.com/opensds/opensds/pkg/utils/logs/")
	file, err := os.Open("conf.json")
	defer file.Close()
	os.Chdir(lastpath)
