# rofi-code
Use rofi to quickly open Visual Studio Code or Codium workspaces. Since I use it, I have more time to enjoy with my friends and family.

## How to use it

Just run `rofi-code` and a popup window should appear showing the list of all your workspaces. Get the list filtered by typing text, use the arrow keys to displace the cursor to the workspace you want to open, and press enter. That's it.

If you want to customize the behaviour you could use any of the options provided using the command line.

For a list of flags and options just type run `rofi-code --help` in your terminal:

```console
foo@bar:~$ rofi-code --help
usage: rofi-code [-h|--help] [-d|--dir "<value>"] [-s|--sort (time|path|name)]
                 [-f|--full] [-o|--output] [-r|--rofi "<value>"] [-c|--code
                 "<value>"]

                 Use rofi to quickly open a VSCode or Codium workspace

Arguments:

  -h  --help    Print help information
  -d  --dir     Comma separated paths to the config directories. Default:
                ~/.config/VSCodium,~/.config/Code
  -s  --sort    How the workspaces should be sorted. Default: time
  -f  --full    Show the full path instead of collapsing the home directory to
                a tilde. Default: false
  -o  --output  Just prints the workspaces to stdout and exit. Default: false
  -r  --rofi    Command line to launch rofi. Default: rofi -dmenu -p "Open
                workspace" -no-custom
  -c  --code    Command line to launch the editor. It will try to detect codium
                or code
```
I use `rofi-code` everyday using a keystroke combination in my [i3 window manager](https://i3wm.org/) setup with the following line in `~/.config/i3/config`.

```
bindsym $mod+Shift+v exec --no-startup-id rofi-code
```

## Requirements

Of course you will need [rofi](https://github.com/davatorium/rofi) installed. 

Also you will be required to have installed the go language tools in your system as a prerequisite to build this software. Why? Because this is coded in go.

On Debian or Ubuntu systems, it should be enough to write this
```console
foo@bar:~$ sudo apt install golang
```

For any other system the [official website](https://golang.org/doc/install) is not a bad starting point.

## How to install

Run the following commands in your terminal

```console
foo@bar:~$ go get github.com/Coffelius/rofi-code
foo@bar:~$ go install github.com/Coffelius/rofi-code
```

At this point you should have the executable file at `~/go/bin/rofi-code` or wherever your environment variables `$GOHOME` points.

Feel free to alter your `$PATH` if it doesn't include the place where the binary was stored or just copy `rofi-code` binary to some sensitive place like `/usr/local/bin`




