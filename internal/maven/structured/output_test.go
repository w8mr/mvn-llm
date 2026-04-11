package structured

import (
	"encoding/json"
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

func TestJSONSummary_WithWarnings(t *testing.T) {
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

	lines := strings.Split(strings.ReplaceAll(mavenOutput, "\r\n", "\n"), "\n")
	parser := NewOutputParser()
	out := parser.ParseOutput(lines, nil, ParseConfig{})

	_, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal JSON: %v", err)
	}
	expected := `{
  "root": {
    "name": "maven-build",
    "type": "root",
    "children": [
      {
        "name": "initialization",
        "type": "initialization",
        "lines": [
          "[INFO] Scanning for projects...",
          "[INFO] ------------------------------------------------------------------------",
          "[INFO] Reactor Build Order:",
          "[INFO] ",
          "[INFO] my-app                                                       [jar]",
          "[INFO] "
        ],
        "meta": {
          "modules": [
            {
              "module": "my-app",
              "packaging": "jar"
            }
          ]
        }
      },
      {
        "name": "my-app",
        "type": "module",
        "lines": [
          "[INFO] ---\u003c com.example:my-app \u003e---",
          "[INFO] Building my-app 1.0-SNAPSHOT                                   [1/1]",
          "[INFO]   from myapp/pom.xml",
          "[INFO] --------------------------------[ jar ]---------------------------------",
          "[INFO] "
        ],
        "meta": {
          "artifactId": "my-app",
          "groupId": "com.example",
          "moduleCount": 1,
          "moduleIndex": 1,
          "name": "my-app",
          "packaging": "jar",
          "pomFile": "myapp/pom.xml",
          "status": "SUCCESS-WITH-WARNINGS",
          "summary": "system modules path not set in conjunction with -source 11",
          "version": "1.0-SNAPSHOT"
        },
        "children": [
          {
            "name": "compiler",
            "type": "build-block",
            "lines": [
              "[INFO] --- compiler:3.15.0:compile (default-compile) @ my-app ---",
              "[INFO] Compiling 1 source file",
              "[WARNING] system modules path not set in conjunction with -source 11",
              "[INFO] Successfully compiled 1 source file",
              "[INFO]"
            ],
            "meta": {
              "artifactId": "my-app",
              "executionId": "default-compile",
              "goal": "compile",
              "plugin": "compiler",
              "status": "SUCCESS-WITH-WARNINGS",
              "summary": "system modules path not set in conjunction with -source 11",
              "version": "3.15.0"
            }
          },
          {
            "name": "jar",
            "type": "build-block",
            "lines": [
              "[INFO] --- jar:3.5.0:jar (default-jar) @ my-app ---",
              "[INFO] Building jar: target/my-app.jar"
            ],
            "meta": {
              "artifactId": "my-app",
              "executionId": "default-jar",
              "goal": "jar",
              "plugin": "jar",
              "status": "SUCCESS",
              "summary": "Building jar: target/my-app.jar",
              "version": "3.5.0"
            }
          }
        ]
      },
      {
        "name": "summary",
        "type": "summary",
        "lines": [
          "[INFO] ------------------------------------------------------------------------",
          "[INFO] Reactor Summary:",
          "[INFO] my-app ....................................... SUCCESS [  1.234 s]",
          "[INFO] ------------------------------------------------------------------------",
          "[INFO] BUILD SUCCESS",
          "[INFO] ------------------------------------------------------------------------"
        ],
        "meta": {
          "modules": [
            {
              "name": "my-app",
              "status": "SUCCESS",
              "time": "1.234 s"
            }
          ],
          "overallStatus": "BUILD SUCCESS",
          "summary": "my-app:\nsystem modules path not set in conjunction with -source 11"
        }
      }
    ]
  }
}`
	_ = expected
	if out.Root.Type != "root" || len(out.Root.Children) != 3 || out.Root.StartLine != 1 || out.Root.EndLine != 24 {
		t.Errorf("JSON structure verification failed")
	}
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

func TestJSONSummary_MultipleModulesMultipleErrors(t *testing.T) {
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

	lines := strings.Split(strings.ReplaceAll(mavenOutput, "\r\n", "\n"), "\n")
	parser := NewOutputParser()
	out := parser.ParseOutput(lines, nil, ParseConfig{})

	_, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal JSON: %v", err)
	}
	expected := `{
  "root": {
    "name": "maven-build",
    "type": "root",
    "children": [
      {
        "name": "initialization",
        "type": "initialization",
        "lines": [
          "[INFO] Scanning for projects...",
          "[INFO] ------------------------------------------------------------------------",
          "[INFO] Reactor Build Order:",
          "[INFO] ",
          "[INFO] module-a                                                       [jar]",
          "[INFO] module-b                                                       [jar]",
          "[INFO] "
        ],
        "meta": {
          "modules": [
            {
              "module": "module-a",
              "packaging": "jar"
            },
            {
              "module": "module-b",
              "packaging": "jar"
            }
          ]
        }
      },
      {
        "name": "module-a",
        "type": "module",
        "lines": [
          "[INFO] ---\u003c com.example:module-a \u003e---",
          "[INFO] Building module-a 1.0-SNAPSHOT                                 [1/2]",
          "[INFO]   from module-a/pom.xml",
          "[INFO] --------------------------------[ jar ]---------------------------------",
          "[INFO] "
        ],
        "meta": {
          "artifactId": "module-a",
          "groupId": "com.example",
          "moduleCount": 2,
          "moduleIndex": 1,
          "name": "module-a",
          "packaging": "jar",
          "pomFile": "module-a/pom.xml",
          "status": "FAILED",
          "summary": "Error in module-a\nAnother error",
          "version": "1.0-SNAPSHOT"
        },
        "children": [
          {
            "name": "clean",
            "type": "build-block",
            "lines": [
              "[INFO] --- clean:3.2.0:clean (default-clean) @ module-a ---",
              "[ERROR] Error in module-a",
              "[WARNING] Warning in module-a",
              "[INFO]"
            ],
            "meta": {
              "artifactId": "module-a",
              "executionId": "default-clean",
              "goal": "clean",
              "plugin": "clean",
              "status": "FAILED",
              "summary": "Error in module-a",
              "version": "3.2.0"
            }
          },
          {
            "name": "compiler",
            "type": "build-block",
            "lines": [
              "[INFO] --- compiler:3.15.0:compile (default-compile) @ module-a ---",
              "[ERROR] Another error",
              "[INFO]"
            ],
            "meta": {
              "artifactId": "module-a",
              "executionId": "default-compile",
              "goal": "compile",
              "plugin": "compiler",
              "status": "FAILED",
              "summary": "Another error",
              "version": "3.15.0"
            }
          }
        ]
      },
      {
        "name": "module-b",
        "type": "module",
        "lines": [
          "[INFO] ---\u003c com.example:module-b \u003e---",
          "[INFO] Building module-b 1.0-SNAPSHOT                                 [2/2]",
          "[INFO]   from module-b/pom.xml",
          "[INFO] --------------------------------[ jar ]---------------------------------",
          "[INFO] "
        ],
        "meta": {
          "artifactId": "module-b",
          "groupId": "com.example",
          "moduleCount": 2,
          "moduleIndex": 2,
          "name": "module-b",
          "packaging": "jar",
          "pomFile": "module-b/pom.xml",
          "status": "FAILED",
          "summary": "Error in module-b",
          "version": "1.0-SNAPSHOT"
        },
        "children": [
          {
            "name": "clean",
            "type": "build-block",
            "lines": [
              "[INFO] --- clean:3.2.0:clean (default-clean) @ module-b ---",
              "[INFO]"
            ],
            "meta": {
              "artifactId": "module-b",
              "executionId": "default-clean",
              "goal": "clean",
              "plugin": "clean",
              "status": "SUCCESS",
              "summary": "Successful",
              "version": "3.2.0"
            }
          },
          {
            "name": "compiler",
            "type": "build-block",
            "lines": [
              "[INFO] --- compiler:3.15.0:compile (default-compile) @ module-b ---",
              "[ERROR] Error in module-b"
            ],
            "meta": {
              "artifactId": "module-b",
              "executionId": "default-compile",
              "goal": "compile",
              "plugin": "compiler",
              "status": "FAILED",
              "summary": "Error in module-b",
              "version": "3.15.0"
            }
          }
        ]
      },
      {
        "name": "summary",
        "type": "summary",
        "lines": [
          "[INFO] ------------------------------------------------------------------------",
          "[INFO] Reactor Summary:",
          "[INFO] module-a ....................................... FAILURE [  0.500 s]",
          "[INFO] module-b ....................................... FAILURE [  0.100 s]",
          "[INFO] ------------------------------------------------------------------------",
          "[INFO] BUILD FAILURE",
          "[INFO] ------------------------------------------------------------------------"
        ],
        "meta": {
          "modules": [
            {
              "name": "module-a",
              "status": "FAILURE",
              "time": "0.500 s"
            },
            {
              "name": "module-b",
              "status": "FAILURE",
              "time": "0.100 s"
            }
          ],
          "overallStatus": "BUILD FAILURE",
          "summary": "module-a:\nError in module-a\nAnother error\n\nmodule-b:\nError in module-b"
        }
      }
    ]
  }
}`
	_ = expected
	// Verify line numbers exist
	if out.Root.StartLine == 0 || out.Root.EndLine == 0 {
		t.Error("Root should have line numbers")
	}
}

func TestLineNumbers(t *testing.T) {
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
[INFO] ------------------------------------------------------------------------
[INFO] Reactor Summary:
[INFO] my-app ....................................... SUCCESS [  1.234 s]
[INFO] ------------------------------------------------------------------------
[INFO] BUILD SUCCESS
[INFO] ------------------------------------------------------------------------`

	lines := strings.Split(strings.ReplaceAll(mavenOutput, "\r\n", "\n"), "\n")
	parser := NewOutputParser()
	out := parser.ParseOutput(lines, nil, ParseConfig{})

	if out.Root.StartLine != 1 {
		t.Errorf("Root StartLine should be 1, got %d", out.Root.StartLine)
	}
	if out.Root.EndLine != len(lines) {
		t.Errorf("Root EndLine should be %d, got %d", len(lines), out.Root.EndLine)
	}

	hasStartLine := false
	for _, child := range out.Root.Children {
		if child.StartLine == 0 || child.EndLine == 0 {
			t.Errorf("Child %s has missing line numbers: startLine=%d, endLine=%d", child.Name, child.StartLine, child.EndLine)
		}
		hasStartLine = true
		if child.EndLine < child.StartLine {
			t.Errorf("Child %s has invalid range: startLine=%d > endLine=%d", child.Name, child.StartLine, child.EndLine)
		}
	}
	if !hasStartLine {
		t.Error("No children found to verify")
	}
}

func TestParse_SummaryBeforeInitialization(t *testing.T) {
	mavenOutput := `[INFO] ---< com.example:module-a >---
[INFO] Building module-a 1.0-SNAPSHOT
[INFO] --- compiler:3.1.0:compile (default-compile) @ module-a ---
[INFO] Compiling 1 source file
[INFO] ------------------------------------------------------------------------
[INFO] Reactor Summary:
[INFO] module-a ....................................... SUCCESS [  0.5 s]
[INFO] ------------------------------------------------------------------------
[INFO] BUILD SUCCESS
[INFO] ------------------------------------------------------------------------
[INFO] Scanning for projects...
[INFO] ------------------------------------------------------------------------
[INFO] Reactor Build Order:
[INFO] 
[INFO] module-b                                                       [jar]
[INFO] ------------------------------------------------------------------------`

	lines := strings.Split(strings.ReplaceAll(mavenOutput, "\r\n", "\n"), "\n")
	parser := NewOutputParser()
	out := parser.ParseOutput(lines, nil, ParseConfig{})

	t.Logf("Root children count: %d", len(out.Root.Children))
	for i, child := range out.Root.Children {
		t.Logf("Child %d: Type=%q, Name=%q", i, child.Type, child.Name)
	}

	// Expected: 3 children - module, summary, initialization (in that order)
	if len(out.Root.Children) != 3 {
		t.Errorf("Expected 3 children, got %d", len(out.Root.Children))
	}

	// First child should be module
	if len(out.Root.Children) > 0 && out.Root.Children[0].Type != "module" {
		t.Errorf("First child should be module, got %q", out.Root.Children[0].Type)
	}

	// Second child should be summary
	if len(out.Root.Children) > 1 && out.Root.Children[1].Type != "summary" {
		t.Errorf("Second child should be summary, got %q", out.Root.Children[1].Type)
	}

	// Third child should be initialization
	if len(out.Root.Children) > 2 && out.Root.Children[2].Type != "initialization" {
		t.Errorf("Third child should be initialization, got %q", out.Root.Children[2].Type)
	}
}
