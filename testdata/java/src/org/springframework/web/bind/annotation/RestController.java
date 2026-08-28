package org.springframework.web.bind.annotation;

import java.lang.annotation.Retention;
import java.lang.annotation.RetentionPolicy;

/**
 * Stands in for Spring's annotation of the same name.
 *
 * <p>The recogniser couples by fully qualified name and nothing else, so a stub
 * declaring the same name is indistinguishable from the real one in bytecode —
 * which makes this fixture a test of exactly the coupling under test, without
 * putting Spring on the classpath.
 *
 * <p>The retention matches the real annotation's. Spring's stereotypes are
 * RUNTIME, the speclink annotations beside them are CLASS, so the fixture
 * exercises both annotation tables rather than only one.
 */
@Retention(RetentionPolicy.RUNTIME)
public @interface RestController {}
