# mvn-llm

A Robust, Structured Maven Build & Analysis CLI for Language Model Agents

## Why?

Most Language Model (LLM) agents struggle to automate or interpret Maven builds, especially in complex multi-module projects and CI/CD pipelines. `mvn-llm` enables LLMs to:
- Get concise, machine-readable summaries of any Maven phase or failure (for follow-up steps, dialogue, or planning)
- Parse dependency trees, errors, affected modules, and ancestry with stable, deterministic output
- Drive build/test and dependency analysis autonomously, using declarative intent and actionable JSON or one-line text summaries
- Integrate seamlessly into LLM orchestration, agent frameworks, and advanced automation

Developers, scripts, and CI tasks can also use the CLI for the same structured outputs and uniform build interface.


## What?

- Provides a predictable, parseable CLI for Maven that speaks the language of LLMs:
  - Outputs single-line, summary text ideal for agents or chat-based feedback ("SUCCESS - Tests run: 3, Failures: 0" or error root cause)
  - Offers a full, normalized JSON schema for all goal/phases (install, test, package, deps, etc.)
  - Clearly traces error locations, resume points, module ancestry, and dependency relationships
- Lets agents request high-level actions or deep introspection using simple arguments and flags
- Also works in CI, build scripts, or directly for developers seeking reliable and regression-proof summaries

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
mvn-llm <goal> [flags]
```
Where `<goal>` is any Maven phase, e.g. `install`, `test`, `package`, `compile`, or `deps`.

#### Parsing existing Maven output (advanced)

If you already have a Maven log file and want to parse it without running Maven:

```sh
mvn-llm < build.log
cat build.log | mvn-llm
```

#### Common flags

| Flag                       | Purpose                                                    |
|---------------------------|------------------------------------------------------------|
| `-o text|json`            | Agent-friendly output: summary text or structured JSON     |
| `-output-file <path>`     | Write JSON/text to file for further LLM consumption        |
| `-rf <module>`            | Resume from Maven module (LLMs can auto-retry)             |
| `-dep-ancestor <id>`      | Show ancestors for a dependency (dependency reasoning)      |
| `-dep-verbose`            | Verbose dependency info (detailed subtrees)                |
| `-project-root <dir>`     | Project root (default: `.`)                                |
| `-no-clean`               | Skip running `mvn clean` before building                   |
| `-dep-filter <expr>`      | (deps only) Filter dependencies                            |

### Output Formats

- **text (`-o text`):** Human-readable summary
- **json (`-o json`):** Simple JSON with status, summary, and failed modules
- **json-full (`-o json-full`):** Full structured JSON without line data
- **json-full-with-lines (`-o json-full-with-lines`):** Full structured JSON with lines
- **maven-output (`-o maven-output`):** Raw Maven output

### Plugin-specific Parsing

The CLI includes specialized parsers for common Maven plugins that extract enhanced metadata:

| Plugin | Alias | Extracted Fields |
|--------|-------|----------------|
| **compiler** | maven-compiler-plugin | `compilationErrors[]`, `compilationWarnings[]`, `sourceFiles`, `compilerArgs[]`, `sourceVersion`, `targetVersion`, `incremental` |
| **surefire** | maven-surefire-plugin | `testResults{runs,failures,errors,skipped}`, `testFailures[]`, `testClass`, `provider` |
| **failsafe** | maven-failsafe-plugin | `testResults{runs,failures,errors,skipped}`, `testFailures[]`, `testClass` |
| **jar** | maven-jar-plugin | `jarFile`, `manifest` |
| **war** | maven-war-plugin | `overlay`, `webXml` |
| **ear** | maven-ear-plugin | `includes[]`, `excludes[]` |
| **install** | maven-install-plugin | `artifact`, `path` |
| **deploy** | maven-deploy-plugin | `artifact`, `path` |
| **resources** | maven-resources-plugin | `skipped`, `encoding`, `resourceType`, `source`, `target` |
| **source** | maven-source-plugin | `sourceJar` |
| **clean** | maven-clean-plugin | `deleted` |

Each plugin block in the JSON output includes these fields when present in the Maven output.

### Examples for LLMs

- **Get a one-line summary for intent/correction:**
  ```sh
  mvn-llm install -o text
  # SUCCESS - All modules built successfully.
  # or:
  # BUILD FAILURE (module: module-a) at src/BadClass.java:42 | cannot find symbol SomeType
  ```

- **Output a full, machine-parsable dependency tree:**
  ```sh
  mvn-llm deps -o json
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

- **Answer the "why" behind a dependency (provenance/reasoning):**
  ```sh
  mvn-llm deps -dep-ancestor junit:junit -o text
  # junit:junit
  #     |
  #     +- ...
  ```

- **Summarize a test failure for agent correction:**
  ```sh
  mvn-llm test -o text
  # TEST_FAILURE (module: module-a) at com.example.CalculatorTest.testFail:9 | This test always fails expected:<0> but was:<1>
  ```

---

## Maintainer notes
- Add new test fixtures by running real Maven builds in `testdata/` and capturing output for regression testing.
- To re-enable snapshot artifact tracking, tweak `.gitignore` as needed.
- Contributions and bug reports welcome!
