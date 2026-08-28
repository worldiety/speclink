package com.example.requirements.fun.quote;

import com.example.requirements.dec.RDecNumbering;
import speclink.Kind;
import speclink.Requirement;
import speclink.Source;
import speclink.Status;

@Requirement(
    id = "R-QUOTE-SUBMIT",
    kind = Kind.FUNCTIONAL,
    status = Status.NORMATIVE,
    title = "Quote number on submission",
    text = "On submitting an approved quote a sequential, duplicate free quote number MUST be drawn.",
    derivedFrom = RDecNumbering.class,
    sources = @Source(doc = "requirements/_sources/sales/quoteflow.md", anchor = "8-abgabe"))
public final class RQuoteSubmit {
    private RQuoteSubmit() {}
}
