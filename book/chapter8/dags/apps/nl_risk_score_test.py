import sys

import sedona.spark as s
import os

bucket_name = os.environ.get("SEDONA_SOURCE_BUCKET", "apache-sedona-book")

CATALOG_NAME = "sedona_catalog"

(processing_date, input_database, output_database) = sys.argv[1:4]

config = s.SedonaContext.builder() \
    .getOrCreate()

sedona = s.SedonaContext.create(config)

risk_score = sedona.table(f"{CATALOG_NAME}.{input_database}.nl_road_risk_score").\
    selectExpr("ST_GeomFromEWKB(geometry) AS geometry", "RISK_SCORE").\
    filter("RISK_SCORE <> 0")

risk_score_count = risk_score.count()

assert risk_score_count > 3800, "Quality check failed: Risk score count is greater than 3800"
assert risk_score_count < 3900, "Quality check failed: Risk score count is less than 3900"
assert risk_score.selectExpr("ST_Area(geometry) AS area")\
            .filter("area <= 0")\
            .count() == 0, "Quality check failed: Invalid area"

assert risk_score.filter("RISK_SCORE > 1.0").count() == 0, "Quality check failed: Invalid risk score"
