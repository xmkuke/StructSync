// StructSync project main.go
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"runtime"
	"runtime/debug"
	"strings"
	"struct_sync/common"
	"struct_sync/service"
	"time"
)

const
	ConfName = "app.conf"


func ReadConf(confFile string) *service.GlobalSet {
	if data, err := ioutil.ReadFile(confFile); err != nil {
		return nil
	} else {
		gs := &service.GlobalSet{}
		if err := json.Unmarshal(data, gs); err != nil {
			fmt.Println("config file parser error: ", err)
			return nil
		} else {
			return gs
		}
	}
}

func main() {
	inputFile := flag.String("i", "", "Default read source schema info from database， use -i，read source schema info from file")
	dropUnnecessary := flag.Bool("c", false, "Use the param execute delete unnecessary field / index ")
	output := flag.String("o", "", "Save adjust SQL to file")
	execute := flag.Bool("e", true, "Execute adjust SQL to dest database, default true")

	flag.Parse()

	// Set CPU Numbers
	runtime.GOMAXPROCS(runtime.NumCPU())
	// Init log param
	now := time.Now()
	nowDate := fmt.Sprintf("%04d%02d%02d", now.Year(), now.Month(), now.Day())
	nowTime := fmt.Sprintf("%02d%02d%02d", now.Hour(), now.Minute(), now.Second())

	// Read config file
	confFile := common.GetConfigFile(os.Args[0], ConfName)
	exists, err := common.PathExists(confFile)
	if err != nil || !exists {
		fmt.Printf("The config file [%s] not exists!\r\n", confFile)
		os.Exit(1)
	}
	globalSetting := ReadConf(confFile)

	if "" == globalSetting.LogFileName {
		globalSetting.LogFileName = "StructSync_${date}.log"
	}
	logFileName := strings.ReplaceAll(globalSetting.LogFileName, "${date}", nowDate)
	logFileName = strings.ReplaceAll(logFileName, "${time}", nowTime)


	common.InitLogConfig(globalSetting.LogLevel, logFileName, globalSetting.LogPath)

	// Delete surplus field
	globalSetting.DropUnecessary = *dropUnnecessary
	// Execute adjust sql
	globalSetting.ExecuteSQL = *execute

	globalSetting.InputSql = *inputFile // Input path (must absolute path)

	if len(*output) > 0 {
		globalSetting.OutputDir = *output // Output path
	}
	if len(*inputFile) > 0 {
		globalSetting.InputMode = service.FileMode
	} else {
		globalSetting.InputMode = service.DbMode
	}

	if globalSetting.TimeOut == "" {
		globalSetting.TimeOut = "600s"
	}

	service.InitGlobalSet(globalSetting)

	if globalSetting.InputMode != service.DbMode { // from file
		fmt.Println("Sync Mode: Use file sync struct")
	} else {
		fmt.Println("Sync Mode: Use database sync struct")
	}

	t := service.NewMyTimer()
	fmt.Println("Database struct sync begin!")

	defer (func() {
		if err := recover(); err != nil {
			t.Stop()
			fmt.Println("Database struct sync interrupt!", err)
			debug.PrintStack()
		}
	})()

	// Start sync struct
	service.StartDatabaseSync()


	t.Stop()

	// Complete
	fmt.Println("Database struct sync finished! Time elapsed:", t.UsedSecond())
}
