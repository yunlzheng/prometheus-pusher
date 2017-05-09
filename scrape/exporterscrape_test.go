package scrape

import (
	"testing"
	"strings"
	"fmt"
)

func TestInstance(t *testing.T) {
	instance := strings.Replace("192.168.0.01", ".", "-", -1)
	fmt.Println(instance)
}