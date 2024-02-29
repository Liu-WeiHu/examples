package cache

import (
	"errors"
	"regexp"
	"strconv"
	"strings"
)

const (
	B = 1 << (iota * 10)
	KB
	MB
	GB
	TB
)

const (
	forever  = -1 // 时间状态值 永久
	notExist = -2 // key不存在
)

func ToBytes(size string) (int64, error) {
	re, err := regexp.Compile("[0-9]+")
	if err != nil {
		return 0, err
	}
	unit := string(re.ReplaceAll([]byte(size), []byte("")))
	num, err := strconv.ParseInt(strings.Replace(size, unit, "", 1), 10, 64)
	if err != nil {
		return 0, err
	}
	unit = strings.ToUpper(unit)
	var res int64
	switch unit {
	case "B", "":
		res = num * B
	case "KB":
		res = num * KB
	case "MB":
		res = num * MB
	case "GB":
		res = num * GB
	case "TB":
		res = num * TB
	default:
		return 0, errors.New("input error")
	}
	if res == 0 {
		return 0, errors.New("input error")
	}
	return res, nil
}
