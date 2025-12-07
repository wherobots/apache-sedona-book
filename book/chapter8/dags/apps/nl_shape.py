import sys

import sedona.spark as s
from pyspark.sql.session import SparkSession
import os


bucket_name = os.environ.get("SEDONA_SOURCE_BUCKET", "apache-sedona-book")
CATALOG_NAME = "sedona_catalog"

(
    processing_date,
    input_database,
    output_database,
) = sys.argv[1:4]

config = s.SedonaContext.builder() \
    .getOrCreate()

sedona = SparkSession.builder.getOrCreate()

sedona.conf.set("spark.sql.sources.partitionOverwriteMode", "dynamic")

nl_shape = sedona \
    .read.format("geoparquet") \
    .load(f"s3a://{bucket_name}/source_data/lakehouse/division/") \
    .select("id", "geometry") \
    .where("country = 'NL' AND subtype = 'country'") \
    .alias("b") \
    .select("b.id", "b.geometry") \
    .createOrReplaceTempView("nl_shape")

sedona.sql(
    f"""
    INSERT OVERWRITE {CATALOG_NAME}.{output_database}.nl_shape
    SELECT ST_AsEWKB(GEOMETRY) AS GEOMETRY, CAST('{processing_date}' AS DATE)
    FROM nl_shape
    """
)
