//运行 VERSION=v1.4.11 go run version.go

package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

func main() {
	cmd := exec.Command("git", "log", "--pretty=format:%s")
	out, err := cmd.Output()
	if err != nil {
		panic(err)
	}

	lines := strings.Split(string(out), "\n")

	added, fixed, changed := []string{}, []string{}, []string{}
	re := regexp.MustCompile(`^(feat|fix|chore|docs|refactor|perf):\s*(.*)$`)

	for _, line := range lines {
		if matches := re.FindStringSubmatch(line); len(matches) > 0 {
			switch matches[1] {
			case "feat":
				added = append(added, "- "+matches[2])
			case "fix":
				fixed = append(fixed, "- "+matches[2])
			case "chore", "refactor", "perf":
				changed = append(changed, "- "+matches[2])
			}
		}
	}

	version := os.Getenv("VERSION")
	if version == "" {
		version = "v" + time.Now().Format("20060102")
	}
	date := time.Now().Format("2006-01-02")

	changelog := fmt.Sprintf("## [%s] - %s\n", version, date)
	if len(added) > 0 {
		changelog += "### Added\n" + strings.Join(added, "\n") + "\n\n"
	}
	if len(changed) > 0 {
		changelog += "### Changed\n" + strings.Join(changed, "\n") + "\n\n"
	}
	if len(fixed) > 0 {
		changelog += "### Fixed\n" + strings.Join(fixed, "\n") + "\n\n"
	}

	// 追加写入到 CHANGELOG.md
	f, err := os.OpenFile("CHANGELOG.md", os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	writer := bufio.NewWriter(f)
	writer.WriteString(changelog + "\n---\n")
	writer.Flush()

	fmt.Println("✅ Changelog updated:", version)
}

