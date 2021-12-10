package tree

import (
	"fmt"
	"strings"
)

const (
	emptySpace   = "    "
	middleItem   = "├── "
	continueItem = "│   "
	lastItem     = "└── "
)

const terminalWidth = 120

//PrintableTreeNode contains the structure for a printable tree.
type PrintableTreeNode struct {
	Headline    string
	WideData    string // will be displayed in '-o wide' mode
	Description string
	Childs      []*PrintableTreeNode
}

func (node *PrintableTreeNode) print(output *strings.Builder, preFix string, isLast bool, rootLevel bool) {
	itemFormatHeading := middleItem
	if isLast {
		itemFormatHeading = lastItem
	}
	if rootLevel {
		itemFormatHeading = ""
	}
	itemFormatDescription := continueItem
	if rootLevel {
		itemFormatDescription = emptySpace
	}

	if node.Headline != "" {
		fmt.Fprintf(output, "%s%s%s\n", preFix, itemFormatHeading, node.Headline)
		if node.WideData != "" {
			wData := node.WideData
			if node.Description != "" {
				// add separator if -d flag is set
				wData = fmt.Sprintf("%s%s", wData, "\n----------")
			}
			fmt.Fprintf(output, "%s", formatDescription(preFix, itemFormatDescription, wData, isLast, true))
		}
		if node.Description != "" {
			fmt.Fprintf(output, "%s", formatDescription(preFix, itemFormatDescription, node.Description, isLast, false))
		}
		preFix = addEmptySpaceOrContinueItem(preFix, isLast)
	}

	for i, subNode := range node.Childs {
		subNode.print(output, preFix, i == len(node.Childs)-1, false)
	}
}

//PrintTrees turns the given PrintableTreeNodes into a formated tree as strings.Builder.
func PrintTrees(nodes []PrintableTreeNode) strings.Builder {
	output := strings.Builder{}
	for _, node := range nodes {
		node.print(&output, "", true, true)
		output.WriteString("\n")
	}
	return output
}

func formatDescription(preFix string, itemFormatDescription string, nodeDescription string, isLast, increaseIndent bool) string {
	if isLast {
		itemFormatDescription = emptySpace
	}

	//break to long lines
	lines := strings.Split(nodeDescription, "\n")
	linesFixedLength := []string{}
	for _, line := range lines {
		if len(line) > terminalWidth {
			for i := 0; i < len(line); i += terminalWidth {
				endIndex := i + terminalWidth
				if endIndex > len(line) {
					endIndex = len(line)
				}
				if increaseIndent && i > 0 {
					// indent broken lines by two spaces
					linesFixedLength = append(linesFixedLength, fmt.Sprintf("  %s", line[i:endIndex]))
				} else {
					linesFixedLength = append(linesFixedLength, line[i:endIndex])
				}
			}
		} else {
			linesFixedLength = append(linesFixedLength, line)
		}
	}

	//add prefix and new line
	for i, line := range linesFixedLength {
		linesFixedLength[i] = fmt.Sprintf("%s%s%s\n", preFix, itemFormatDescription, line)
	}

	return strings.Join(linesFixedLength, "")
}

func addEmptySpaces(s string) string {
	return s + emptySpace
}

func addEmptySpaceOrContinueItem(s string, isLast bool) string {
	if isLast {
		return addEmptySpaces(s)
	}
	return s + continueItem
}
