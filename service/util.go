package service

import (
	"fmt"
	"struct_sync/logger"
	"reflect"
	"regexp"
	"strings"
	"time"
)

func inStringSlice(str string, strSli []string) bool {
	for _, v := range strSli {
		if str == v {
			return true
		}
	}
	return false
}

func simpleMatch(patternStr string, str string, msg ...string) bool {
	str = strings.TrimSpace(str)
	patternStr = strings.TrimSpace(patternStr)
	if patternStr == str {
		logger.Info("simple_match:suc,equal", msg, "patternStr:", patternStr, "str:", str)
		return true
	}
	pattern := "^" + strings.Replace(patternStr, "*", `.*`, -1) + "$"
	match, err := regexp.MatchString(pattern, str)
	if err != nil {
		logger.Fatal("simple_match:error", msg, "patternStr:", patternStr, "pattern:", pattern, "str:", str, "err:", err)
	}
	if match {

	}
	return match
}

func maxMapKeyLen(data interface{}, ext int) string {
	l := 0

	for _, k := range reflect.ValueOf(data).MapKeys() {
		if k.Len() > l {
			l = k.Len()
		}
	}
	return fmt.Sprintf("%d", l+ext)
}

func getSyncKey() string {
	return time.Now().Format("20060102150405")
}
