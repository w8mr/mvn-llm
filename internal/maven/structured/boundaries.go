package structured

// ParseUntilNextBlock is the core boundary detection helper used by all parsers.
//
// Purpose:
// - Reads lines from startIdx until ANY parser's start marker is encountered
// - Stops BEFORE the next block's start marker (whether sibling or potential child)
// - Allows the outer OutputParser to encounter child blocks and parse them separately
//
// Parameters:
// - lines: The full log output
// - startIdx: Position to start reading (typically after the current block's header marker)
// - parsers: All available parsers (used to detect any block start)
// - currentNodeType: Type of node being parsed (currently unused but kept for potential future optimization)
//
// Returns:
// - []string: Lines consumed (from startIdx to just before next block)
// - int: Number of lines consumed
//
// Design rationale:
//   - By stopping at ANY marker (not just siblings), we allow child blocks to be parsed
//     as separate nodes by the main OutputParser loop
//   - The main loop uses bubble-up insertion logic to place children in the correct parent
//   - StartMarker should be strict (strong markers only) to avoid false positives in boundary detection
//   - This approach keeps block extraction simple while enabling rich hierarchical structures
//
// Example: When parsing a module block that contains build blocks:
//  1. Module parser calls ParseUntilNextBlock after its header
//  2. Function stops when it hits the first build block marker
//  3. Module's lines include only the header + content before first child
//  4. Main loop then parses the build blocks and inserts them as module children
func ParseUntilNextBlock(lines []string, startIdx int, parsers []Parser, currentNodeType string) ([]string, int) {
	end := startIdx
	for end < len(lines) {
		for _, parser := range parsers {
			if ok, _ := parser.StartMarker(lines, end); ok {
				// Stop at any block marker - let the outer parser handle it
				return lines[startIdx:end], end - startIdx
			}
		}
		end++
	}
	return lines[startIdx:], len(lines) - startIdx
}
