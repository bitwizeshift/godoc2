package highlight

import "github.com/alecthomas/chroma/v3"

// The highlighting palette follows the Go documentation colour scheme: Go
// blue for keywords, neutral text for names, and muted tones for literals.
var (
	lightStyle = chroma.MustNewStyle("godoc", chroma.StyleEntries{
		chroma.Background:       "#202224 bg:#f5f7f9",
		chroma.Text:             "#202224",
		chroma.Comment:          "italic #6e7781",
		chroma.CommentPreproc:   "#6e7781",
		chroma.Keyword:          "#007d9c",
		chroma.KeywordType:      "#007d9c",
		chroma.KeywordConstant:  "#007d9c",
		chroma.Name:             "#202224",
		chroma.NameFunction:     "#202224",
		chroma.NameBuiltin:      "#007d9c",
		chroma.LiteralString:    "#2f6f44",
		chroma.LiteralNumber:    "#8a4a00",
		chroma.Operator:         "#202224",
		chroma.Punctuation:      "#202224",
		chroma.LineNumbers:      "#8a8f94",
		chroma.LineNumbersTable: "#8a8f94",
		chroma.LineHighlight:    "bg:#e6f4f8",
		chroma.GenericDeleted:   "#b31d28",
		chroma.GenericInserted:  "#22863a",
		chroma.GenericEmph:      "italic",
		chroma.GenericStrong:    "bold",
		chroma.Error:            "#b31d28",
	})

	darkStyle = chroma.MustNewStyle("godoc-dark", chroma.StyleEntries{
		chroma.Background:       "#e6e6e6 bg:#2b2d30",
		chroma.Text:             "#e6e6e6",
		chroma.Comment:          "italic #9aa0a6",
		chroma.CommentPreproc:   "#9aa0a6",
		chroma.Keyword:          "#5dc9e2",
		chroma.KeywordType:      "#5dc9e2",
		chroma.KeywordConstant:  "#5dc9e2",
		chroma.Name:             "#e6e6e6",
		chroma.NameFunction:     "#e6e6e6",
		chroma.NameBuiltin:      "#5dc9e2",
		chroma.LiteralString:    "#9ccc9c",
		chroma.LiteralNumber:    "#e0b070",
		chroma.Operator:         "#e6e6e6",
		chroma.Punctuation:      "#e6e6e6",
		chroma.LineNumbers:      "#8a8f94",
		chroma.LineNumbersTable: "#8a8f94",
		chroma.LineHighlight:    "bg:#1f3c47",
		chroma.GenericDeleted:   "#ff7b72",
		chroma.GenericInserted:  "#7ee787",
		chroma.GenericEmph:      "italic",
		chroma.GenericStrong:    "bold",
		chroma.Error:            "#ff7b72",
	})
)
