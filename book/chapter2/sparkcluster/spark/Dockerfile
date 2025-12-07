FROM apache/spark:3.5.6

USER root

RUN pip install apache-sedona
RUN pip install pandas

ENV SEDONA_VERSION=1.8.0
ENV GEOTOOLS_VERSION=33.1
ENV SPARK_VERSION=3.5
ENV SCALA_VERSION=2.12

# org.apache.hadoop:hadoop-client-api:3.3.4,org.apache.hadoop:hadoop-common:3.3.4

RUN wget https://repo1.maven.org/maven2/org/apache/sedona/sedona-spark-shaded-${SPARK_VERSION}_${SCALA_VERSION}/${SEDONA_VERSION}/sedona-spark-shaded-${SPARK_VERSION}_${SCALA_VERSION}-${SEDONA_VERSION}.jar && \
    wget https://repo1.maven.org/maven2/org/datasyslab/geotools-wrapper/${SEDONA_VERSION}-${GEOTOOLS_VERSION}/geotools-wrapper-${SEDONA_VERSION}-${GEOTOOLS_VERSION}.jar

RUN cp sedona-spark-shaded-${SPARK_VERSION}_${SCALA_VERSION}-${SEDONA_VERSION}.jar /opt/spark/jars && \
    cp geotools-wrapper-${SEDONA_VERSION}-${GEOTOOLS_VERSION}.jar /opt/spark/jars

RUN wget https://repo1.maven.org/maven2/org/apache/hadoop/hadoop-aws/3.3.2/hadoop-aws-3.3.2.jar && \
    wget https://repo1.maven.org/maven2/com/amazonaws/aws-java-sdk-bundle/1.12.262/aws-java-sdk-bundle-1.12.262.jar

RUN cp hadoop-aws-3.3.2.jar /opt/spark/jars && \
    cp aws-java-sdk-bundle-1.12.262.jar /opt/spark/jars

USER spark
