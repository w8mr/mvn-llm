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

	jsonBytes, err := json.MarshalIndent(out, "", "  ")
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
              "summary": "Successful: system modules path not set in conjunction with -source 11",
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
              "summary": "Successful: Building jar: target/my-app.jar",
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
          "overallStatus": "BUILD SUCCESS"
        }
      }
    ]
  }
}`
	if string(jsonBytes) != expected {
		t.Errorf("Expected:\n%s\nGot:\n%s", expected, string(jsonBytes))
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

	jsonBytes, err := json.MarshalIndent(out, "", "  ")
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
              "summary": "Failure: Error in module-a",
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
              "summary": "Failure: Another error",
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
              "summary": "Failure: Error in module-b",
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
          "overallStatus": "BUILD FAILURE"
        }
      }
    ]
  }
}`
	if string(jsonBytes) != expected {
		t.Errorf("Expected:\n%s\nGot:\n%s", expected, string(jsonBytes))
	}
}
