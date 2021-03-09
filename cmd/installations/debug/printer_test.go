package tree

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPrinter(t *testing.T) {
	treesToPrint := []TreeElement{
		{
			Headline:    "Headline1",
			Description: "Description1",
			Childs: []*TreeElement{
				{
					Headline:    "Headline1-1",
					Description: "Description1-1",
					Childs: []*TreeElement{
						{
							Headline:    "Headline1-1-1",
							Description: "Description1-1-1",
							Childs:      []*TreeElement{},
						},
						{
							Headline:    "Headline1-1-2",
							Description: "Description1-1-2",
							Childs:      []*TreeElement{},
						},
					},
				},
				{
					Headline:    "Headline1-2",
					Description: "Description1-2",
					Childs: []*TreeElement{
						{
							Headline:    "Headline1-2-1",
							Description: "Description1-2-1",
							Childs:      []*TreeElement{},
						},
					},
				},
			},
		},
		{
			Headline:    "Headline2",
			Description: "Description2",
			Childs: []*TreeElement{
				{
					Headline:    "Headline2-1",
					Description: "Description2-1",
					Childs: []*TreeElement{
						{
							Headline:    "Headline2-1-1",
							Description: "Description2-1-1",
							Childs:      []*TreeElement{},
						},
						{
							Headline:    "Headline2-1-2",
							Description: "Description2-1-2",
							Childs:      []*TreeElement{},
						},
						{
							Headline:    "Headline2-1-3",
							Description: "Description2-1-3",
							Childs:      []*TreeElement{},
						},
					},
				},
				{
					Headline:    "Headline2-2",
					Description: "Description2-2",
					Childs: []*TreeElement{
						{
							Headline:    "Headline2-2-1",
							Description: "Description1-2-1",
							Childs:      []*TreeElement{},
						},
						{
							Headline:    "Headline2-2-2",
							Description: "",
							Childs:      []*TreeElement{},
						},
					},
				},
			},
		},
	}

	treesAsStringBuilder := PrintTrees(treesToPrint)
	expectedTreeString := `Headline1
    Description1
    ├── Headline1-1
    │   Description1-1
    │   ├── Headline1-1-1
    │   │   Description1-1-1
    │   └── Headline1-1-2
    │       Description1-1-2
    └── Headline1-2
        Description1-2
        └── Headline1-2-1
            Description1-2-1

Headline2
    Description2
    ├── Headline2-1
    │   Description2-1
    │   ├── Headline2-1-1
    │   │   Description2-1-1
    │   ├── Headline2-1-2
    │   │   Description2-1-2
    │   └── Headline2-1-3
    │       Description2-1-3
    └── Headline2-2
        Description2-2
        ├── Headline2-2-1
        │   Description1-2-1
        └── Headline2-2-2

`
	t.Run("Correct printing of nested tree structure", func(t *testing.T) {
		assert.Equal(t, expectedTreeString, treesAsStringBuilder.String())
	})
}
