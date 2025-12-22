package install

import (
	"bufio"
	"os"
	"strings"
)

func parseChecksumFile(path string) (map[string]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	m := map[string]string{}
	s := bufio.NewScanner(f)
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		h := strings.TrimSpace(fields[0])
		name := strings.TrimSpace(fields[len(fields)-1])
		m[name] = h
	}
	if err := s.Err(); err != nil {
		return nil, err
	}
	return m, nil
}
