oq
---

_oq_ is a command line utility that takes text input from other commands, files or stdin and renders it
it in 3D to get a quick birds eye view of the metrics.

### Examlpe
For example the following command will poll the output from the Unix `ps` command:
``` 
oq view -c "ps aux"
```
   
And will then detect the columns and allow you to click on them to view the different values and inspect
each cube on the plane to see the details of it:

![screenshot](https://raw.githubusercontent.com/cove/oview/master/screenshot-anim.gif)

### Usage

```
Usage:
  oq view [flags]

Flags:
  -c, --command string   Command to run to get data from
  -f, --file string      Load data from file or use '-' to read from stdin
  -h, --help             help for view
  -i, --interval int     Refresh data interval in seconds (default 5)
  -p, --pause            Start up with rotation paused to improve performance
      --profile          Profile CPU and memory usage
  -r, --rotations int    How many seconds each rotation takes (default 32)
  -s, --size int         Size of cube plane (default 20)
  -w, --wireframe        Render cubes as wireframes to improve performance

Global Flags:
      --config string   config file (default is $HOME/.oq.yaml)
```
