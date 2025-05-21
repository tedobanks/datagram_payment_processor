
# Datagram Payment Processor

This Go backend server powers the financial and currency operations for the Datagram application. Its core functionalities include:

#### Payment Processing

Integrates with Paystack to allow users to purchase datacredit (an in-app currency pegged to NGN kobo).

#### In-App Currency Management

Manages user balances for two types of in-app currencies: datacredit and databyte.
Facilitates the conversion of datacredit into databyte based on a defined exchange rate.

#### Withdrawal System

Enables users to withdraw their datacredit balance, initiating payouts via Paystack.

#### Webhook Handling

Securely processes incoming webhooks from Paystack to update transaction statuses and user balances in real-time (e.g., crediting datacredit upon successful payment).

#### Database Interaction

Connects to a Supabase (PostgreSQL) database to store and manage user profiles, wallet balances, and detailed transaction histories.

#### API Endpoints

Exposes a RESTful API built with the Gin framework for all the above operations.

#### Authentication

Secures protected API routes by validating JWTs issued by Supabase Auth, ensuring only authenticated users can perform sensitive actions.

#### API Documentation

Provides Swagger (OpenAPI) documentation for clear and interactive API exploration.

## Authors

- [@tedobanks](https://www.github.com/tedobanks)

## Feedback

If you have any feedback, please reach out to us at <bankong.ted@gmail.com>
