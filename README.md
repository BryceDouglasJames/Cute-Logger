<div align="center">
<h1>Cute Logger</h1>

![Build Status](https://github.com/BryceDouglasJames/Cute-Logger/actions/workflows/go-test.yaml/badge.svg) [![Go Report Card](https://goreportcard.com/badge/github.com/BryceDouglasJames/Cute-Logger)](https://goreportcard.com/report/github.com/BryceDouglasJames/Cute-Logger)
</div>

## Overview

This project develops a comprehensive log system focusing on scalability, efficiency, and reliability. It's designed to manage sequential data records, suitable for applications ranging from logging to complex distributed event sourcing systems.

## Table of Contents

- [Overview](#overview)
- [Project Structure](#project-structure)
  - [Record](#record)
  - [Store](#store)
  - [Index](#index)
  - [Segment](#segment)
  - [Log](#log)
- [Key Features](#key-features)
- [Getting Started](#getting-started)
  - [Prerequisites](#prerequisites)
  - [Installation](#installation)
  - [Usage](#usage)
- [Contributing](#contributing)
- [License](#license)
- [Acknowledgments](#acknowledgments)

## Project Structure

### [Record](#)
The Record type represents the fundamental data unit within the system. Defined using protocol buffers (protobuf), it ensures a standardized and efficient format for serializing and deserializing data. This structure is crucial for data interchange between components, such as between producers and consumers in our messaging system, or for persisting and retrieving data in the adapted storage system.

### [Store](#)
The Store component is responsible for the durable storage of Record instances on disk. It manages the append-only log structure, allowing for efficient writes by sequentially adding records to the end of the log. The Store also supports reading records from any position within the log, enabling efficient data retrieval based on offsets. This component plays a critical role in ensuring data persistence and recoverability.

### [Index](#)
Indexes provide a way to quickly locate data within segments. An index entry will point to the location of data within a segment, significantly speeding up the retrieval of data without scanning entire segments.

Feature Considerations/Plans:

❌ Index Types: Support different index types (e.g., B-Tree Indexing, Bloom Filters) to optimize for various query patterns and performance needs.

✅ Memory Mapping: Utilize memory-mapped files for indexes to improve read performance, especially for large indexes. Currently being achieved by [this go module](https://github.com/tysonmote/gommap/tree/master)

❌ Dynamic Indexing: Allow dynamic index creation and modification to support evolving data structures and query requirements without significant downtime.

### [Segment](#)
These are typically larger blocks of data or files that store the actual data entries (e.g., log records). Segments help in organizing data in a way that is manageable and scalable, allowing systems to append new data efficiently and, when necessary, perform compaction or deletion on older segments.

Feature Considerations/Plans:

❌ Segment Compaction: Implement segment compaction to merge segments and remove redundant or obsolete data, optimizing storage usage and improving query performance.

❌ Segment Caching: Explore caching frequently accessed segments in memory to speed up data retrieval.

❌ Write-Ahead Logging: Consider integrating a WAL mechanism for segments to ensure data integrity and allow for recovery in case of unexpected failures.

### [Log](#)
The Log component encapsulates the entire logging system, tying together Record, Store, Index, and Segment components to provide a unified and efficient mechanism for data storage, retrieval, and management. It serves as the central interface for writing to and reading from the log, offering a comprehensive solution for handling log data at scale.

Feature Considerations/Plans:

❌  Data Retention Policies: Implement customizable data retention policies allowing for automatic data aging and removal, balancing storage utilization with data availability needs.

❌  Query Support: Extend the log with capabilities to perform more complex queries on the data, facilitating the extraction of meaningful insights directly from the log.

😳 Replication, Consensus and Fault Tolerance: Incorporating the Raft consensus algorithm enhances the logging system's fault tolerance and data consistency. Raft ensures that all changes to the log are replicated across cluster nodes in a consistent manner, even in the face of failures.

✅ Strongly Typed Interfaces: gRPC provides strongly typed interfaces for communication, ensuring that interactions between distributed components are well-defined and less prone to errors. This is crucial for maintaining the integrity of log operations across different parts of a distributed system.

❌ Security measures: Incorporate robust security measures, including access controls, encryption, and audit trails, to safeguard sensitive data and comply with regulatory requirements.


### [Key Features](#)
- **Scalable Architecture**: Designed to efficiently handle growing data volumes, the log system organizes data into segments, allowing for scalable storage solutions that can grow with the system's needs.

- **Efficient Data Access**: By using indexing strategies, the log system optimizes read and write operations, enabling quick data access and high-throughput performance.

- **Reliability and Durability**: The system ensures data integrity and availability through durable storage mechanisms, recovery procedures, and potential replication features.

- **Flexible Data Management**: Supports various data retention policies and compaction strategies to manage storage space efficiently, removing obsolete data while preserving the log's integrity.


### [Getting Started](#)

### [Prerequisites](#)

### [Installation](#)

### [Contributing](#)
Contributions are welcome :) Submit bug reports, feature requests, and pull requests through GitHub Issues and PRs.

### [License](#)
This project is licensed under the Apache 2.0 License.

### [Acknowledgments](#)
Inspired by the desire for having more hands on towards building robust, scalable applications in a distributed enviornment.

Big thank you to MIT 6.824 and [this book](https://pragprog.com/titles/tjgo/distributed-services-with-go/) for the guard-rails, ideas and some boilerplate for this project.

Thanks to the Go community for support and resources :)
