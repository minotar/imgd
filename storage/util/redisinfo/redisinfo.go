package redisinfo

import (
	"errors"
	"strconv"
	"strings"
)

// Parses a reply from redis INFO into a nice map.
func ParseStats(resp []byte) (map[string]string, error) {
	raw := strings.Split(string(resp), "\r\n")
	output := map[string]string{}

	for _, line := range raw {
		// Skip blank lines or comment lines
		if len(line) == 0 || string(line[0]) == "#" {
			continue
		}
		// Get the position the seperator breaks
		sep := strings.Index(line, ":")
		if sep == -1 {
			return nil, errors.New("Invalid line: " + line)
		}
		output[line[:sep]] = line[sep+1:]
	}

	return output, nil
}

func ParseUsedMemory(resp []byte) uint64 {
	data, err := ParseStats(resp)
	if err != nil {
		return 0
	}

	mem, _ := strconv.ParseUint(data["used_memory"], 10, 64)
	return mem
}
