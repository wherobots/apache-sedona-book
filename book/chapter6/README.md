# Chapter 6. Apache Sedona and the PyData Ecosystem

In this chapter you will learn ow to efficiently use Apache Sedona in your Python application we will also show how 
to integrate Sedona with popular Python geospatial libraries such as GeoPandas, Shapely, Kepler.gl and others. You will
also learn how to use Apache Sedona with schedulers like Airflow and Prefect to build robust geospatial data pipelines.
We will also cover how to integrate your Sedona functionality to the dbt framework to build geospatial data transformations.
Python application might suffer from serialization and deserialization overhead, we will show you how to minimize this overhead
by using Apache Arrow format for data exchange between Sedona and other Python libraries.

Apache Sedona allows you to write your own custom geospatial functions in Python and use them inside your Sedona SQL queries.
We will show you how to do that using User Defined Functions (UDF) and how to optimize their performance using vectorized UDFs.

Each Subchapter contains a notebook with examples:
- Manipulating Geospatial Vector Data
  [ManipulatingGeospatialVectorData.ipynb](ManipulatingGeospatialVectorData.ipynb)

- Raster Data Tools
  [RasterDataTools.ipynb](RasterDataTools.ipynb)

- Scheduling Your Geospatial Code
  [dags](dags)

- Transforming Your Geospatial Data with dbt
  [geospatialdbt](geospatialdbt)

  // load the flat files
  dbt seed

  // run the transformations
  dbt run

  // test the transformations
  dbt test

- Vector Geospatial Visualization
  [VectorGeospatialVisualization.ipynb](VectorGeospatialVisualization.ipynb)

To run Apache Airflow and dbt example please follow the main README instructions in the root of the repository.