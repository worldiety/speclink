package org.springframework.web.bind.annotation;

import java.lang.annotation.Retention;
import java.lang.annotation.RetentionPolicy;

/** Stands in for Spring's annotation of the same name. */
@Retention(RetentionPolicy.RUNTIME)
public @interface GetMapping { String value() default ""; }
