set terminal pngcairo size 1280,720
set output 'perfect_build.png'

set title 'Perfect hash: build'
set xlabel 'Number of keys'
set ylabel 'Time per item, ns'
set grid
set key left top
set logscale x 10

plot \
  'perfect_build.dat' using 1:2 with linespoints lw 2 pt 7 title 'Build', \
  'perfect_build.dat' using 1:2:(sprintf("%.1f", $2)) with labels offset 0,1 notitle