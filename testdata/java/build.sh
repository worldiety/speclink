#!/bin/sh
# Recompiles the fixture.
#
# The .class files are committed, because the reader tests are about the class
# file format and would otherwise need a JDK on every machine that runs them.
# They are also the input under test: regenerating them is a deliberate act, and
# the diff of this directory is where a toolchain change becomes visible.
#
# --release 17 rather than the newest available, so the fixture exercises the
# version an ordinary project produces rather than whatever this machine has.
set -eu
cd "$(dirname "$0")"
rm -rf classes
mkdir -p classes
find src -name '*.java' | xargs javac --release 17 -d classes
echo "recompiled $(find classes -name '*.class' | wc -l | tr -d ' ') classes"
