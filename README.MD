# Rule Engine with Caching

## Overview
This project is a high-performance Rule Engine server built in Go, designed to respond to HTTP requests based on predefined rules. The engine uses caching to store JSON response files and request headers, enabling a response time of under 15ms. Rules are matched based on specific headers in incoming requests, with caching layers ensuring high efficiency and quick access to frequently used data.

## Features
- **Header-Based Rule Matching**: Matches incoming request headers against a list of predefined rules, returning JSON responses based on matching conditions.
- **Response Caching**: Caches JSON responses from disk, reducing I/O operations and enhancing response time.
- **Header Caching**: Stores incoming request headers to speed up rule lookups and avoid redundant processing.
- **Efficient Performance**: Optimized to deliver responses in under 15ms by leveraging Go's concurrency and caching mechanisms.
- **Docker-Ready**: Easily deployable in containerized environments.

## Getting Started

### Prerequisites
- Docker

### Installation

1. **Clone the repository**:
    ```bash
    git clone https://github.com/arankeparth/rule-engine.git
    cd rule-engine
    ```

2. **Set up the configuration**:
   - Define your rules in a JSON file (e.g., `rules.json`) within the `responses` directory. Rules should match the format in the sample file.
   - Ensure each response file is stored in the `responses` folder.

### Run the Server
To start the server, use:
```bash
chmod +x deploy.sh
./deploy.sh
```
### Usage
The server will be availalble at localhost:5000.
##### sample request
```bash
curl -X GET http://localhost:5000 \
  -H "Car: tata" \
  -H "Vegetable: tomato" \
  -H "Fruit: mango" \
  -H "Metal: copper" \
  -H "Planet: mercury" \
  -H "Animal: lion" \
  -H "Bird: owl"
```
