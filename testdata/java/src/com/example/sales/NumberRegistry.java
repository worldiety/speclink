package com.example.sales;

import com.example.requirements.dec.RDecNumbering;
import org.springframework.stereotype.Service;
import speclink.Satisfies;

/** Hands out gapless business numbers. */
@Service
public class NumberRegistry {

    private long last;

    /** Draws the next number. Sequential and never repeated. */
    @Satisfies(RDecNumbering.class)
    public String next() {
        last = last + 1;
        return "Q-" + last;
    }

    /** Not an operation anybody asked for; it exists to keep the field honest. */
    private long peek() {
        return last;
    }
}
