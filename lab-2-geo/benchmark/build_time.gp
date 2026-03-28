set terminal pngcairo size 1280,720
set output 'build_time.png'

set title 'KD-tree build time'
set xlabel 'Number of points'
set ylabel 'Time per build, ms'
set grid
set key left top
set logscale x 10
set logscale y 10

plot \
  'build.dat' using 1:2 with linespoints lw 2 pt 7 title 'KD-tree build', \
  'build.dat' using 1:2:(sprintf("%.1f", $2)) with labels offset 0,1 notitle