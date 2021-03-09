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

//TreeElement contains the structure for a printable tree.
type TreeElement struct {
	Headline    string
	Description string
	Childs      []*TreeElement
}

//PrintTrees turns the given TreeElements into a formated tree as strings.Builder.
func PrintTrees(nodes []TreeElement) strings.Builder {
	output := strings.Builder{}
	for _, node := range nodes {
		printNode(&node, "", &output, true, true)
		output.WriteString("\n")
	}
	return output
}

func printNode(node *TreeElement, preFix string, output *strings.Builder, isLast bool, rootLevel bool) {
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
		if node.Description != "" {
			fmt.Fprintf(output, "%s", formatDescription(preFix, itemFormatDescription, node.Description, isLast))
		}
		preFix = addEmptySpaceOrContinueItem(preFix, isLast)
	}

	for i, subNodes := range node.Childs {
		printNode(subNodes, preFix, output, i == len(node.Childs)-1, false)
	}
}

func formatDescription(preFix string, itemFormatDescription string, nodeDescription string, isLast bool) string {
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
				linesFixedLength = append(linesFixedLength, line[i:endIndex])
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
