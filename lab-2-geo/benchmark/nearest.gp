set terminal pngcairo size 1280,720
set output 'nearest.png'

set title 'Nearest neighbor search'
set xlabel 'Number of points'
set ylabel 'Time per query, ms'
set grid
set key left top
set logscale x 10
set logscale y 10

plot \
  'nearest.dat' using 1:2 with linespoints lw 2 pt 7 title 'KD-tree', \
  'nearest.dat' using 1:2:(sprintf("%.3f", $2)) with labels offset 0,1 notitle, \
  'nearest.dat' using 1:5 with linespoints lw 2 pt 5 title 'Linear', \
  'nearest.dat' using 1:5:(sprintf("%.3f", $5)) with labels offset 0,-1 notitle