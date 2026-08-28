package speclink;

/** Names the requirements a construct was written for. */
public @interface Satisfies {
    Class<?>[] value();
}
