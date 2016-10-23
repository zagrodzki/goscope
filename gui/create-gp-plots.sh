#!/bin/bash

OPTS="set lmargin 0; set rmargin 0; set tmargin 0; set bmargin 0; unset border; unset key; unset xtics; unset ytics; set terminal png size 800,600; set samples 1000; set yrange [-1:1];"
LW="linewidth 3"

gnuplot -e "$OPTS plot [0:12.566370614359172] sin(x) $LW" > sin-gp.png
gnuplot -e "$OPTS plot [0:1] 0 $LW" > zero-gp.png
gnuplot -e "$OPTS plot [-1:5] x < 1 ? x : x < 3 ? 2 - x : x - 4 $LW" > triangle-gp.png
gnuplot -e "$OPTS plot [0:4] x <= 1 ? 1 : x <= 2 ? -1 : x <= 3 ? 1 : -1 $LW" > square-gp.png
