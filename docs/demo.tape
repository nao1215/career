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

# List the available templates.
Type "career templates"
Sleep 300ms
Enter
Sleep 2500ms

Hide
Type "clear"
Enter
Show
Sleep 400ms

# Generate an English CV from a YAML file.
Type "career generate examples/cv.yaml --template cv --output cv.pdf"
Sleep 300ms
Enter
Sleep 1800ms

# Generate the Japanese 履歴書 and 職務経歴書 from one file.
Type "career generate examples/resume.yaml --template japanese-resume"
Sleep 300ms
Enter
Sleep 1500ms
Type "career generate examples/resume.yaml --template career-history"
Sleep 300ms
Enter
Sleep 1800ms

Type "ls *.pdf"
Sleep 300ms
Enter
Sleep 2500ms
