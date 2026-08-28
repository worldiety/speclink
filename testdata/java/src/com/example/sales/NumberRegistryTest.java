package com.example.sales;

import com.example.requirements.dec.RDecNumbering;
import speclink.Verifies;

/**
 * Hand written stand-in for a JUnit test class.
 *
 * <p>The fixture declares no JUnit, because nothing here depends on it: the
 * claim is read from the annotation in the bytecode and the result from the
 * report a build writes. What a real project has instead of this is the same
 * class with @Test on it.
 */
public class NumberRegistryTest {

    @Verifies(RDecNumbering.class)
    public void nextIsSequentialAndNeverRepeats() {
        NumberRegistry registry = new NumberRegistry();
        String first = registry.next();
        String second = registry.next();
        if (first.equals(second)) {
            throw new AssertionError("a number was drawn twice");
        }
    }
}
