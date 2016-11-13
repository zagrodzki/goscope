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

gnuplot -e "$OPTS f(x)=abs(0.1/x); g(x)=-0.03-0.25*(2.718**(2*-x**2)); set xrange [-10:10]; plot '+' using 1:(f(\$1)):(g(\$1)) with filledcurves closed" > spike-int-gp.png
gnuplot -e "$OPTS f(x)=x<-1.05 ? -0.5-0.02/(x+1) : x<1 ? 0.52+(x/2.5)**2 : -0.5+0.02/(x-0.95); g(x)=x<-1 ? -0.5+0.03/(x+0.9) : x<0.95 ? 0.48-(x/2.5)**2 : -0.5-0.03/(x-0.85); set xrange [-3:3]; plot '+' using 1:(f(\$1)):(g(\$1)) with filledcurves closed" > square-int-gp.png
