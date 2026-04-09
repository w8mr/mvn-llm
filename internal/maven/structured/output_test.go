package structured

import (
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
[INFO] Successfully compiled 1 source file
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
[WARNING] Warning in module
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
[WARNING] Warning in module-a
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
[WARNING] Warning in module-b
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

func TestTextSummary_MultipleModulesMultipleErrors(t *testing.T) {
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
[ERROR] Error in module-a
[WARNING] Warning in module-a
[INFO]
[INFO] --- compiler:3.15.0:compile (default-compile) @ module-a ---
[ERROR] Another error
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
[INFO] module-a ....................................... FAILURE [  0.500 s]
[INFO] module-b ....................................... FAILURE [  0.100 s]
[INFO] ------------------------------------------------------------------------
[INFO] BUILD FAILURE
[INFO] ------------------------------------------------------------------------`

	// When there's a failure, show errors first, then successes
	runTextSummaryFromMaven(t, mavenOutput, "Failure:\nmodule-a:\nError in module-a\nAnother error\n\nmodule-b:\nError in module-b")
}
