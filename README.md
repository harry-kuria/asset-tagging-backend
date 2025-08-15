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

## 📊 Project Status

### ✅ **Completed Tasks:**
- 🚀 **Backend Migration**: Complete Node.js to Go migration
- 🏗️ **Infrastructure as Code**: Full AWS infrastructure setup
- 🔧 **Terraform Configuration**: All issues resolved and validated
- 📚 **Documentation**: Comprehensive setup guides and troubleshooting
- 🐳 **Docker Support**: Containerized development and deployment
- 🔒 **Security**: Enterprise-grade security configuration
- 📊 **Monitoring**: Prometheus + Grafana monitoring stack

### 🎯 **Current State:**
- **Backend**: Production-ready Go application
- **Infrastructure**: Validated Terraform configuration
- **Documentation**: Complete setup and deployment guides
- **Repositories**: Both backend and infrastructure pushed to GitHub

### 🚀 **Ready for:**
- Local development setup
- Production deployment on AWS
- Infrastructure scaling and management
- Monitoring and maintenance

## ⚡ Quick Start

### Prerequisites
- Go 1.18+ | MySQL 5.7+ | Git

### 1. Install Dependencies
```bash
# Install Go
sudo apt update && sudo apt install golang-go

# Install Terraform
wget https://releases.hashicorp.com/terraform/1.5.7/terraform_1.5.7_linux_amd64.zip
unzip terraform_1.5.7_linux_amd64.zip && sudo mv terraform /usr/local/bin/

# Install AWS CLI
sudo apt install awscli -y
```

### 2. Setup Backend
```bash
cd backend
go mod tidy
cp .env.example .env  # Edit with your database credentials
go run .
```

### 3. Deploy to Cloud (Optional)
```bash
git clone https://github.com/harry-kuria/moowi-IAC.git
cd moowi-IAC
aws configure  # Set up AWS credentials
make quickstart
make deploy
```

**🎯 Your application will be running at `http://localhost:5000`**

## Prerequisites

- Go 1.18 or higher
- MySQL 5.7 or higher
- Git

## 🚀 Complete Setup Guide

### Local Development Setup

#### 1. Install Go
```bash
# Ubuntu/Debian
sudo apt update
sudo apt install golang-go

# Or download from https://golang.org/dl/
```

#### 2. Install Terraform (for infrastructure deployment)
```bash
# Download and install Terraform
wget https://releases.hashicorp.com/terraform/1.5.7/terraform_1.5.7_linux_amd64.zip
unzip terraform_1.5.7_linux_amd64.zip
sudo mv terraform /usr/local/bin/
rm terraform_1.5.7_linux_amd64.zip

# Verify installation
terraform --version
```

#### 3. Install AWS CLI
```bash
# Ubuntu/Debian
sudo apt install awscli -y

# Or install AWS CLI v2
curl "https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip" -o "awscliv2.zip"
unzip awscliv2.zip
sudo ./aws/install
rm -rf aws awscliv2.zip
```

#### 4. Install Docker (optional, for containerized development)
```bash
sudo apt update
sudo apt install docker.io docker-compose
sudo usermod -aG docker $USER
# Log out and back in for group changes to take effect
```

### Backend Application Setup

#### 1. Navigate to the backend directory
```bash
cd backend
```

#### 2. Install Go dependencies
```bash
go mod tidy
```

#### 3. Set up environment variables
Create a `.env` file in the backend directory:
```env
HOST=localhost:3306
USERNAME=your_mysql_username
PASSWORD=your_mysql_password
DB=asset_management
JWT_SECRET=your_jwt_secret_key
PORT=5000
```

#### 4. Set up the database
```bash
# Create MySQL database
mysql -u root -p
CREATE DATABASE asset_management;
USE asset_management;
SOURCE asset_management.sql;
EXIT;
```

#### 5. Run the application locally
```bash
go run .
```

### 🏗️ Infrastructure Setup (Production Deployment)

#### 1. Clone the Infrastructure Repository
```bash
git clone https://github.com/harry-kuria/moowi-IAC.git
cd moowi-IAC
```

#### 2. Configure AWS Credentials
```bash
aws configure
# Enter your AWS Access Key ID
# Enter your AWS Secret Access Key
# Enter your default region (e.g., us-east-1)
# Enter your output format (json)
```

#### 3. Create SSH Key Pair
```bash
aws ec2 create-key-pair --key-name asset-tagging-key --query 'KeyMaterial' --output text > ~/.ssh/asset-tagging-key.pem
chmod 400 ~/.ssh/asset-tagging-key.pem
```

#### 4. Configure Deployment Variables
```bash
# Copy example configuration
cp terraform/terraform.tfvars.example terraform/terraform.tfvars

# Edit configuration
nano terraform/terraform.tfvars
```

**Important variables to set:**
```hcl
ssh_key_name = "asset-tagging-key"
db_password = "your-secure-password"
domain_name = "your-domain.com"  # Optional
```

#### 5. Deploy Infrastructure
```bash
# Quick deployment
make quickstart
make deploy

# Or manual deployment
cd terraform
terraform init
terraform validate  # Verify configuration is valid
terraform plan      # Review what will be created
terraform apply     # Deploy infrastructure
```

**✅ All Terraform issues have been resolved:**
- Missing files created (`user_data.sh`, `terraform.tfvars.example`)
- Variable interpolation fixed
- Configuration validates successfully
- Ready for production deployment

#### 6. Deploy Application
```bash
# Build and push your Go application
cd ../backend
docker build -t asset-tagging-backend:latest .

# The application will be automatically deployed via user data script
# Wait 5-10 minutes for the EC2 instance to fully initialize
```

#### 7. Verify Deployment
```bash
# Check application health
make health

# Get access information
make monitoring
make ssh
```

### 🐳 Docker Development Setup

#### 1. Start Development Environment
```bash
cd backend
docker-compose up -d
```

#### 2. Access Services
- **Application**: http://localhost:5000
- **MySQL**: localhost:3306
- **Nginx**: http://localhost:80

#### 3. View Logs
```bash
docker-compose logs -f
```

#### 4. Stop Development Environment
```bash
docker-compose down
```

### 🔧 Management Commands

```bash
# Infrastructure management
make init      # Initialize Terraform
make plan      # Plan infrastructure changes
make apply     # Apply infrastructure changes
make destroy   # Destroy infrastructure
make status    # Check infrastructure status

# Development environment
make dev-up    # Start development environment
make dev-down  # Stop development environment
make dev-logs  # View development logs

# Health checks
make health    # Check application health
make monitoring # Access monitoring dashboards
make ssh       # SSH access information
```

### 📊 Access URLs

After deployment, you'll have access to:
- **Application**: `http://<load-balancer-dns>`
- **Prometheus**: `http://<server-ip>:9090`
- **Grafana**: `http://<server-ip>:3000` (admin/admin123)

### 💰 Cost Estimation

Estimated monthly costs:
- **EC2 t3.medium**: ~$30/month
- **RDS db.t3.micro**: ~$15/month
- **ALB**: ~$20/month
- **Data Transfer**: ~$5-10/month
- **Total**: ~$70-80/month

### 🔒 Security Notes

- SSH key is required for server access
- Database is in private subnet
- All traffic goes through load balancer
- SSL/TLS encryption enabled
- Automated backups configured

### 🚨 Troubleshooting

#### Common Issues:

1. **Terraform errors**: Check AWS credentials and permissions
2. **Application not responding**: Wait 5-10 minutes for initialization
3. **Database connection issues**: Check security groups
4. **SSL certificate issues**: Verify domain configuration

#### Recent Fixes Applied:

✅ **Missing user_data.sh**: Fixed - Comprehensive startup script created  
✅ **Missing terraform.tfvars.example**: Fixed - Configuration template added  
✅ **DATE variable interpolation**: Fixed - Escaped variables in template  
✅ **Output type consistency**: Fixed - Monitoring URLs return consistent types  
✅ **Terraform validation**: Fixed - Configuration now validates successfully  

#### Terraform-Specific Issues:

**If you encounter "Invalid function argument" errors:**
```bash
# Ensure all required files exist
ls -la terraform/
# Should show: main.tf, variables.tf, outputs.tf, user_data.sh, terraform.tfvars.example

# Validate configuration
terraform validate

# If validation fails, check for:
# - Missing variables in terraform.tfvars
# - Incorrect variable types
# - Missing required files
```

**If you get "templatefile" errors:**
- Ensure `user_data.sh` exists in the terraform directory
- Check that all variables referenced in the script are provided in the templatefile call
- Verify that shell variables are properly escaped (use `$$` instead of `$`)

#### Logs:
```bash
# Application logs
ssh -i ~/.ssh/asset-tagging-key.pem ubuntu@<server-ip>
sudo journalctl -u asset-tagging.service -f

# Docker logs
docker logs asset-tagging-app

# Terraform logs
terraform plan -detailed-exitcode
terraform apply -auto-approve
```

## 🏗️ Infrastructure Repository

This backend is designed to work with the **Moowi Infrastructure as Code** repository:

### Repository: [moowi-IAC](https://github.com/harry-kuria/moowi-IAC.git)

**Features:**
- 🚀 **Complete AWS Infrastructure**: VPC, EC2, RDS, ALB, CloudWatch
- 🔒 **Enterprise Security**: Private subnets, security groups, SSL/TLS
- 📊 **Full Monitoring**: Prometheus + Grafana dashboards
- 🐳 **Docker Ready**: Containerized deployment
- ⚡ **One-Command Deployment**: `make deploy`
- 💰 **Cost Optimized**: ~$70-80/month

**Quick Infrastructure Setup:**
```bash
git clone https://github.com/harry-kuria/moowi-IAC.git
cd moowi-IAC
make quickstart
make deploy
```

**Repository Structure:**
```
moowi-IAC/
├── terraform/          # AWS infrastructure as code
├── docker/            # Local development setup
├── deployment/        # Automated deployment scripts
├── monitoring/        # Prometheus & Grafana configs
├── Makefile          # Easy management commands
└── README.md         # Complete documentation
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

This project is licensed under the ISC License. # CI Test
