package go_linux_proc_format

import "C"
import (
	"bufio"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
)

/*
#include <unistd.h>
*/
import "C"

type MilliValue int
type ProgramRunSec float64

// ProcStatus 表示 /proc/{pid}/status 文件中的信息
type ProcStatus struct {
	Name      string
	Umask     string
	State     string
	Tgid      int
	Ngid      int
	Pid       int
	PPid      int
	TracerPid int
	Uid       []int
	Gid       []int
	FDSize    int
	Groups    []int
	VmPeak    int
	VmSize    int
	VmLck     int
	VmPin     int
	VmHWM     int
	VmRSS     int
	RssAnon   int
	RssFile   int
	RssShmem  int
	VmData    int
	VmStk     int
	VmExe     int
	VmLib     int
	VmPTE     int
	VmSwap    int
	Threads   int
}

// ParseProcStatus 解析 /proc/{pid}/status 文件并返回 ProcessStatus 结构体
func ParseProcStatus(pid int) (*ProcStatus, error) {
	// 打开 /proc/{pid}/status 文件
	file, err := os.Open(fmt.Sprintf("/proc/%d/status", pid))
	if err != nil {
		return nil, err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			return
		}
	}(file)

	// 创建 ProcessStatus 结构体
	status := &ProcStatus{}

	// 使用 bufio.Scanner 读取文件内容
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		// 将每一行分割成 key 和 value
		line := scanner.Text()
		parts := strings.SplitN(line, ":", 2)
		if len(parts) < 2 {
			continue
		}
		//TODO
		key, value := strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
		value = strings.Replace(value, "kB", "", -1)
		value = strings.TrimSpace(value)
		// 根据 key 解析 value 并赋值给结构体字段
		switch key {
		case "Name":
			status.Name = value
		case "Umask":
			status.Umask = value
		case "State":
			re := regexp.MustCompile(`\((.*?)\)`)
			match := re.FindStringSubmatch(value)
			if match != nil {
				status.State = match[1]
			}
		case "Tgid":
			status.Tgid, _ = strconv.Atoi(value)
		case "Ngid":
			status.Ngid, _ = strconv.Atoi(value)
		case "Pid":
			status.Pid, _ = strconv.Atoi(value)
		case "PPid":
			status.PPid, _ = strconv.Atoi(value)
		case "TracerPid":
			status.TracerPid, _ = strconv.Atoi(value)
		case "Uid":
			status.Uid = parseToIntArray(value)
		case "Gid":
			status.Gid = parseToIntArray(value)
		case "FDSize":
			status.FDSize, _ = strconv.Atoi(value)
		case "Groups":
			status.Groups = parseToIntArray(value)
		case "VmPeak":
			status.VmPeak, _ = strconv.Atoi(value)
		case "VmSize":
			status.VmSize, _ = strconv.Atoi(value)
		case "VmLck":
			status.VmLck, _ = strconv.Atoi(value)
		case "VmPin":
			status.VmPin, _ = strconv.Atoi(value)
		case "VmHWM":
			status.VmHWM, _ = strconv.Atoi(value)
		case "VmRSS":
			status.VmRSS, _ = strconv.Atoi(value)
		case "RssAnon":
			status.RssAnon, _ = strconv.Atoi(value)
		case "RssFile":
			status.RssFile, _ = strconv.Atoi(value)
		case "RssShmem":
			status.RssShmem, _ = strconv.Atoi(value)
		case "VmData":
			status.VmData, _ = strconv.Atoi(value)
		case "VmStk":
			status.VmStk, _ = strconv.Atoi(value)
		case "VmExe":
			status.VmExe, _ = strconv.Atoi(value)
		case "VmLib":
			status.VmLib, _ = strconv.Atoi(value)
		case "VmPTE":
			status.VmPTE, _ = strconv.Atoi(value)
		case "VmSwap":
			status.VmSwap, _ = strconv.Atoi(value)
		case "Threads":
			status.Threads, _ = strconv.Atoi(value)
		}
	}

	// 检查是否有错误
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	// 返回 ProcessStatus 结构体
	return status, nil
}

// parseToIntArray 将字符串解析为整数数组
func parseToIntArray(str string) []int {
	var result []int
	for _, s := range strings.Fields(str) {
		i, err := strconv.Atoi(s)
		if err != nil {
			continue
		}
		result = append(result, i)
	}
	return result
}

func getProcStat(pid int) ([]string, error) {
	data, err := os.ReadFile(fmt.Sprintf("/proc/%d/stat", pid))
	if err != nil {
		return nil, err
	}
	return strings.Fields(string(data)), nil
}

func getUptime() (float64, error) {
	data, err := os.ReadFile("/proc/uptime")
	if err != nil {
		return 0, err
	}
	fields := strings.Fields(string(data))
	if len(fields) < 1 {
		return 0, fmt.Errorf("invalid uptime format")
	}
	return strconv.ParseFloat(fields[0], 64)
}

func GetMilliCPUUsage(pid int) (MilliValue, ProgramRunSec, error) {
	fields, err := getProcStat(pid)
	if err != nil {
		return 0, 0, err
	}
	if len(fields) < 22 {
		return 0, 0, fmt.Errorf("invalid stat file format")
	}

	utime, _ := strconv.ParseInt(fields[13], 10, 64)
	stime, _ := strconv.ParseInt(fields[14], 10, 64)
	starttime, _ := strconv.ParseInt(fields[21], 10, 64)

	uptime, err := getUptime()
	if err != nil {
		return 0, 0, err
	}
	clkTck := float64(C.sysconf(C._SC_CLK_TCK))
	currentTime := uptime * clkTck

	totalCPUTime := float64(utime + stime)
	processTime := currentTime - float64(starttime)

	cpuUsage := int(totalCPUTime / processTime * 1000)
	puptime := float64(utime+stime) / clkTck
	return MilliValue(cpuUsage), ProgramRunSec(puptime), nil
}

func getChildPIDs(pid int) ([]int, error) {
	childrenFile := fmt.Sprintf("/proc/%d/task/%d/children", pid, pid)
	content, err := os.ReadFile(childrenFile)
	if err != nil {
		log.Printf("error reading %s: %v \n", childrenFile, err)
		return nil, err
	}

	pidStrings := strings.Fields(string(content))
	childPIDs := make([]int, len(pidStrings))

	for i, pidString := range pidStrings {
		childPID, err := strconv.Atoi(pidString)
		if err != nil {
			return nil, err
		}
		childPIDs[i] = childPID
	}
	return childPIDs, nil
}
