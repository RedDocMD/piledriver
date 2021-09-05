#!/bin/bash

# This script changes some files in the diff_dir
# Structure:
# - a.dat
# - b.dat
# - c
#   |- e.dat
#   |- f.dat
# - d

BASE=${PWD}/data/diff_dir

# Modify a.dat
mv ${BASE}/a.dat /tmp/a.dat
dd if=/dev/random of=${BASE}/a.dat bs=1k count=10 &> /dev/null

# Remove e.dat
mv ${BASE}/c/e.dat /tmp/e.dat

# Rename f.dat to g.dat
mv ${BASE}/c/f.dat ${BASE}/c/g.dat

# Create h.dat
cp ${BASE}/a.dat ${BASE}/c/h.dat

# Replace dir d with file d
rmdir ${BASE}/d
touch ${BASE}/d

# Final structure
# - a.dat (modified)
# - b.dat
# - c
#   |- g.dat
#   |- h.dat
# - d (file)

exit 0
