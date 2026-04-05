package maven

// Output-related structures (LogBlock, ModuleBlock, SummaryBlock)

type LogLevel string

const (
	Info  LogLevel = "INFO"
	Warn  LogLevel = "WARN"
	Error LogLevel = "ERROR"
	Debug LogLevel = "DEBUG"
	None  LogLevel = "NONE" // For no prefix
)

type LogBlock struct {
	Level LogLevel
	Lines []string
}

type ModuleBlock struct {
	Name      string
	Artifact  string
	Packaging string
	Index     int
	Logs      []LogBlock
	Errors    []string
	Status    string // SUCCESS | FAILURE | SKIPPED
}

type ModuleSummary struct {
	Name     string
	Status   string
	Duration string
}

type SummaryBlock struct {
	Modules     []ModuleSummary
	FinalStatus string
	TotalTime   string
}
