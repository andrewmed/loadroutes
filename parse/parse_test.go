package parse

import (
	"bufio"
	"strings"
	"testing"
)

// first line should be skipped
const dump = `Updated: 2019-01-29 01:07:00 +0000
1.179.201.18;;;<F1><F3><E4>;2-946/13;2013-06-10
104.20.25.94 | 104.20.26.94 | 104.24.112.70 | 104.24.113.70 | 104.25.135.114 | 104.25.136.114 | 104.28.8.110 | 104.28.9.110;*.fon-infosport.info;;<D4><CD><D1>;2-6-27 2016-06-07-90-<C0><C8>;2016-06-14
31.220.3.10/24;2ndfl-moskva.ru;;<F1><F3><E4>;2<E0>-1584/16;2016-11-17
`

func TestParse(t *testing.T) {
	var processed int
	reader := bufio.NewReader(strings.NewReader(dump))
	for {
		inLine := Parse(reader)
		if len(inLine) == 0 {
			break
		}
		processed += len(inLine)
	}
	if processed != 10 {
		t.Error("Error parsing dump file")
	}
}
