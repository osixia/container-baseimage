package helpers

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func ChangelogGetVersionSection(changelog, version string) (string, error) {
	f, err := os.Open(changelog)
	if err != nil {
		return "", err
	}
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Print(err.Error())
		}
	}()

	sectionHeaderPrefix := "## "
	startPrefix := sectionHeaderPrefix + version

	var b strings.Builder
	inSection := false

	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := sc.Text()

		if !inSection && isVersionHeader(line, startPrefix) {
			inSection = true
			continue
		}

		if inSection {
			if strings.HasPrefix(line, sectionHeaderPrefix) && !isVersionHeader(line, startPrefix) {
				break
			}
			b.WriteString(line)
			b.WriteByte('\n')
		}
	}
	if err := sc.Err(); err != nil {
		return "", err
	}

	if !inSection {
		return "", nil
	}

	return strings.TrimRight(b.String(), "\n"), nil
}

// isVersionHeader reports whether line is the header for startPrefix's
// version, rather than a different version that merely shares it as a
// textual prefix (e.g. "## 1.2.3-1" or "## 1.2.30" when startPrefix is
// "## 1.2.3").
func isVersionHeader(line, startPrefix string) bool {

	if !strings.HasPrefix(line, startPrefix) {
		return false
	}

	rest := line[len(startPrefix):]
	if rest == "" {
		return true
	}

	c := rest[0]

	return c != '.' && c != '-' && (c < '0' || c > '9') && (c < 'a' || c > 'z') && (c < 'A' || c > 'Z')
}
