package speclink;

/** Where a requirement came from. Exactly one of doc and extern applies. */
public @interface Source {
    String doc() default "";
    String anchor() default "";
    String extern() default "";
    String note() default "";
}
