# Chapter 4. Points, Lines, and Polygons: Vector Data Analysis with Spatial SQL

In this chapter we will explore how to perform vector data analysis using Apache Sedona
and Spatial SQL. We will cover the following topics:
- Vector Data Model and Spatial Relationship, we cover the basics of vector data model like
  - point, lines, polygons
  - spatial relationships like intersects, contains, within
  - DE-9IM model
  
  [SpatialRelationships.ipynb](SpatialRelationships.ipynb)

- Spatial Reference System and the Geography Model, you will learn about the spatial model,
  map projections, map transformations
- Spatial SQL and Vector Data Manipulation, we cover the spatial SQL function which helps you
  to manipulate vector data, you will learn about function like:
  - ST_GeomFromText
  - ST_GeometryType
  - ST_SRID
  - ST_Transform
  - ST_DistanceSphere
  
  and many more.

  [SpatialSQLAndVectorManipulation.ipynb](SpatialSQLAndVectorManipulation.ipynb)

- Spatial Queries, this is important part of the chapter where we discuss, join types, distributed joins
  and then we seamlessly move to spatial joins, distributed spatial joins, spatial partitioning, spatial indexing
  and distributed KNN join.

  [SpatialQueries.ipynb](SpatialQueries.ipynb)

- Hands-On Use Case: Real Estate Analysis in the last part we will use the knowledge gained in this chapter
  to analyze real estate data.

  [RealEstateAnalysis.ipynb](RealEstateAnalysis.ipynb)