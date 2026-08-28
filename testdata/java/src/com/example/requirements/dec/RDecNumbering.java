package com.example.requirements.dec;

import speclink.Kind;
import speclink.Requirement;
import speclink.Source;
import speclink.Status;

@Requirement(
    id = "R-DEC-NUMBERING",
    kind = Kind.DECISION,
    status = Status.NORMATIVE,
    title = "Central number registry",
    text = "Business numbers MUST be drawn from one central, gapless registry.",
    rationale = "A number drawn twice cannot be repaired after the fact.",
    sources = @Source(extern = "GoBD Rz. 36"))
public final class RDecNumbering {
    private RDecNumbering() {}
}
