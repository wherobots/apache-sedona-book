import sys

import sedona.spark as s
import pyspark.sql.functions as f
import os


bucket_name = os.environ.get("SEDONA_SOURCE_BUCKET", "apache-sedona-book")
CATALOG_NAME = "sedona_catalog"

(processing_date, input_database, output_database) = sys.argv[1:4]

config = s.SedonaContext.builder() \
    .getOrCreate()

sedona = s.SedonaContext.create(config)
sedona.conf.set("spark.sql.sources.partitionOverwriteMode", "dynamic")

nl_shape = sedona.table(f"{CATALOG_NAME}.{input_database}.nl_shape") \
    .selectExpr("ST_GeomFROMEWKB(geometry) AS geometry")

sedona \
    .read.format("geoparquet") \
    .load(f"s3a://{bucket_name}/source_data/lakehouse/segments/") \
    .alias("t") \
    .join(nl_shape.alias("n"), f.expr("ST_Intersects(n.geometry, t.geometry)")) \
    .select("t.*") \
    .where("class = 'motorway'") \
    .selectExpr("id", "geometry", "speed_limits.max_speed.value AS max_speed") \
    .createOrReplaceTempView("nl_transportation")

sedona.sql(
    f"""
    INSERT OVERWRITE {CATALOG_NAME}.{output_database}.nl_transportation
    SELECT ID, ST_AsEWKB(GEOMETRY) AS GEOMETRY, MAX_SPEED, CAST('{processing_date}' AS DATE)
    FROM nl_transportation
    """
)
