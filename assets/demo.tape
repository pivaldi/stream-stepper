# demo.tape

Output assets/demo.gif

Set Shell zsh

# Set up the visual style
Set FontSize 16
Set FontFamily "DejaVu Sans Mono"
Set Width 900
Set Height 900
Set TypingSpeed 0.08
# Set Theme "Dracula"
Set Padding 20

# 1. Run tree in the terminal
Type "ERROR=true go run . --steps 7 ./examples/deploy.sh"
Enter
Sleep 10s
