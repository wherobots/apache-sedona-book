import sys

import sedona.spark as s
import os

bucket_name = os.environ.get("SEDONA_SOURCE_BUCKET", "apache-sedona-book")

CATALOG_NAME = "sedona_catalog"

(processing_date, input_database, output_database) = sys.argv[1:4]

config = s.SedonaContext.builder() \
    .getOrCreate()

sedona = s.SedonaContext.create(config)
sedona.conf.set("spark.sql.sources.partitionOverwriteMode", "dynamic")

sedona.sql(
    f"""
       SELECT
            DISTINCT h3
        FROM {CATALOG_NAME}.{input_database}.nl_transportation
        LATERAL VIEW EXPLODE (ST_H3CellIDs(ST_GeomFromEWKB(geometry), 8, true)) AS h3
    """
).createOrReplaceTempView("h3_ids")

sedona.sql(
    """
       SELECT
            h3,
            geometry
        FROM h3_ids
        LATERAL VIEW EXPLODE (ST_H3ToGeom(ARRAY(h3))) AS geometry
    """
).createOrReplaceTempView("hex_bins")


sedona.sql(
    f"""
        SELECT
            t1.h3,
            FIRST(ARRAY_MAX(t2.max_speed)) AS speed_limit,
            FIRST(t1.geometry) AS geometry,
            CAST('{processing_date}' AS DATE) AS processing_date
        FROM hex_bins t1
        LEFT JOIN {CATALOG_NAME}.{input_database}.nl_transportation t2 ON ST_INTERSECTS(t1.geometry, ST_GeomFromEWKB(t2.geometry))
        GROUP BY t1.h3
    """
).createOrReplaceTempView("hex_grids")

sedona.sql(
    f"""
    INSERT OVERWRITE {CATALOG_NAME}.{output_database}.nl_transportation_hex_bins
    SELECT H3, SPEED_LIMIT, ST_AsEWKB(GEOMETRY), PROCESSING_DATE
    FROM hex_grids
    """
)
