package scrape

import (
	"testing"
	"strings"
	"fmt"
)

func TestInstance(t *testing.T) {
	ip := strings.TrimLeft("http://192.168.0.01:8080", "http://")
	instanceLabel := strings.Replace(ip, ".", "-", -1)
	instanceLabel = strings.Replace(instanceLabel, ":", "-", -1)
	fmt.Println(instanceLabel)
}