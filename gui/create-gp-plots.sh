#!/bin/bash

OPTS="set lmargin 0; set rmargin 0; set tmargin 0; set bmargin 0; unset border; unset key; unset xtics; unset ytics; set terminal png size 800,600; set samples 1000; set yrange [-1:1];"
LW3="linewidth 3"

gnuplot -e "$OPTS plot [0:12.566370614359172] sin(x) $LW3" > sin-gp.png
gnuplot -e "$OPTS plot [0:1] 0 $LW3" > zero-gp.png
gnuplot -e "$OPTS plot [-1:5] x < 1 ? x : x < 3 ? 2 - x : x - 4 $LW3" > triangle-gp.png
gnuplot -e "$OPTS plot [0:4] x <= 1 ? 1 : x <= 2 ? -1 : x <= 3 ? 1 : -1 $LW3" > square-gp.png

LW5="linewidth 5"

gnuplot -e "$OPTS plot [0:12.566370614359172] -0.7*sin(4*x) $LW5" > sin2-gp.png
gnuplot -e "$OPTS plot [0:12.566370614359172] 0.5*sin(2*x) + 0.5*sin(3*x) $LW5" > sin-sum-gp.png
gnuplot -e "$OPTS plot [-1:5] x<1 ? 0.5*x+0.5 : x<2 ? 2-x : x<3 ? 0 : x<4 ? 3-x : 2*x-9 $LW5" > lines-gp.png
gnuplot -e "$OPTS plot [0.5:3.5] x <= 1 ? 1 : x <= 2 ? -1 : x <= 3 ? 1 : -1 $LW5" > square-short-gp.png
