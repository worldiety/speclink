package com.example.sales;

import com.example.requirements.fun.quote.RQuoteApprove;
import com.example.requirements.fun.quote.RQuoteSubmit;
import speclink.Verifies;

/** Hand written stand-in for a JUnit test class; see NumberRegistryTest. */
public class QuoteControllerTest {

    @Verifies(RQuoteSubmit.class)
    public void submitDrawsANumber() {
        NumberRegistry registry = new NumberRegistry();
        QuoteController controller = new QuoteController(id -> registry.next(), (id, ok) -> ok);
        if (controller.submit("q1").isEmpty()) {
            throw new AssertionError("no number was drawn");
        }
    }

    @Verifies(RQuoteApprove.class)
    public void approvalNeedsLegalSignOff() {
        QuoteController controller = new QuoteController(id -> "Q-1", (id, ok) -> ok);
        if (controller.approve("q1", false)) {
            throw new AssertionError("a quote passed without legal sign-off");
        }
    }
}
