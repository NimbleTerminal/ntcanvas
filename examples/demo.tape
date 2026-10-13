# Run from the repository root with task record-demo.
# ntrecord renders text and Kitty graphics without a browser or GUI terminal.
Output dist/viewports.gif
Set Shell /bin/sh
Env ENV ""
Env PS1 "demo> "
Env NO_COLOR ""
Env COLORTERM "truecolor"
Env CLICOLOR_FORCE "1"
Set FontFamily "Go Mono"
Set FontSize 18
Set Columns 80
Set Rows 24
Set Theme "Catppuccin Mocha"
Set Padding 20
Set Margin 12
Set MarginFill "#11111b"
Set WindowBar Colorful
Set BorderRadius 10
Set Framerate 20
Set TypingSpeed 100ms
Set WaitTimeout 10s

# Launch prebuilt examples while hidden; never record a local shell prompt.
Hide
Type@1ms "./bin/ntcanvas-viewports"
Enter
Wait+Screen /Viewports: focus=left left=\(0,0\)/
Show
Sleep 2s

# Pan only the focused viewport, then switch and pan the other one.
Right@200ms 4
Down@200ms 2
Wait+Screen /left=\(4,2\) right=\(12,4\)/
Sleep 1s
Enter
Wait+Screen /Viewports: focus=right/
Right@200ms 3
Down
Wait+Screen /left=\(4,2\) right=\(15,5\)/
Sleep 1s

# One edit is visible through both independent views of the shared grid.
Type "e"
Wait+Screen /marks=1/
Screenshot dist/viewports.png
Sleep 3s
Hide
Type "q"
Wait+Screen /demo> /

# Also capture the grid comparison for local inspection, outside the GIF.
Type@1ms "./bin/ntcanvas-grids"
Enter
Wait+Screen /SingleColumn replaces wide\/combining runes/
Screenshot dist/grids.png
Type "q"
Wait+Screen /demo> /
