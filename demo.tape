# kspec Demo - Host Security Scanning
# Run with: vhs demo.tape

Output docs/demo.gif

Set Shell "bash"
Set FontSize 14
Set FontFamily "JetBrains Mono"
Set Width 1200
Set Height 700
Set Theme "Catppuccin Mocha"
Set Padding 20
Set BorderRadius 8
Set WindowBar Colorful
Set TypingSpeed 40ms
Set PlaybackSpeed 1

# Title
Type "# kspec - Policy-as-Code Security Scanner"
Enter
Sleep 500ms
Type "# Scanning julian.pro for TLS & Certificate Security"
Enter
Enter
Sleep 1s

# Run the scan
Type "./kspec scan host julian.pro -d policies"
Sleep 500ms
Enter

# Wait for scan to complete
Sleep 8s

# Navigate through resources
Sleep 1s
Down
Sleep 500ms
Down
Sleep 500ms

# Show certificate resource
Down
Sleep 1s

# Switch to checks panel with Tab
Tab
Sleep 500ms

# Navigate through checks
Down
Sleep 300ms
Down
Sleep 300ms
Down
Sleep 300ms

# Enter detail view
Enter
Sleep 1.5s

# Switch focus to detail panel
Tab
Sleep 500ms

# Scroll detail content
Down
Sleep 300ms
Down
Sleep 300ms
Down
Sleep 300ms

# Switch back to check list
Tab
Sleep 500ms

# Navigate to another check
Down
Sleep 300ms
Down
Sleep 500ms

# Go back
Escape
Sleep 1s

# Navigate to TLS resource
Up
Sleep 300ms
Up
Sleep 300ms
Up
Sleep 500ms

# Show TLS checks
Tab
Sleep 500ms
Down
Sleep 300ms
Down
Sleep 300ms

# Final pause
Sleep 2s

# Quit
Type "q"
Sleep 500ms
