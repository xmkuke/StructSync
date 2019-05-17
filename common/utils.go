package common

import (
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"struct_sync/logger"
)

// Check file or path is exists
func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// Get config file path
func GetConfigFile(filePath string, fileName string) string {
	file, _ := exec.LookPath(filePath)
	dir, _ := path.Split(strings.Replace(file, "\\", "/", -1))

	return dir + "conf/" + fileName
}

// Get log file path
func GetLogPath() string {
	file, _ := exec.LookPath(os.Args[0])
	path, _ := filepath.Abs(file)
	splitstring := strings.Split(path, "\\")
	size := len(splitstring)
	splitstring = strings.Split(path, splitstring[size-1])
	ret := strings.Replace(splitstring[0], "\\", "/", size-1) + "log"

	return ret
}

// Init the log config
func InitLogConfig(level int, logFile string, logPath string) {
	logger.SetConsole(true)

	lpath := logPath
	if logPath == "" {
		lpath = GetLogPath()
	}
	pathInfo, err := os.Stat(lpath)
	if nil != err || !pathInfo.IsDir() {
		os.Mkdir(lpath, os.ModePerm)
	}

	logger.SetRollingDaily(lpath, logFile)
	logger.SetLevel(level)
}

// Remove 'auto inc' word
func RemoveAutoIncrement(schema string) string {
	reg := regexp.MustCompile("AUTO_INCREMENT=[0-9]{1,} ")
	return reg.ReplaceAllString(schema, "")
}