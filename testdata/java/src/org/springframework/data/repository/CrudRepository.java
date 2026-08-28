package org.springframework.data.repository;

/** Stands in for Spring Data's interface of the same name. */
public interface CrudRepository<T, ID> {
    T save(T entity);
}
