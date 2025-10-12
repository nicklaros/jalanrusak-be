# JalanRusak Backend Service

JalanRusak is a backend service that enables citizens to report damaged roads in Indonesia. The service allows users to submit reports with photos, GPS coordinates, and location details, while providing a verification system for tracking repairs.

## Features

- 🚗 Submit damaged road reports with photos and GPS coordinates
- 📍 Location-based organization (subdistrict, district, city, province)
- ✅ Mark reports as repaired with verification workflow
- 👥 User role management (regular users and verificators)
- 🔐 JWT-based authentication and authorization
- 📱 RESTful API for mobile and web client integration

## Technology Stack

- **Language**: Go 1.21+
- **Web Framework**: Gin
- **Architecture**: Hexagonal Architecture (Ports & Adapters)
- **Database**: PostgreSQL
- **Image Storage**: Dynamic (default to Local filesystem)
- **Authentication**: JWT tokens

## Project Structure

```
jalanrusak-be/
- cmd/
  - server/
    - main.go           # Application entry point
- core/
  - services/           # Business use cases implementing `usecases` ports
  - ports/              # Interfaces/ports
    - usecases/         # Interface to define how the outside world drives the application
    - external/         # Interface to define how the core calls external services
  - domain/             # Domain layer
    - entities/         # Domain entities
    - errors/           # Domain errors
- adapters/             # External adapters
  - in/
    - http/
      - handlers/       # HTTP request handlers
      - middleware/     # HTTP middleware
      - routes/         # Route definitions
  - out/
    - repository/
      - postgres/       # PostgreSQL repositories
    - services/         # Third-party service integrations
    - storage/
      - filesystem/     # File storage
- config/               # Configuration
- migrations/           # Database migrations

```