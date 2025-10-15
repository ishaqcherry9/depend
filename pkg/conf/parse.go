package conf

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"path"
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

func Parse(configFile string, obj interface{}, reloads ...func()) error {
	confFileAbs, err := filepath.Abs(configFile)
	if err != nil {
		return err
	}

	filePathStr, filename := filepath.Split(confFileAbs)
	ext := strings.TrimLeft(path.Ext(filename), ".")
	if ext != "toml" {
		filename = strings.ReplaceAll(filename, "."+ext, "")
	}

	viper.AddConfigPath(filePathStr)
	viper.SetConfigName(filename)
	viper.SetConfigType(ext)
	err = viper.ReadInConfig()
	if err != nil {
		return err
	}

	err = viper.Unmarshal(obj)
	if err != nil {
		return err
	}

	if len(reloads) > 0 {
		watchConfig(obj, reloads...)
	}

	return nil
}

func ParseConfigData(data []byte, format string, obj interface{}) error {
	viper.SetConfigType(format)
	err := viper.ReadConfig(bytes.NewBuffer(data))
	if err != nil {
		return err
	}

	return viper.Unmarshal(obj)
}

func watchConfig(obj interface{}, reloads ...func()) {
	viper.WatchConfig()

	viper.OnConfigChange(func(e fsnotify.Event) {
		err := viper.Unmarshal(obj)
		if err != nil {
			fmt.Println("viper.Unmarshal error: ", err)
		} else {
			for _, reload := range reloads {
				reload()
			}
		}
	})
}

func Show(obj interface{}, fields ...string) string {
	var out string

	data, err := json.MarshalIndent(obj, "", "    ")
	if err != nil {
		fmt.Println("json.MarshalIndent error: ", err)
		return ""
	}

	buf := bufio.NewReader(bytes.NewReader(data))
	for {
		line, err := buf.ReadString('\n')
		if err != nil {
			break
		}
		fields = append(fields, `"dsn"`, `"password"`, `"pwd"`)

		out += hideSensitiveFields(line, fields...)
	}

	return out
}

func hideSensitiveFields(line string, fields ...string) string {
	for _, field := range fields {
		if strings.Contains(line, field) {
			index := strings.Index(line, field)
			if strings.Contains(line, "@") && strings.Contains(line, ":") {
				return replaceDSN(line)
			}
			return fmt.Sprintf("%s: \"******\",\n", line[:index+len(field)])
		}
	}

	if strings.Contains(line, "@") && strings.Contains(line, ":") {
		return replaceDSN(line)
	}

	return line
}

func replaceDSN(str string) string {
	data := []byte(str)
	start, end := 0, 0
	for k, v := range data {
		if v == ':' {
			start = k
		}
		if v == '@' {
			end = k
			break
		}
	}

	if start >= end {
		return str
	}

	return fmt.Sprintf("%s******%s", data[:start+1], data[end:])
}
