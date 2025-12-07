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
    INSERT OVERWRITE {CATALOG_NAME}.{output_database}.nl_road_risk_score
    SELECT H3, GEOMETRY, RISK_SCORE, PROCESSING_DATE
    FROM {CATALOG_NAME}.{input_database}.nl_road_risk_score
    """
)
