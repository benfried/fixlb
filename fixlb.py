#!/opt/local/bin/python3.4

import re, os, sys
from stat import *
import distance

guide_file = "/Users/bf/Google Drive/lb.html"
tivo_dir = "/Volumes/media/Videos/Little Bear"
athome = False
scriptfilename = "/Users/bf/fix-lb.sh"

season=0
episode=0
tivo_to_plex = {}
renames = {}
unmatched = []

with open(guide_file, encoding="utf-8") as season_guide:
    for line in season_guide:
        match = re.match("Season ([0-9]+)", line)
        if match:
            season=int(match.group(1))
            episode=0
            continue
        line = re.sub("[?!]", "", line) # remove ? and !
        match = re.match('"(.+) / (.+) / (.+)"', line)
        if match:
            episode += 1
            tivo_filename = "Little Bear - ''{0}; {1}; {2}''".format(match.group(1), match.group(2), match.group(3))
            plex_filename = "Little Bear s{0}e{1} - {2}; {3}; {4}.mp4".format(season, episode, match.group(1), match.group(2), match.group(3))

            tivo_to_plex[tivo_filename] = plex_filename

if athome:
    os.chdir(tivo_dir)
    files = os.listdir()
else:
    files = []
    with open("/Users/bf/lb-ls.txt", encoding="utf-8") as listing:
        for line in listing:
            line = line.rstrip("\r\n")
            files.append(line)

with open(scriptfilename, "w") as scriptfile:
    scriptfile.write("#!/bin/sh\n")

    for f in files:
        matched = False
        for t in tivo_to_plex:
            if f.casefold().startswith(t.casefold()) :
                matched = True
                scriptfile.write('mv "{0}" "{1}"\n'.format(f, tivo_to_plex[t]))
                renames[f] = tivo_to_plex[t]
        if not matched:
            unmatched.append(f)

    scriptfile.write("\n#Below were set by a fuzzy match, so give them a quick scan\n\n")

    for u in unmatched:
        position = u.find(" (R")
        if position > -1:
                trimmed_filename = u[0:position]
                for t in tivo_to_plex:
                    if (distance.levenshtein(t, trimmed_filename) < 15):
                        # print("{0} is close to {1}\n".format(u, t))
                        renames[u] = tivo_to_plex[t]
                        scriptfile.write('mv "{0}" "{1}"\n'.format(u, renames[u]))
                        unmatched.remove(u)
        else:
            print("{0} is not an episode file\n".format(u))

    scriptfile.write("\n#Below could not be matched\n\n")

    for u in unmatched:
        scriptfile.write('# mv "{0}" \n'.format(u))

            
