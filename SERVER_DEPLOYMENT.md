# Server Deployment Guide

## 🚀 Quick Server Setup

### Step 1: Clone the Repository
```bash
git clone <your-infrastructure-repo-url>
cd asset-tagging-infrastructure
```

### Step 2: Install Dependencies
```bash
# Install Terraform
curl -fsSL https://apt.releases.hashicorp.com/gpg | sudo apt-key add -
sudo apt-add-repository "deb [arch=amd64] https://apt.releases.hashicorp.com $(lsb_release -cs) main"
sudo apt update && sudo apt install terraform

# Install AWS CLI
curl "https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip" -o "awscliv2.zip"
unzip awscliv2.zip
sudo ./aws/install
rm -rf aws awscliv2.zip

# Install Docker (if needed)
sudo apt update
sudo apt install docker.io docker-compose
sudo usermod -aG docker $USER
```

### Step 3: Configure AWS
```bash
aws configure
# Enter your AWS Access Key ID
# Enter your AWS Secret Access Key
# Enter your default region (e.g., us-east-1)
# Enter your output format (json)
```

### Step 4: Create SSH Key Pair
```bash
aws ec2 create-key-pair --key-name asset-tagging-key --query 'KeyMaterial' --output text > ~/.ssh/asset-tagging-key.pem
chmod 400 ~/.ssh/asset-tagging-key.pem
```

### Step 5: Configure Deployment
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

### Step 6: Deploy Infrastructure
```bash
# Quick deployment
make quickstart
make deploy

# Or manual deployment
cd terraform
terraform init
terraform plan
terraform apply
```

### Step 7: Deploy Application
```bash
# Build and push your Go application
cd ../backend
docker build -t asset-tagging-backend:latest .

# The application will be automatically deployed via user data script
# Wait 5-10 minutes for the EC2 instance to fully initialize
```

### Step 8: Verify Deployment
```bash
# Check application health
make health

# Get access information
make monitoring
make ssh
```

## 🔧 Management Commands

```bash
# Check status
make status

# View logs
make dev-logs

# Health check
make health

# Access monitoring
make monitoring

# SSH access
make ssh

# Destroy infrastructure (if needed)
make destroy
```

## 📊 Access URLs

After deployment, you'll have access to:
- **Application**: `http://<load-balancer-dns>`
- **Prometheus**: `http://<server-ip>:9090`
- **Grafana**: `http://<server-ip>:3000` (admin/admin123)

## 🔒 Security Notes

- SSH key is required for server access
- Database is in private subnet
- All traffic goes through load balancer
- SSL/TLS encryption enabled
- Automated backups configured

## 🚨 Troubleshooting

### Common Issues:
1. **Terraform errors**: Check AWS credentials and permissions
2. **Application not responding**: Wait 5-10 minutes for initialization
3. **Database connection issues**: Check security groups
4. **SSL certificate issues**: Verify domain configuration

### Logs:
```bash
# Application logs
ssh -i ~/.ssh/asset-tagging-key.pem ubuntu@<server-ip>
sudo journalctl -u asset-tagging.service -f

# Docker logs
docker logs asset-tagging-app
```

## 💰 Cost Estimation

Estimated monthly costs:
- EC2 t3.medium: ~$30/month
- RDS db.t3.micro: ~$15/month
- ALB: ~$20/month
- Data Transfer: ~$5-10/month
- **Total: ~$70-80/month**

---

**Your infrastructure is now ready! 🎉** 