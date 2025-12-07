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


flood_layers_dir = f"s3a://{bucket_name}/source_data/lakehouse/flood_map/"

# Create a DataFrame then temporary view containing raster data
flood_layers = sedona.read.format("binaryFile") \
    .load(flood_layers_dir) \
    .withColumn("rast", f.expr("RS_FromGeoTiff(content)")) \
    .drop("content")

nl_shape = sedona.table(f"{CATALOG_NAME}.{input_database}.nl_shape") \
    .selectExpr("ST_GeomFromEWKB(geometry) AS geometry")

# tile the raster into 256x256 images and save as a temporary view
flood_layers \
    .selectExpr("RS_TileExplode(rast, 256, 256) AS (x, y, rast)", "rp AS event") \
    .alias("f") \
    .join(nl_shape.alias("n"), f.expr("RS_INTERSECTS(f.rast, n.geometry)")) \
    .select("f.*") \
    .selectExpr("rast", "event") \
    .where(f.col("event").isin([10, 50, 100])) \
    .createOrReplaceTempView("flood_raster_tiled")

sedona.sql(
    f"""
    INSERT OVERWRITE {CATALOG_NAME}.{output_database}.nl_flood_map
    SELECT EVENT, RS_AsGeoTiff(RAST) AS RAST, CAST('{processing_date}' AS DATE)
    FROM flood_raster_tiled
    """
)
