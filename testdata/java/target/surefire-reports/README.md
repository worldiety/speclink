# Test reports

Written by hand rather than by a runner, for the same reason the compiled
classes beside them are committed: the format is what is under test, and a
fixture that needed Maven and JUnit on every machine would test the toolchain
instead of the reader.

The dialect is Surefire's, which Failsafe and Gradle both write close enough
variants of. Only the fields every runner agrees on are read.
