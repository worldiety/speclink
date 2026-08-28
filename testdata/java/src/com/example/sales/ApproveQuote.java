package com.example.sales;

import com.example.requirements.fun.quote.RQuoteApprove;
import speclink.Satisfies;

/** Approves a quote, including legal sign-off. */
@Satisfies(RQuoteApprove.class)
public interface ApproveQuote {
    boolean approve(String quoteId, boolean legalApproved);
}
