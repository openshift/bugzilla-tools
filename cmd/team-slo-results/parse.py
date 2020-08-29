#!/usr/bin/env python

import json
import sys

data = json.load(sys.stdin)

for team in sorted(data.keys()):
    print("***", team, "***")
    sloResults = data[team]
    members = 0
    if "members" in sloResults:
        members = int(sloResults["members"])
    resDict = {}
    for result in sloResults["results"]:
        resDict[result["name"]] = result
    for sloName in sorted(resDict.keys()):
        sloResult = resDict[sloName]
        current = int(sloResult["current"])
        if members == 0 or sloName == "urgents":
            print("  ", sloName, "\t*"+str(current))
        else:
            print("  ", sloName, "\t", current/members)
print("* indicates value is absolute, instead of 'per team member'")
