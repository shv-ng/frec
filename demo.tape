Output frec-demo.gif
Set FontSize 18
Set Width 900
Set Height 600
Set Padding 20
Set Theme "Catppuccin Mocha"

# Clean state so demo is deterministic
Hide
Type "rm -rf ~/.local/share/frec/projects.tsv"
Enter
Type "rm -rf /tmp/frec-demo"
Enter

Type "mkdir -p /tmp/frec-demo/projects"
Enter
Type "cd /tmp/frec-demo/projects"
Enter
Type "touch api cli dotfiles infra notes"
Enter
Show

Enter
Type "fd . -t f | frec sync projects"
Enter
Sleep 1

Enter
Type "frec list projects"
Enter
Sleep 1

Enter
Type "frec add projects cli"
Enter
Type "frec add projects cli"
Enter
Type "frec add projects dotfiles"
Enter
Type "frec add projects cli"
Enter
Type "frec add projects infra"
Enter
Type "frec star projects dotfiles"
Enter
Enter
Type "frec list projects"
Enter
Sleep 2

