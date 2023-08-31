Ad-hoc Casts
============

A simple application for creating ad-hoc podcasts from directories of mp3s.

------------------
About Ad-hoc Casts

Ad-hoc Casts is an easy way to generate ad-hoc podcasts from directories of mp3s.
I often find talks on the internet that provide an mp3. I want to listen to them
on my phone but listening to individual mp3's from the internet is much more
cumbersome than listening to a podcast. For example my podcast app can play the
audio at a higher speed, automatically delete audio when it's been played,
download the file when I'm not on a cellular connection, queue up multiple files
to play etc. With Ad-hoc Casts. I can quickly download one or more mp3's to my
laptop and turn them into a podcast that can be subscribed to on my phone.

-----
Usage

The program takes three arguments. The first -url is the public base url to publish the
podcasts under. The hostname should be the hostname of your machine. The -port is the TCP port to listen on. The -dir is
the root directory where the application will look for podcast directories.

./adhoccasts -dir /Users/efarrer/Podcasts/ -url http://adhoccasts.duckdns.org:8080

------------------------------------
Directory layout for ad-hoc podcasts
Each directory is it's own podcast and any mp3s under that directory will be
treated as episodes. The title of the podcast is the portion of the directory
name before a double underscore. . The description is the portion after the
double underscore. Any single underscores will be converted to spaces. So a
podcast titled "Old Yeller" with the description of "A sad podcast about a dog"
Could be created by the following directory name:
"Old_Yeller__A_sad_podcast_about_a_dog"

------------
Code Quality

This is a quick and dirty program that's currently at the "runs on my machine"
state. It could probably use a refactor and improved unit tests. Feel free to
use it, modify it and send patches. Don't use it as a good example of Go code
(it's not).
