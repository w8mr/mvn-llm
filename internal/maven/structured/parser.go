package structured

type Parser interface {
	Parse(lines []string, startIdx int) (*Node, int, bool)
}
