import sys

import sedona.spark as s
import pyspark.sql.functions as f
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

buildings_categories = ['residential', 'apartments']

nl_shape = sedona.table(f"{CATALOG_NAME}.{input_database}.nl_shape")\
    .selectExpr("ST_GEOMFromEWKB(geometry) AS geometry")

nl_shape.show()

sedona \
    .read.format("geoparquet") \
    .load(f"s3a://{bucket_name}/source_data/lakehouse/buildings/").\
    show()

nl_buildings = sedona \
    .read.format("geoparquet")\
    .load(f"s3a://{bucket_name}/source_data/lakehouse/buildings/") \
    .alias("b") \
    .join(nl_shape.alias("n"), f.expr("ST_Intersects(n.geometry, b.geometry)")) \
    .select("b.*") \
    .where(f.col("class").isin(buildings_categories)) \
    .select("id", "geometry") \
    .alias("b") \
    .select("b.id", "b.geometry")

nl_buildings.show()

nl_buildings.createOrReplaceTempView("nl_buildings")

sedona.sql(
    f"""
    INSERT OVERWRITE {CATALOG_NAME}.{output_database}.nl_buildings
    SELECT ID, ST_AsEWKB(GEOMETRY) AS GEOMETRY, CAST('{processing_date}' AS DATE)
    FROM nl_buildings
    """
)
