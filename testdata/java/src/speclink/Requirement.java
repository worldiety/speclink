package speclink;

/**
 * Marks a class as a requirement declaration.
 *
 * <p>A requirement is a class rather than a record or a constant because Java
 * annotations may only carry compile time constants. Only a class literal lets
 * one requirement name another and have the compiler check it — with a string
 * the reference would be unverified, which is the one thing this design will
 * not accept.
 *
 * <p>The default retention is CLASS, which is what speclink reads. There is
 * deliberately no RUNTIME retention and no library to depend on: a project
 * declares these annotations itself and speclink recognises them by name.
 */
public @interface Requirement {
    String id();
    Kind kind();
    Status status();
    String title() default "";
    String text() default "";
    String rationale() default "";
    Class<?>[] derivedFrom() default {};
    Class<?>[] supersedes() default {};
    Source[] sources() default {};
}
