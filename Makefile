# Asset Tagging Infrastructure Makefile
# Provides easy commands for managing the infrastructure

.PHONY: help init plan apply destroy validate clean dev-up dev-down logs status backup restore

# Default target
help:
	@echo "🚀 Asset Tagging Infrastructure Management"
	@echo "=========================================="
	@echo ""
	@echo "Available commands:"
	@echo "  init      - Initialize Terraform"
	@echo "  plan      - Plan infrastructure changes"
	@echo "  apply     - Apply infrastructure changes"
	@echo "  destroy   - Destroy infrastructure"
	@echo "  validate  - Validate Terraform configuration"
	@echo "  clean     - Clean up temporary files"
	@echo "  dev-up    - Start development environment"
	@echo "  dev-down  - Stop development environment"
	@echo "  logs      - View application logs"
	@echo "  status    - Check infrastructure status"
	@echo "  backup    - Create backup"
	@echo "  restore   - Restore from backup"
	@echo "  deploy    - Full deployment (init, plan, apply)"
	@echo "  setup     - Setup server manually"
	@echo ""

# Terraform commands
init:
	@echo "🔧 Initializing Terraform..."
	cd terraform && terraform init

plan:
	@echo "📋 Planning infrastructure changes..."
	cd terraform && terraform plan

apply:
	@echo "🚀 Applying infrastructure changes..."
	cd terraform && terraform apply -auto-approve

destroy:
	@echo "🗑️  Destroying infrastructure..."
	cd terraform && terraform destroy -auto-approve

validate:
	@echo "✅ Validating Terraform configuration..."
	cd terraform && terraform validate

# Development environment
dev-up:
	@echo "🐳 Starting development environment..."
	cd docker && docker-compose up -d

dev-down:
	@echo "🛑 Stopping development environment..."
	cd docker && docker-compose down

dev-logs:
	@echo "📋 Viewing development logs..."
	cd docker && docker-compose logs -f

# Infrastructure status
status:
	@echo "📊 Checking infrastructure status..."
	cd terraform && terraform show

# Cleanup
clean:
	@echo "🧹 Cleaning up temporary files..."
	rm -f terraform/.terraform.lock.hcl
	rm -f terraform/terraform.tfstate.backup
	rm -f deployment_outputs.json
	rm -f terraform/tfplan

# Backup and restore
backup:
	@echo "💾 Creating backup..."
	./deployment/backup.sh

restore:
	@echo "📥 Restoring from backup..."
	@echo "Please specify backup file: make restore BACKUP_FILE=backup_20231201_120000.tar.gz"
	@if [ -z "$(BACKUP_FILE)" ]; then \
		echo "Error: BACKUP_FILE not specified"; \
		exit 1; \
	fi
	./deployment/restore.sh $(BACKUP_FILE)

# Full deployment
deploy:
	@echo "🚀 Starting full deployment..."
	./deployment/deploy.sh

# Server setup
setup:
	@echo "🔧 Setting up server..."
	@echo "This command should be run on the target server as root"
	./deployment/setup-server.sh

# Health checks
health:
	@echo "🏥 Running health checks..."
	@if [ -f "deployment_outputs.json" ]; then \
		APP_URL=$$(jq -r '.application_url.value' deployment_outputs.json); \
		echo "Checking application at $$APP_URL"; \
		curl -f "$$APP_URL/api/health" >/dev/null 2>&1 && echo "✅ Application is healthy" || echo "❌ Application is not responding"; \
	else \
		echo "⚠️  No deployment outputs found. Run 'make deploy' first."; \
	fi

# Monitoring
monitoring:
	@echo "📊 Opening monitoring dashboards..."
	@if [ -f "deployment_outputs.json" ]; then \
		PROMETHEUS_URL=$$(jq -r '.monitoring_urls.prometheus' deployment_outputs.json 2>/dev/null); \
		GRAFANA_URL=$$(jq -r '.monitoring_urls.grafana' deployment_outputs.json 2>/dev/null); \
		if [ "$$PROMETHEUS_URL" != "null" ]; then \
			echo "🌐 Prometheus: $$PROMETHEUS_URL"; \
		fi; \
		if [ "$$GRAFANA_URL" != "null" ]; then \
			echo "📈 Grafana: $$GRAFANA_URL"; \
		fi; \
	else \
		echo "⚠️  No deployment outputs found. Run 'make deploy' first."; \
	fi

# SSH access
ssh:
	@echo "🔑 SSH access information..."
	@if [ -f "deployment_outputs.json" ]; then \
		SSH_CMD=$$(jq -r '.ssh_command.value' deployment_outputs.json); \
		echo "SSH Command: $$SSH_CMD"; \
	else \
		echo "⚠️  No deployment outputs found. Run 'make deploy' first."; \
	fi

# Quick start
quickstart:
	@echo "⚡ Quick start setup..."
	@echo "1. Copying example configuration..."
	cp terraform/terraform.tfvars.example terraform/terraform.tfvars
	@echo "2. Please edit terraform/terraform.tfvars with your configuration"
	@echo "3. Run 'make deploy' to deploy the infrastructure"
	@echo "✅ Quick start setup completed" 