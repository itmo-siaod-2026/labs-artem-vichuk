set terminal pngcairo size 1280,720
set output 'build_memory.png'

set title 'KD-tree build memory'
set xlabel 'Number of points'
set ylabel 'Allocated memory per build, MB'
set grid
set key left top
set logscale x 10
set logscale y 10

plot \
  'build.dat' using 1:3 with linespoints lw 2 pt 7 title 'KD-tree build', \
  'build.dat' using 1:3:(sprintf("%.1f", $3)) with labels offset 0,1 notitle