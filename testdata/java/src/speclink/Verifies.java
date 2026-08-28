package speclink;

/**
 * Names the requirements a test demonstrates.
 *
 * <p>Unlike the Go side, which writes a line when it runs, this is only a
 * claim: the annotation says a test was written for something, and a claim is
 * not evidence. The evidence comes from the test report a build writes anyway —
 * Surefire, Failsafe and Gradle all record every test and its result — which is
 * why there is nothing to run here and no library to depend on.
 *
 * <p>It sits on the method, so it cannot be put in a branch that never
 * executes. The Go form has to be placed deliberately at the end of a test for
 * the same guarantee.
 */
public @interface Verifies {
    Class<?>[] value();
}
