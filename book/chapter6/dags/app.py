import sys

import sedona.spark as s
import pyspark.sql.functions as f

args = sys.argv[1:]
processing_date = args[0]

config = (
    s.SedonaContext.builder() \
    .appName("Sedona") \
    .config("spark.shuffle.partitions", "8") \
    .getOrCreate()
)

sedona = s.SedonaContext.create(config)

transportation = sedona.read \
    .format("geoparquet") \
    .load(f"s3a://apache-sedona-book-local/transportation/releases/{processing_date}")

transportation\
    .withColumn("geohash", f.expr("ST_GeoHash(geometry, 6)")) \
    .sort("geohash") \
    .write \
    .format("geoparquet") \
    .mode("overwrite") \
    .save(f"s3a://apache-sedona-book-local/transportation_processed/{processing_date}")
