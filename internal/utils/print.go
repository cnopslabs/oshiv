package utils

import "github.com/fatih/color"

var FaintMagenta = color.New(color.Italic, color.FgMagenta)
var HeaderFmt = color.New(color.FgCyan, color.Underline).SprintfFunc()
var ColumnFmt = color.New(color.FgYellow).SprintfFunc()
var Yellow = color.New(color.FgYellow)
var Faint = color.New(color.Faint)
var Blue = color.New(color.FgCyan)
