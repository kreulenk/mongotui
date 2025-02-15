# Where should we write the GIF?
Output ./docs/demo/demo.gif

# Set up a 3600x1800 terminal with 46px font.
Set FontSize 46
Set Width 3600
Set Height 1800

Type@300ms "mongotui raspberrypi"
Sleep 500ms
Enter
Sleep 5s

#Show first set of example docs from admin database
Type@1500ms "llhh"
Sleep 1s

# Navigate to vendingMachine database
Type@1s "jjjj"
Sleep 1s

# Navigate into the sodas collection
Type@1s "ll"
Sleep 5s

# Search the collection
Type@500ms `k"cost": 2`
Sleep 2s
Enter
Sleep 3

# View a document

Type@1s "jv"
Sleep 6s
Type "b"
Sleep 3s

# Enter the users collection
Type@1500ms "hjl"

# Arrow around within the users collection
Type@1s "jjjjjjkkkkkk"

# Edit a document with vim
Type "e"
Sleep 2s
Type@1s "/100"
Enter
Sleep 1s
Type "a"
Sleep 500ms
Backspace
Sleep 500ms
Type "2"
Sleep 2s
Escape
Type@1s ":wq"
Enter
Sleep 4s

# Type in a bad query
Type@500ms "kbadQuery"
Sleep 1s
Enter
Sleep 3s
Enter
Sleep 4s