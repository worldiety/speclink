package com.example.sales;

import com.example.requirements.fun.quote.RQuoteApprove;
import com.example.requirements.fun.quote.RQuoteSubmit;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RestController;
import speclink.Satisfies;

/** The web boundary of the quotation domain. */
@RestController
public class QuoteController {

    private final SubmitQuote submitQuote;
    private final ApproveQuote approveQuote;

    public QuoteController(SubmitQuote submitQuote, ApproveQuote approveQuote) {
        this.submitQuote = submitQuote;
        this.approveQuote = approveQuote;
    }

    @PostMapping("/quotes/{id}/submit")
    @Satisfies(RQuoteSubmit.class)
    public String submit(String id) {
        return submitQuote.submit(id);
    }

    @PostMapping("/quotes/{id}/approve")
    @Satisfies(RQuoteApprove.class)
    public boolean approve(String id, boolean legalApproved) {
        return approveQuote.approve(id, legalApproved);
    }

    /** Not an endpoint: no request mapping, so nothing is asked of it. */
    private String describe(String id) {
        return "quote " + id;
    }
}
