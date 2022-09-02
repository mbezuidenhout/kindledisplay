# kindledisplay
Use your old kindle to display date, time or any image from a URL.

# Abstract
You will have to have a jailbroken kindle in order to upload the files.

## Tested on
Kindle 4 NT

# Usage
./pagegen [-c config.yml]

# Compiling for kindle

## Kindle 4 No Touch
env GOOS=linux GOARCH=arm GOARM=6 go build

# Config
Using the config.yml.sample file as a template create your own config.yml file.

# References
https://fnordig.de/2015/05/14/using-a-kindle-for-status-information/
https://wiki.mobileread.com/wiki/Kindle4NTHacking
https://www.mobileread.com/forums/showthread.php?t=191158
