oq
---

_oq_ is a command line utility that takes text input from other command line utilities and renders it
on a 3D to get a quick birds eye view of the output from the command.

### Examlpe
For example the following command will poll the output from the Unix `ps` command:
``` 
oq view -c "ps aux"
```
   
And will then detect the columns and allow you to click on them to view the different values:

![screenshot](https://raw.githubusercontent.com/cove/oq/master/screenshot.png)

You can also specifc files to poll and stdin.
