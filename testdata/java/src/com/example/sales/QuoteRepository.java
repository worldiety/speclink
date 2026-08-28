package com.example.sales;

import org.springframework.data.repository.CrudRepository;

/** Stores quotes as current state. */
public interface QuoteRepository extends CrudRepository<Quote, String> {}
