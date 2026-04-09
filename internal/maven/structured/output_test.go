package structured

import (
	"encoding/json"
	"regexp"
	"strings"
	"testing"
)

// runTextSummaryFromMaven parses mavenOutput through the full parser
// and checks the expected text summary.
func runTextSummaryFromMaven(t *testing.T, mavenOutput, expectedSummary string) {
	lines := strings.Split(strings.ReplaceAll(mavenOutput, "\r\n", "\n"), "\n")
	parser := NewOutputParser()
	out := parser.ParseOutput(lines, nil, ParseConfig{})
	result := TextSummary(out)
	if result != expectedSummary {
		t.Errorf("Expected:\n%q\nGot:\n%q", expectedSummary, result)
	}
}

func TestTextSummary_SimpleSuccess(t *testing.T) {
	mavenOutput := `[INFO] Scanning for projects...
[INFO] ------------------------------------------------------------------------
[INFO] Reactor Build Order:
[INFO] 
[INFO] my-app                                                       [jar]
[INFO] 
[INFO] ---< com.example:my-app >---
[INFO] Building my-app 1.0-SNAPSHOT                                   [1/1]
[INFO]   from myapp/pom.xml
[INFO] --------------------------------[ jar ]---------------------------------
[INFO] 
[INFO] --- clean:3.2.0:clean (default-clean) @ my-app ---
[INFO] 
[INFO] --- compiler:3.15.0:compile (default-compile) @ my-app ---
[INFO] Compiling 1 source file
[INFO] 
[INFO] --- jar:3.5.0:jar (default-jar) @ my-app ---
[INFO] Building jar: target/my-app.jar
[INFO] ------------------------------------------------------------------------
[INFO] Reactor Summary:
[INFO] my-app ....................................... SUCCESS [  1.234 s]
[INFO] ------------------------------------------------------------------------
[INFO] BUILD SUCCESS
[INFO] ------------------------------------------------------------------------`

	runTextSummaryFromMaven(t, mavenOutput, "Successful:\nmy-app:\nBuilding jar: target/my-app.jar")
}

func TestTextSummary_WithWarnings(t *testing.T) {
	mavenOutput := `[INFO] Scanning for projects...
[INFO] ------------------------------------------------------------------------
[INFO] Reactor Build Order:
[INFO] 
[INFO] my-app                                                       [jar]
[INFO] 
[INFO] ---< com.example:my-app >---
[INFO] Building my-app 1.0-SNAPSHOT                                   [1/1]
[INFO]   from myapp/pom.xml
[INFO] --------------------------------[ jar ]---------------------------------
[INFO] 
[INFO] --- compiler:3.15.0:compile (default-compile) @ my-app ---
[INFO] Compiling 1 source file
[WARNING] system modules path not set in conjunction with -source 11
[INFO] 
[INFO] --- jar:3.5.0:jar (default-jar) @ my-app ---
[INFO] Building jar: target/my-app.jar
[INFO] ------------------------------------------------------------------------
[INFO] Reactor Summary:
[INFO] my-app ....................................... SUCCESS [  1.234 s]
[INFO] ------------------------------------------------------------------------
[INFO] BUILD SUCCESS
[INFO] ------------------------------------------------------------------------`

	runTextSummaryFromMaven(t, mavenOutput, "Successful:\nmy-app:\nsystem modules path not set in conjunction with -source 11")
}

func TestTextSummary_WithErrors(t *testing.T) {
	mavenOutput := `[INFO] Scanning for projects...
[INFO] ------------------------------------------------------------------------
[INFO] Reactor Build Order:
[INFO] 
[INFO] my-app                                                       [jar]
[INFO] 
[INFO] ---< com.example:my-app >---
[INFO] Building my-app 1.0-SNAPSHOT                                   [1/1]
[INFO]   from myapp/pom.xml
[INFO] --------------------------------[ jar ]---------------------------------
[INFO] 
[INFO] --- compiler:3.15.0:compile (default-compile) @ my-app ---
[ERROR] COMPILATION ERROR
[INFO] src/main/java/App.java:10: error
[INFO] ------------------------------------------------------------------------
[INFO] Reactor Summary:
[INFO] my-app ....................................... FAILURE [  0.123 s]
[INFO] ------------------------------------------------------------------------
[INFO] BUILD FAILURE
[INFO] ------------------------------------------------------------------------`

	runTextSummaryFromMaven(t, mavenOutput, "Failure:\nmy-app:\nCOMPILATION ERROR")
}

func TestTextSummary_MultipleModules(t *testing.T) {
	mavenOutput := `[INFO] Scanning for projects...
[INFO] ------------------------------------------------------------------------
[INFO] Reactor Build Order:
[INFO] 
[INFO] module-a                                                       [jar]
[INFO] module-b                                                       [jar]
[INFO] 
[INFO] ---< com.example:module-a >---
[INFO] Building module-a 1.0-SNAPSHOT                                 [1/2]
[INFO]   from module-a/pom.xml
[INFO] --------------------------------[ jar ]---------------------------------
[INFO] 
[INFO] --- clean:3.2.0:clean (default-clean) @ module-a ---
[INFO] 
[INFO] --- compiler:3.15.0:compile (default-compile) @ module-a ---
[INFO] Compiling module-a
[INFO] 
[INFO] ---< com.example:module-b >---
[INFO] Building module-b 1.0-SNAPSHOT                                 [2/2]
[INFO]   from module-b/pom.xml
[INFO] --------------------------------[ jar ]---------------------------------
[INFO] 
[INFO] --- clean:3.2.0:clean (default-clean) @ module-b ---
[INFO] 
[INFO] --- compiler:3.15.0:compile (default-compile) @ module-b ---
[ERROR] Error in module-b
[INFO] ------------------------------------------------------------------------
[INFO] Reactor Summary:
[INFO] module-a ....................................... SUCCESS [  0.500 s]
[INFO] module-b ....................................... FAILURE [  0.100 s]
[INFO] ------------------------------------------------------------------------
[INFO] BUILD FAILURE
[INFO] ------------------------------------------------------------------------`

	// When there's a failure, show errors first, then successes
	runTextSummaryFromMaven(t, mavenOutput, "Failure:\nmodule-b:\nError in module-b")
}

func TestTextSummary_DebugParse(t *testing.T) {
	mavenOutput := `[INFO] Scanning for projects...
[INFO] ------------------------------------------------------------------------
[INFO] Reactor Build Order:
[INFO] 
[INFO] my-app                                                       [jar]
[INFO] 
[INFO] ---< com.example:my-app >---
[INFO] Building my-app 1.0-SNAPSHOT                                   [1/1]
[INFO]   from myapp/pom.xml
[INFO] --------------------------------[ jar ]---------------------------------
[INFO] 
[INFO] --- clean:3.2.0:clean (default-clean) @ my-app ---
[INFO] 
[INFO] --- compiler:3.15.0:compile (default-compile) @ my-app ---
[INFO] Compiling 1 source file
[INFO] 
[INFO] --- jar:3.5.0:jar (default-jar) @ my-app ---
[INFO] Building jar: target/my-app.jar
[INFO] ------------------------------------------------------------------------
[INFO] Reactor Summary:
[INFO] my-app ....................................... SUCCESS [  1.234 s]
[INFO] ------------------------------------------------------------------------
[INFO] BUILD SUCCESS
[INFO] ------------------------------------------------------------------------`

	lines := strings.Split(strings.ReplaceAll(mavenOutput, "\r\n", "\n"), "\n")
	parser := NewOutputParser()
	out := parser.ParseOutput(lines, nil, ParseConfig{})

	t.Logf("Root children: %d", len(out.Root.Children))
	for i, c := range out.Root.Children {
		t.Logf("Child %d: Type=%q, Name=%q", i, c.Type, c.Name)
		if c.Type == "module" {
			t.Logf("  Module Lines: %v", c.Lines)
			for j, b := range c.Children {
				t.Logf("  Build %d: Type=%q, Meta=%+v", j, b.Type, b.Meta)
			}
		}
		if c.Type == "initialization" {
			t.Logf("  Initialization Lines: %d lines", len(c.Lines))
			t.Logf("  First 3: %v", c.Lines[:min(3, len(c.Lines))])
		}
		if c.Type == "unparsable" {
			t.Logf("  Unparsable Lines count: %d", len(c.Lines))
			t.Logf("  First unparsable line: %q", c.Lines[0])
		}
	}

	// Now check what the regex actually matches for problematic line
	testLine := "[INFO] ---------------------< com.example:my-app >----------------------"
	ModuleHeaderRegex := regexp.MustCompile(`^\[INFO\] [-]+< ([^>]+) >[-]+$`)
	t.Logf("Testing line: %q", testLine)
	t.Logf("ModuleHeaderRegex.Match: %v", ModuleHeaderRegex.MatchString(testLine))
}

func TestTextSummary_DebugJSON(t *testing.T) {
	mavenOutput := `[INFO] Scanning for projects...
[INFO] ------------------------------------------------------------------------
[INFO] Reactor Build Order:
[INFO] 
[INFO] my-app                                                       [jar]
[INFO] 
[INFO] ---< com.example:my-app >---
[INFO] Building my-app 1.0-SNAPSHOT                                   [1/1]
[INFO]   from myapp/pom.xml
[INFO] --------------------------------[ jar ]---------------------------------
[INFO] 
[INFO] --- clean:3.2.0:clean (default-clean) @ my-app ---
[INFO] 
[INFO] --- compiler:3.15.0:compile (default-compile) @ my-app ---
[INFO] Compiling 1 source file
[INFO] 
[INFO] --- jar:3.5.0:jar (default-jar) @ my-app ---
[INFO] Building jar: target/my-app.jar
[INFO] ------------------------------------------------------------------------
[INFO] Reactor Summary:
[INFO] my-app ....................................... SUCCESS [  1.234 s]
[INFO] ------------------------------------------------------------------------
[INFO] BUILD SUCCESS
[INFO] ------------------------------------------------------------------------`

	lines := strings.Split(strings.ReplaceAll(mavenOutput, "\r\n", "\n"), "\n")
	parser := NewOutputParser()
	out := parser.ParseOutput(lines, nil, ParseConfig{})

	jsonBytes, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal JSON: %v", err)
	}
	t.Logf("JSON Output:\n%s", string(jsonBytes))
}
