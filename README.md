# photostock API

**photostock API** is a backend service designed to power a subscription-based platform for digital content licensing. Whether you're serving images, videos, or creative assets, this API provides secure authentication, flexible subscription plans, asset management, and licensing features for creators and customers alike.

---

## 📦 Version

- **API Version**: `v1.0.0`

---

## 🚀 Features

### 🔐 Authentication & Authorization
- JWT-based user authentication
- Role-based access control (`admin`, `subscriber`)

### 📁 Digital Asset Management
- Upload, organize, and manage content (images, videos, etc.)
- Watermark support and format handling

### 💳 Subscriptions & Licensing
- Subscription plans (monthly, yearly, pay-per-download)
- Usage tracking and download limits
- Auto-generated license certificates with download logs

### 🔎 Content Discovery
- Full-text search with keyword tagging
- Filters by category, format, resolution, orientation
- Public asset preview endpoint

### 📊 Contributor & Admin Tools
- Contributor dashboard API: earnings, uploads, performance
- Admin dashboard API: user moderation, asset review, analytics

### 📈 Analytics & Reporting
- Download counts, revenue reports, top contributors
- Daily/Monthly usage summaries
- Event logging for audit trails

### 🛠️ Developer & Deployment Ready
- RESTful API design
- Swagger/OpenAPI documentation (coming soon)
- Dockerized for container-based deployments

---

## 🧱 Tech Stack

- **Go (1.22+)** – High-performance, scalable backend
- **Gin** – Fast and flexible HTTP router
- **PostgreSQL** – Relational database for structured data
- **JWT** – Token-based authentication

---

## ⚙️ Getting Started

### Prerequisites

- Go 1.22+
- PostgreSQL 16+
- Docker (optional but recommended for local dev)

### Setup

```bash
# Clone the repository
git clone https://github.com/yourusername/photostock-api.git
cd photostock-api

# Set environment variables (or copy .env.example to .env)
cp .env.example .env

# Run the application
go run cmd/main.go
