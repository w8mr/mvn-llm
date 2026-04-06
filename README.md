# mvn-llm

A Robust, Structured Maven Build & Analysis CLI for LLMs and Developers

## Why?

Modern LLM-based agents and many developer tools often struggle to understand, reason about, and automate Maven builds—especially for complex multi-module projects and CI/CD. This tool provides a predictable and parseable interface to Maven that includes:
- Structured, reliable output for dependency trees, build/test phases, errors, and affected modules
- Usable by LLM agents, scripts, GitHub Actions, or humans who want clean summaries and actionable details
- Unified handler for all Maven phases (install, test, compile, package, etc.) and robust dependency/ancestor tracing

## What?

- Wraps Maven with a CLI that:  
  - Outputs structured machine- and human-friendly summaries for all phases/goals
  - Understands and tracks module ancestry, dependency trees, and errors precisely  
  - Detects resume points, build errors, test results, and full ancestry for any dependency or module
- Works cross-platform in CI and developer shells
- Maintains up-to-date test fixtures so outputs are trustworthy and regression-proof

## How?

### Install

#### Homebrew (Recommended - macOS & Linux)

```sh
brew tap w8mr/homebrew-tap
brew install mvn-llm
```

- This will fetch and install the latest release from the Homebrew tap.
- To upgrade later:
  ```sh
  brew upgrade mvn-llm
  ```

#### Manual Build (Advanced/Other Platforms)
If you don't use Homebrew or want the latest dev version:

```sh
git clone https://github.com/w8mr/maven-tool.git
cd maven-tool
make      # or: go build -o mvn-llm ./cmd/mvn-llm
```

### Usage

```sh
./mvn-llm <goal> [flags]
```
where `<goal>` is any Maven phase/goal, e.g. `install`, `test`, `package`, `compile`, `deps`.

#### Common command-line flags

| Flag                       | Purpose                                                    |
|---------------------------|------------------------------------------------------------|
| `-project-root <dir>`     | Project root (default: `.`)                                |
| `-no-clean`               | Skip running `mvn clean` before building                   |
| `-rf <module>`            | Resume build from this Maven module                        |
| `-o text|json`            | Choose output format (text summary or full JSON)           |
| `-output-file <path>`     | Write JSON output to file                                  |
| `-dep-filter <expr>`      | (deps only) Filter dependencies with Maven `includes` expr |
| `-dep-ancestor <id>`      | (deps only) Show ancestors for this `groupId:artifactId`   |
| `-dep-verbose`            | (deps only) Show verbose dependency information            |

### Examples

- Run build and get a summary:
  ```sh
  ./mvn-llm install
  # SUCCESS or concise error summary, module, location, etc.
  ```

- Parse and pretty-print the full dependency tree:
  ```sh
  ./mvn-llm deps -o json
  {
    "modules": [
      {
        "moduleName": "my-app-module-a",
        "root": {
          "groupId": "com.example",
          "artifactId": "module-a",
          "children": [ ... ]
        }
      }, ...
    ]
  }
  ```

- Get the ancestors (why?) of a dependency:
  ```sh
  ./mvn-llm deps -dep-ancestor junit:junit
  # junit:junit
  #     |
  #     +- ...
  ```

- Summarize a test failure in CI (for LLM agent):
  ```sh
  ./mvn-llm test -o text
  # TEST_FAILURE (module: module-a) at com.example.CalculatorTest.testFail:9 | This test always fails expected:<0> but was:<1>
  ```

### Output Summary

- **Text (`-o text`):** A one-line summary for LLMs/agents (`SUCCESS - Tests run: 2, Failures: 0, Errors: 0`, or `COMPILE_ERROR (module: foo) at path/Line | error`)
- **JSON (`-o json`):** Structured data for automation/analysis, with:
  - `status` (`SUCCESS`, `BUILD FAILURE`, `TEST_FAILURE`, ...)
  - `failedModule`, `resumeCommand`, `errors`, `failureLocation`, etc.
  - For `deps`, a full recursive tree and ancestor list.

---

## Maintainer notes
- Add new test fixtures by running real Maven builds in `testdata/` and capturing output for regression testing.
- To re-enable snapshot artifact tracking, tweak `.gitignore` as needed.
- Contributions and bug reports welcome!
