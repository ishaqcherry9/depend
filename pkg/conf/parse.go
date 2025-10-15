package conf

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
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

// ConfigFormat represents supported configuration formats
type ConfigFormat string

const (
	FormatYAML    ConfigFormat = "yaml"
	FormatJSON    ConfigFormat = "json"
	FormatTOML    ConfigFormat = "toml"
	FormatINI     ConfigFormat = "ini"
	FormatXML     ConfigFormat = "xml"
	FormatProps   ConfigFormat = "properties"
	FormatUnknown ConfigFormat = "unknown"
)

// AutoDetectFormat automatically detects the configuration format from data content
func AutoDetectFormat(data []byte) ConfigFormat {
	content := strings.TrimSpace(string(data))

	// Check for YAML format (more flexible, check last)
	if isYAML(content) {
		return FormatYAML
	}

	// Check for JSON format first (most specific)
	if isJSON(content) {
		return FormatJSON
	}

	// Check for XML format (very specific)
	if isXML(content) {
		return FormatXML
	}

	// Check for TOML format (has specific section syntax)
	if isTOML(content) {
		return FormatTOML
	}

	// Check for INI format (similar to TOML but different syntax)
	if isINI(content) {
		return FormatINI
	}

	// Check for Properties format (simple key=value, no sections)
	if isProperties(content) {
		return FormatProps
	}

	return FormatUnknown
}

// isJSON checks if content is valid JSON
func isJSON(content string) bool {
	var js json.RawMessage
	return json.Unmarshal([]byte(content), &js) == nil
}

// isYAML checks if content looks like YAML
func isYAML(content string) bool {
	// YAML indicators: key: value pairs, lists with -, etc.
	yamlPatterns := []string{
		`^\s*[a-zA-Z_][a-zA-Z0-9_]*\s*:\s*`, // key: value
		`^\s*-\s*`,                          // list item
		`^\s*#`,                             // comment
		`^\s*---`,                           // document separator
	}

	for _, pattern := range yamlPatterns {
		if matched, _ := regexp.MatchString(pattern, content); matched {
			// Try to parse as YAML to confirm
			var y interface{}
			if err := yaml.Unmarshal([]byte(content), &y); err == nil {
				return true
			}
		}
	}
	return false
}

// isTOML checks if content looks like TOML
func isTOML(content string) bool {
	lines := strings.Split(content, "\n")
	hasSection := false
	hasKeyValueWithSpaces := false

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Check for section header [section]
		if matched, _ := regexp.MatchString(`^\s*\[.*\]\s*$`, line); matched {
			hasSection = true
		}

		// Check for key = value (TOML uses spaces around =)
		// Look for the specific pattern: key = value (with spaces around =)
		if matched, _ := regexp.MatchString(`^\s*[a-zA-Z_][a-zA-Z0-9_]*\s*=\s*.*$`, line); matched {
			// Check if the line contains " = " (with spaces around =)
			if strings.Contains(line, " = ") {
				hasKeyValueWithSpaces = true
			}
		}
	}

	// TOML is characterized by spaces around = and sections
	return hasSection && hasKeyValueWithSpaces
}

// isINI checks if content looks like INI
func isINI(content string) bool {
	lines := strings.Split(content, "\n")
	hasSection := false
	hasKeyValueWithoutSpaces := false

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, ";") {
			continue
		}

		// Check for section header [section]
		if matched, _ := regexp.MatchString(`^\s*\[.*\]\s*$`, line); matched {
			hasSection = true
		}

		// Check for key=value (no spaces around =, typical INI style)
		if matched, _ := regexp.MatchString(`^\s*[a-zA-Z_][a-zA-Z0-9_]*\s*=\s*.*$`, line); matched {
			// Check if it's key=value (no spaces around =) vs key = value (with spaces)
			if matched, _ := regexp.MatchString(`^\s*[a-zA-Z_][a-zA-Z0-9_]*\s*=\s*.*$`, line); matched {
				// If there are spaces around =, it's more likely TOML
				if !strings.Contains(line, " = ") {
					hasKeyValueWithoutSpaces = true
				}
			}
		}
	}

	// INI is characterized by sections and key=value (no spaces around =)
	return hasSection && hasKeyValueWithoutSpaces
}

// isXML checks if content looks like XML
func isXML(content string) bool {
	xmlPattern := `^\s*<\?xml\s+version=`
	matched, _ := regexp.MatchString(xmlPattern, content)
	return matched
}

// isProperties checks if content looks like Properties file
func isProperties(content string) bool {
	lines := strings.Split(content, "\n")
	hasKeyValue := false
	hasSections := false

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Check for section headers
		if strings.Contains(line, "[") && strings.Contains(line, "]") {
			hasSections = true
		}

		// Check for key=value pattern (no spaces around =, no sections)
		if matched, _ := regexp.MatchString(`^\s*[a-zA-Z_][a-zA-Z0-9_.]*\s*=\s*.*$`, line); matched {
			hasKeyValue = true
		}
	}

	// Properties files typically don't have sections, just key=value pairs
	// If it has sections, it's more likely INI or TOML
	return hasKeyValue && !hasSections
}

// ParseConfigDataWithAutoDetection parses configuration data with automatic format detection
func ParseConfigDataWithAutoDetection(data []byte, obj interface{}) error {
	format := AutoDetectFormat(data)
	if format == FormatUnknown {
		return fmt.Errorf("unable to detect configuration format from data")
	}

	return ParseConfigData(data, string(format), obj)
}

// ParseConfigDataWithFallback parses configuration data with format detection and fallback
func ParseConfigDataWithFallback(data []byte, preferredFormat string, obj interface{}) error {
	// First try the preferred format
	err := ParseConfigData(data, preferredFormat, obj)
	if err == nil {
		return nil
	}

	// If preferred format fails, try auto-detection
	detectedFormat := AutoDetectFormat(data)
	if detectedFormat != FormatUnknown && detectedFormat != ConfigFormat(preferredFormat) {
		return ParseConfigData(data, string(detectedFormat), obj)
	}

	return fmt.Errorf("failed to parse configuration data with format '%s' and auto-detection failed: %v", preferredFormat, err)
}

// ParseConfigDataMultiFormat tries multiple formats until one succeeds
func ParseConfigDataMultiFormat(data []byte, obj interface{}) error {
	formats := []ConfigFormat{FormatYAML, FormatJSON, FormatTOML, FormatINI}

	for _, format := range formats {
		err := ParseConfigData(data, string(format), obj)
		if err == nil {
			return nil
		}
	}

	return fmt.Errorf("failed to parse configuration data with any supported format")
}

// ValidateConfigFormat validates if the given format is supported
func ValidateConfigFormat(format string) error {
	supportedFormats := []string{"yaml", "yml", "json", "toml", "ini", "xml", "properties"}
	format = strings.ToLower(format)

	for _, supported := range supportedFormats {
		if format == supported {
			return nil
		}
	}

	return fmt.Errorf("unsupported configuration format: %s", format)
}

// GetSupportedFormats returns a list of supported configuration formats
func GetSupportedFormats() []string {
	return []string{"yaml", "yml", "json", "toml", "ini", "xml", "properties"}
}
