# Asset Tagging Backend (Go)

A complete Go backend for the Asset Tagging system, providing RESTful API endpoints for asset management, user authentication, barcode generation, and reporting.

## Features

- **User Authentication**: JWT-based authentication with role-based access control
- **Asset Management**: CRUD operations for assets with comprehensive metadata
- **Barcode Generation**: Generate barcodes for individual assets or bulk operations
- **Reporting**: Generate Excel reports and PDF invoices
- **Trial Management**: Built-in trial period management with license activation
- **File Upload**: Support for asset logo uploads
- **Database**: MySQL database with connection pooling

## Prerequisites

- Go 1.18 or higher
- MySQL 5.7 or higher
- Git

## Installation

1. **Navigate to the backend directory**
   ```bash
   cd backend
   ```

2. **Install dependencies**
   ```bash
   go mod tidy
   ```

3. **Set up environment variables**
   Create a `.env` file in the backend directory:
   ```env
   HOST=localhost:3306
   USERNAME=your_mysql_username
   PASSWORD=your_mysql_password
   DB=asset_management
   JWT_SECRET=your_jwt_secret_key
   PORT=5000
   ```

4. **Set up the database**
   - Create a MySQL database named `asset_management`
   - Import the SQL schema from `asset_management.sql`

5. **Run the application**
   ```bash
   go run .
   ```

## API Endpoints

### Authentication

#### POST /api/login
Login with username and password.
```json
{
  "username": "admin",
  "password": "password123"
}
```

#### POST /api/create-account
Create a new user account.
```json
{
  "username": "newuser",
  "password": "password123"
}
```

### User Management

#### GET /api/users
Get all users (requires authentication).

#### GET /api/users/:id
Get user by ID (requires authentication).

#### POST /api/addUser
Add a new user (requires authentication).
```json
{
  "username": "newuser",
  "password": "password123",
  "roles": ["userManagement", "assetManagement", "encodeAssets"]
}
```

#### PUT /api/users/:id
Update user (requires authentication).

#### DELETE /api/users/:id
Delete user (requires authentication).

### Asset Management

#### GET /api/assets
Get all assets (requires authentication).

#### POST /api/addAsset
Add a new asset (requires authentication).
```json
{
  "assetName": "Laptop Dell XPS",
  "assetType": "Computer",
  "institutionName": "University of Technology",
  "department": "IT Department",
  "functionalArea": "Administration",
  "manufacturer": "Dell",
  "modelNumber": "XPS 13",
  "serialNumber": "DL123456789",
  "location": "Room 101",
  "status": "Active",
  "purchaseDate": "2023-01-15",
  "purchasePrice": 1500.00
}
```

#### PUT /api/assets/:id
Update asset (requires authentication).

#### DELETE /api/assets/:id
Delete asset (requires authentication).

#### GET /api/searchAssets?query=search_term
Search assets (requires authentication).

#### GET /api/getAssetDetails?assetId=123
Get asset details (requires authentication).

#### POST /api/addMultipleAssets/:assetType
Add multiple assets (requires authentication).
```json
{
  "assets": [
    {
      "assetName": "Asset 1",
      "institutionName": "University",
      "department": "IT",
      "location": "Room 101",
      "status": "Active",
      "purchaseDate": "2023-01-15",
      "purchasePrice": 1000.00
    }
  ]
}
```

### Reference Data

#### GET /api/institutions
Get all unique institutions (requires authentication).

#### GET /api/departments
Get all unique departments (requires authentication).

#### GET /api/functionalAreas
Get all unique functional areas (requires authentication).

#### GET /api/categories
Get all asset categories (requires authentication).

#### GET /api/manufacturers
Get all unique manufacturers (requires authentication).

### Barcode Generation

#### POST /api/generateBarcodes
Generate barcodes for specific assets (requires authentication).
```json
{
  "assetIds": [1, 2, 3, 4, 5]
}
```

#### POST /api/generateBarcodesByInstitution
Generate barcodes for all assets in an institution (requires authentication).
```json
{
  "institution": "University of Technology"
}
```

#### POST /api/generateBarcodesByInstitutionAndDepartment
Generate barcodes for assets in a specific institution and department (requires authentication).
```json
{
  "institution": "University of Technology",
  "department": "IT Department"
}
```

### Reports

#### GET /api/generateReport
Generate filtered report (requires authentication).
```
/api/generateReport?assetType=Computer&location=Room%20101&status=Active
```

#### POST /api/generateAssetReport
Generate detailed Excel report (requires authentication).
```json
{
  "assetType": "Computer",
  "location": "Room 101",
  "status": "Active",
  "startDate": "2023-01-01",
  "endDate": "2023-12-31",
  "manufacturer": ["Dell", "HP"],
  "institutionName": "University of Technology",
  "department": "IT Department"
}
```

#### POST /api/generate_invoice
Generate invoice PDF (requires authentication).
```json
{
  "customerName": "John Doe",
  "customerAddress": "123 Main St, City, Country",
  "items": [
    {
      "description": "Laptop Dell XPS",
      "quantity": 2,
      "unitPrice": 1500.00
    }
  ]
}
```

#### GET /api/download?filename=report.xlsx
Download generated files (requires authentication).

### Asset Fetching

#### POST /api/fetchAssetsByInstitution
Fetch assets by institution (requires authentication).
```json
{
  "institution": "University of Technology"
}
```

## Authentication

All protected endpoints require a JWT token in the Authorization header:
```
Authorization: Bearer <your_jwt_token>
```

## Database Schema

The application uses the following main tables:
- `users`: User accounts and authentication
- `user_roles`: User permissions and roles
- `assets`: Asset information and metadata
- `asset_categories`: Asset type categories

## File Structure

```
backend/
├── main.go              # Main application entry point
├── models.go            # Data structures and models
├── auth.go              # Authentication and authorization
├── users.go             # User management handlers
├── assets.go            # Asset management handlers
├── barcodes.go          # Barcode generation handlers
├── reports.go           # Report generation handlers
├── go.mod               # Go module dependencies
├── go.sum               # Dependency checksums
├── README.md            # This file
├── .env                 # Environment variables (create this)
├── asset_management.sql # Database schema
├── assetLogos/          # Directory for uploaded asset logos
├── LicenseKey/          # License key files
└── files/               # Additional files
```

## Building and Deployment

### Build for production
```bash
go build -o asset-tagging-backend .
```

### Run the binary
```bash
./asset-tagging-backend
```

### Docker deployment
```dockerfile
FROM golang:1.18-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o main .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/main .
EXPOSE 5000
CMD ["./main"]
```

## Error Handling

The API returns consistent error responses:
```json
{
  "success": false,
  "error": "Error message"
}
```

## Security Features

- JWT-based authentication
- Password hashing with bcrypt
- CORS configuration
- SQL injection prevention with parameterized queries
- Input validation and sanitization

## Performance Features

- Database connection pooling
- Efficient query optimization
- Static file serving
- Concurrent request handling

## Troubleshooting

### Common Issues

1. **Database connection failed**
   - Check your `.env` file configuration
   - Ensure MySQL is running
   - Verify database credentials

2. **JWT token invalid**
   - Check if the token is expired
   - Verify the JWT_SECRET environment variable
   - Ensure proper Authorization header format

3. **File upload issues**
   - Check if the `assetLogos` directory exists and has write permissions
   - Verify file size limits

### Logs

The application logs important events to stdout. Check the console output for debugging information.

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## License

This project is licensed under the ISC License. 