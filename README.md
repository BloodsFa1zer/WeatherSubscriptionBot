# WeatherSubscriptionBot

## Description

WeatherSubscriptionBot is a Go (Golang) application designed to manage weather subscriptions for users. It allows users to subscribe to weather updates for specific locations and receive notifications based on their chosen frequency. The application leverages a PostgreSQL database for storing subscription data and provides both gRPC and RESTful endpoints for interaction.

## Features

- Weather subscription management
- CRUD operations for weather subscriptions
- Integration with a PostgreSQL database
- Docker support for containerization
- Environment variable configuration for flexible deployment

## Installation

### Prerequisites

- Go 
- MongoDB
- Docker
- Telegram-API

### Steps

1. Clone the repository:

    ```sh
    git clone https://github.com/BloodsFa1zer/WeatherSubscriptionBot.git
    cd WeatherSubscriptionBot
    ```

2. Install dependencies:

    ```sh
    go mod tidy
    ```

3. Set up the PostgreSQL database:
    - Create a database named `weather_subscription_bot`.
    - Configure your environment variables as specified in the `env.dist` file.

4. Run the application:

    ```sh
    go run main.go
    ```

### Docker Setup

1. Build the Docker image:

    ```sh
    docker build -t weather-subscription-bot .
    ```

2. Run the Docker container:

    ```sh
    docker run -p 50051:50051 weather-subscription-bot
    ```

## Usage

### gRPC Endpoints

- **CreateSubscription** - Create a new subscription
- **GetSubscription** - Get a subscription by ID
- **UpdateSubscription** - Update a subscription by ID
- **DeleteSubscription** - Delete a subscription by ID
- **ListSubscriptions** - List all subscriptions

### Example gRPC Client

You can use a gRPC client to interact with the service. Here is an example using the `grpcurl` command-line tool:

- List all subscriptions:

    ```sh
    grpcurl -plaintext localhost:50051 list
    ```

- Create a new subscription:

    ```sh
    grpcurl -plaintext -d '{"location": "New York", "frequency": "daily"}' localhost:50051 SubscriptionService/CreateSubscription
    ```

- Get a subscription by ID:

    ```sh
    grpcurl -plaintext -d '{"id": "1"}' localhost:50051 SubscriptionService/GetSubscription
    ```

## Configuration

The application configuration is managed through environment variables. Ensure that you set the appropriate values for your environment. You can refer to the `env.dist` file for the required variables:

```text
URL_WEATHER=put_your_weather_url_here
API_KEY_WEATHER=put_your_api_for_weather_here
BOT_API_KEY=put_bot_key_here
URI_MongoDB=put_url_to_database_connection
