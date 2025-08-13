package cpu

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func readFile(path string) (string, error) {
	contents, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(contents)), nil
}

func parseUint(s string) (uint64, error) {
	v, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		intValue, intErr := strconv.ParseInt(s, 10, 64)
		if intErr == nil && intValue < 0 {
			return 0, nil
		} else if intErr != nil &&
			intErr.(*strconv.NumError).Err == strconv.ErrRange &&
			intValue < 0 {
			return 0, nil
		}
		return 0, err
	}
	return v, nil
}

func ParseUintList(val string) (map[int]bool, error) {
	if val == "" {
		return map[int]bool{}, nil
	}

	availableInts := make(map[int]bool)
	split := strings.Split(val, ",")
	errInvalidFormat := fmt.Errorf("os/stat: invalid format: %s", val)
	for _, r := range split {
		if !strings.Contains(r, "-") {
			v, err := strconv.Atoi(r)
			if err != nil {
				return nil, errInvalidFormat
			}
			availableInts[v] = true
		} else {
			ss := strings.SplitN(r, "-", 2)
			minVal, err := strconv.Atoi(ss[0])
			if err != nil {
				return nil, errInvalidFormat
			}
			maxVal, err := strconv.Atoi(ss[1])
			if err != nil {
				return nil, errInvalidFormat
			}
			if maxVal < minVal {
				return nil, errInvalidFormat
			}
			for i := minVal; i <= maxVal; i++ {
				availableInts[i] = true
			}
		}
	}
	return availableInts, nil
}

func readLines(filename string) ([]string, error) {
	return readLinesOffsetN(filename, 0, -1)
}

func readLinesOffsetN(filename string, offset uint, n int) ([]string, error) {
	f, err := os.Open(filename)
	if err != nil {
		return []string{""}, err
	}
	defer f.Close()

	var ret []string

	r := bufio.NewReader(f)
	for i := 0; i < n+int(offset) || n < 0; i++ {
		line, err := r.ReadString('\n')
		if err != nil {
			break
		}
		if i < int(offset) {
			continue
		}
		ret = append(ret, strings.Trim(line, "\n"))
	}

	return ret, nil
}
