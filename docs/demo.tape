# Demo for the README. Regenerate with: make demo (requires vhs).
# Renders image/demo.gif.

Output image/demo.gif

Set Shell "bash"
Set FontSize 18
Set Width 1100
Set Height 760
Set Padding 16
Set TypingSpeed 35ms
Set Theme "Catppuccin Mocha"

Hide
Type "export PATH=$PWD:$PATH PS1='$ '"
Enter
Type "clear"
Enter
Show

Sleep 600ms

# Start a new resume from a template.
Type "career init"
Sleep 300ms
Enter
Sleep 1800ms

# List the available templates.
Type "career templates"
Sleep 300ms
Enter
Sleep 2200ms

Hide
Type "clear"
Enter
Show
Sleep 400ms

# One YAML file renders all three documents in the right language.
Type "career generate examples/resume.yaml --template cv --output cv.pdf"
Sleep 300ms
Enter
Sleep 1500ms
Type "career generate examples/resume.yaml --template japanese-resume"
Sleep 300ms
Enter
Sleep 1500ms
Type "career generate examples/resume.yaml --template work-history"
Sleep 300ms
Enter
Sleep 1800ms

Type "ls *.pdf"
Sleep 300ms
Enter
Sleep 2500ms
