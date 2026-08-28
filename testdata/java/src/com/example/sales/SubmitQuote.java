package com.example.sales;

import com.example.requirements.fun.quote.RQuoteSubmit;
import speclink.Satisfies;

/** Submits an approved quote and draws its number. */
@Satisfies(RQuoteSubmit.class)
public interface SubmitQuote {
    String submit(String quoteId);
}
