// Package ntcanvasgrid is ntcanvas's cell storage and drawing without its
// terminal UI: RuneGrid and GraphemeGrid, the compact CellStyle they render
// with, points and coordinate conversions, and the RasterLine and RasterLineAA
// line rasterizers.
//
// It imports neither Bubble Tea, Lip Gloss, nor bubblezone, so a program that
// only fills a grid and renders it to a string, such as a renderer or a
// command-line tool, links none of them. Package ntcanvas re-exports
// everything here under the same names, and adds the Lip Gloss styled grid and
// the Bubble Tea model; a type from either package is the same type.
package ntcanvasgrid
