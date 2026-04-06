package maven

import (
	"fmt"
	"strings"
)

func main() {
	output := `[INFO] --- dependency:3.7.0:tree (default-cli) @ module-a ---
[INFO] com.example:module-a:jar:1.0-SNAPSHOT
[INFO] \- junit:junit:jar:4.13.2:test
[INFO]    \- org.hamcrest:hamcrest-core:jar:1.3:test`

	lines := strings.Split(output, "\n")
	for i, line := range lines {
		fmt.Printf("Line %d: %q\n", i, line)
		if len(line) > 0 && (line[0] == '\\' || line[0] == '+' || line[0] == '|') {
			fmt.Printf("  -> MATCHES prefix check\n")
		}
	}
}
