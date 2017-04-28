/*
@version: 1.0
@author: allanpan
@license:  Apache Licence
@contact: panjf2000@gmail.com  
@site: 
@file: fileutil.go
@time: 2017/3/22 19:18
@tag: 1,2,3
@todo: ...
*/
package tool

import (
	"errors"
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/robfig/cron"
	"io/ioutil"
	"os"
	"time"
)

// 列出指定路径中的文件和目录
// 如果目录不存在，则返回空slice
func ScanDir(directory string) []string {
	file, err := os.Open(directory)
	if err != nil {
		return []string{}
	}
	names, err := file.Readdirnames(-1)
	if err != nil {
		return []string{}
	}
	return names
}

// 判断给定文件名是否是一个目录
// 如果文件名存在并且为目录则返回 true。如果 filename 是一个相对路径，则按照当前工作目录检查其相对路径。
func IsDir(filename string) bool {
	return isFileOrDir(filename, true)
}

// 判断给定文件名是否为一个正常的文件
// 如果文件存在且为正常的文件则返回 true
func IsFile(filename string) bool {
	return isFileOrDir(filename, false)
}

// 判断是文件还是目录，根据decideDir为true表示判断是否为目录；否则判断是否为文件
func isFileOrDir(filename string, decideDir bool) bool {
	fileInfo, err := os.Stat(filename)
	if err != nil {
		return false
	}
	isDir := fileInfo.IsDir()
	if decideDir {
		return isDir
	}
	return !isDir
}

func CheckFileIsExist(filepath string) bool {
	var exist = true
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		exist = false
	}
	return exist
}

func CheckDirEmpty(dirPath string) (bool, []os.FileInfo) {
	fis, _ := ioutil.ReadDir(dirPath)
	if len(fis) == 0 {
		return true, fis
	} else {
		return false, fis
	}

}

func DelDir(dirPath string) error {
	if !IsDir(dirPath) {
		return errors.New("given path is not a dir.")
	}
	if empty, _ := CheckDirEmpty(dirPath); !empty {
		err := os.RemoveAll(dirPath)
		if err != nil {
			return err
		}
	}
	err := os.Remove(dirPath)
	return err

}

func InitLog(logpath string) (*logrus.Logger, error) {
	newLogpath := fmt.Sprintf("%s.%s", logpath, time.Now().Format("20060102"))
	logger := logrus.New()
	// Log as JSON instead of the default ASCII formatter.
	logger.Formatter = &logrus.TextFormatter{}

	// Output to stderr instead of stdout, could also be a file.
	f, err := os.OpenFile(newLogpath, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0660)
	if err != nil {
		return nil, err
	}
	logger.Out = f

	// Only log the warning severity or above.
	logger.Level = logrus.DebugLevel
	logger.SetNoLock()

	// set crontab task to generate a new log file with names ending in date-format suffix.
	spec := "0 0 1 * * *" // 1:00 am every day.
	c := cron.New()
	c.AddFunc(spec, func() {
		cronLogpath := fmt.Sprintf("%s.%s", logpath, time.Now().Format("20060102"))
		logger.Out, _ = os.OpenFile(cronLogpath, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0660)
	})
	c.Start()

	return logger, nil

}
