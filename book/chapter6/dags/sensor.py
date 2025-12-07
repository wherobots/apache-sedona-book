from datetime import datetime, timedelta

import botocore
from airflow.sensors.base import BaseSensorOperator
from airflow.utils.decorators import apply_defaults
from airflow.exceptions import AirflowException, AirflowSkipException
import boto3

bucket_name = "apache-sedona-book-local"
prefix = "transportation/releases/"


class GeoParquetDataReleaseSensor(BaseSensorOperator):
    @apply_defaults
    def __init__(self, *args, **kwargs):
        super().__init__(*args, **kwargs)

    def _get_possible_releases(self):
        s3 = boto3.client(
            's3',
            endpoint_url='http://minio:9000',
            aws_access_key_id='sedona',
            aws_secret_access_key='sedona_password',
            region_name='us-east-1'
        )

        response = s3.list_objects_v2(Bucket=bucket_name, Prefix=prefix, Delimiter="/")

        if 'CommonPrefixes' not in response:
            return []

        result = []
        for obj in response['CommonPrefixes']:

            result.append(
                obj['Prefix'].replace(prefix, "")[:10]
            )

        return result

    def poke(self, context):
        releases = self._get_possible_releases()
        dt = context['execution_date']
        date_formatted = dt.strftime('%Y-%m-%d')
        self.log.info(f"Checking for release {date_formatted}")
        self.log.info(releases)

        if not releases:
            self.log.info("No releases found.")
            raise AirflowException("No releases found.")

        if date_formatted in releases:
            self.log.info(f"Found release {date_formatted}")
            return True

        current_date = datetime.now()

        if releases[-1] > date_formatted:
            self.log.info(f"Latest release is {releases[-1]}.")
            raise AirflowSkipException("Skipping this task no release with that date")

        if (current_date - timedelta(days=1)).strftime('%Y-%m-%d') > date_formatted:
            self.log.info(f"Last release is in the past")
            raise AirflowSkipException("Last release is in the past")

        return False

