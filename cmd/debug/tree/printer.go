package tree

import (
	"fmt"
	"strings"

	lsv1alpha1 "github.com/gardener/landscaper/apis/core/v1alpha1"
)

const (
	newLine      = "\n"
	emptySpace   = "    "
	middleItem   = "├── "
	continueItem = "│   "
	lastItem     = "└── "
)

type TreeElement struct {
	Headline    string
	Description string
	Childs      []TreeElement
}

func PrintTree(nodes []TreeElement) strings.Builder {
	output := strings.Builder{}
	for _, node := range nodes {
		printNode(node, "", &output, true)
	}
	return output
}

func printNode(node TreeElement, preFix string, output *strings.Builder, isLast bool) {
	itemFormat := middleItem
	if isLast {
		itemFormat = lastItem
	}
	spaces := preFix

	fmt.Fprintf(output, "%s%s%s", spaces, itemFormat, node.Headline)
	spaces = addEmptySpaceOrContinueItem(preFix, isLast)
	if node.Description != "" {
		fmt.Fprintf(output, "%s%s%s", spaces, itemFormat, node.Description)
	}

	for i, subNodes := range node.Childs {
		printNode(subNodes, spaces, output, i == len(node.Childs)-1)
	}
}

func formatEmptySpaces(depth int) string {
	emptySpaces := ""
	for i := 0; i < depth; i++ {
		emptySpaces += emptySpace
	}
	return emptySpaces
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

func printErrorIfNecessary(err *lsv1alpha1.Error, preFix string, output *strings.Builder) {
	if err != nil {
		fmt.Fprintf(output, "%s%s", preFix, err.Message)
	}
}
