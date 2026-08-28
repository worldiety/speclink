package com.example.sales;

import jakarta.persistence.Entity;

/** A quote, as it is stored. */
@Entity
public class Quote {

    private String id;
    private String number;
    private String status;

    public String getId() {
        return id;
    }
}
