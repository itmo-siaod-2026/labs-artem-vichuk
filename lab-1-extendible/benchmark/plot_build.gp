set terminal pngcairo size 1280,720
set output 'ext_build.png'

set title 'Extendible hashing: Build'
set xlabel 'Number of keys'
set ylabel 'Time per item, ms'
set grid
set key left top
set logscale x 10

plot \
  'ext_build.dat' using 1:2 with linespoints lw 2 pt 7 title 'limit=32', \
  'ext_build.dat' using 1:2:(sprintf("%.6f", $2)) with labels offset 0,1 notitle, \
  'ext_build.dat' using 1:3 with linespoints lw 2 pt 5 title 'limit=64', \
  'ext_build.dat' using 1:3:(sprintf("%.6f", $3)) with labels offset 0,-1 notitle, \
  'ext_build.dat' using 1:4 with linespoints lw 2 pt 9 title 'limit=128', \
  'ext_build.dat' using 1:4:(sprintf("%.6f", $4)) with labels offset 0,1.8 notitle