CREATE INDEX articles_search_pgroonga_idx ON articles
    USING pgroonga ((title || ' ' || summary || ' ' || content));