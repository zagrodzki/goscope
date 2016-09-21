#!/bin/sh
find . -name '*.go' -print0 | xargs -0 awk '
  FNR==1 && $0 !~ "^//  Copyright .... The goscope Authors" {
          missing[FILENAME] = 1
  }
  END {
          if (length(missing) > 0) {
                  print "Files with missing license note:"
                  for (f in missing) {
                          print f
                  }
                  print "Copy the contents of docs/license.snippet to the beginning of the file."
                  exit 1
          }
  }'
