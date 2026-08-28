package com.example.requirements.fun.quote;

import speclink.Kind;
import speclink.Requirement;
import speclink.Source;
import speclink.Status;

@Requirement(
    id = "R-QUOTE-APPROVE",
    kind = Kind.FUNCTIONAL,
    status = Status.NORMATIVE,
    title = "Approval gate",
    text = "A quote MUST pass an approval gate including legal sign-off before it can be submitted.",
    sources = @Source(doc = "requirements/_sources/sales/quoteflow.md", anchor = "9-versand"))
public final class RQuoteApprove {
    private RQuoteApprove() {}
}
