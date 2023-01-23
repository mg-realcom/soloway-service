CREATE TABLE IF NOT EXISTS placements.soloway
(
    clicks INT64,
    cost INT64,
    placement_id STRING,
    placement_name STRING,
    exposures INT64,
    date DATE,
    date_update TIMESTAMP

) PARTITION BY date;