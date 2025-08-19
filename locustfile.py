import json
import random
import time
from locust import HttpUser, task, between
from datetime import datetime, timedelta

class AssetTaggingUser(HttpUser):
    wait_time = between(1, 3)  # Wait 1-3 seconds between requests
    
    def on_start(self):
        """Login and get authentication token"""
        # Login credentials for company ID 8
        login_data = {
            "username": "Terminal Reality Admin",  # Updated username
            "password": "H@rri50n"  # Updated password
        }
        
        response = self.client.post("/api/login", json=login_data)
        if response.status_code == 200:
            data = response.json()
            if data.get("success"):
                self.token = data["data"]["token"]
                self.headers = {"Authorization": f"Bearer {self.token}"}
                print(f"✅ Login successful for company ID 8")
            else:
                print(f"❌ Login failed: {data.get('error', 'Unknown error')}")
                self.token = None
        else:
            print(f"❌ Login request failed with status {response.status_code}")
            self.token = None

    @task(3)
    def create_asset(self):
        """Create a single asset"""
        if not self.token:
            print("❌ No authentication token available")
            return
            
        # Kenyan schools in uppercase
        kenyan_schools = [
            "KENYATTA UNIVERSITY",
            "UNIVERSITY OF NAIROBI", 
            "JOMO KENYATTA UNIVERSITY OF AGRICULTURE AND TECHNOLOGY",
            "MOI UNIVERSITY",
            "EGERTON UNIVERSITY",
            "MASENO UNIVERSITY",
            "KENYATTA UNIVERSITY OF AGRICULTURE AND TECHNOLOGY",
            "UNIVERSITY OF ELDORET",
            "MULTIMEDIA UNIVERSITY OF KENYA",
            "TECHNICAL UNIVERSITY OF KENYA",
            "KENYA METHODIST UNIVERSITY",
            "CATHOLIC UNIVERSITY OF EASTERN AFRICA",
            "UNITED STATES INTERNATIONAL UNIVERSITY",
            "STRATHMORE UNIVERSITY",
            "DAYSTAR UNIVERSITY",
            "AFRICA NAZARENE UNIVERSITY",
            "SCOTT THEOLOGICAL COLLEGE",
            "GREAT LAKES UNIVERSITY OF KISUMU",
            "KABARAK UNIVERSITY",
            "MOUNT KENYA UNIVERSITY"
        ]
        
        # Kenyan departments
        departments = [
            "ICT",
            "INFORMATION TECHNOLOGY", 
            "COMPUTER SCIENCE",
            "ENGINEERING",
            "BUSINESS ADMINISTRATION",
            "ACCOUNTING",
            "FINANCE",
            "HUMAN RESOURCES",
            "MARKETING",
            "SALES",
            "OPERATIONS",
            "RESEARCH",
            "DEVELOPMENT",
            "QUALITY ASSURANCE",
            "CUSTOMER SERVICE",
            "ADMINISTRATION",
            "FACILITIES",
            "SECURITY",
            "MAINTENANCE",
            "LOGISTICS"
        ]
        
        # Asset types
        asset_types = [
            "LAPTOP",
            "DESKTOP",
            "PRINTER",
            "SCANNER",
            "PROJECTOR",
            "TELEVISION",
            "AIR CONDITIONER",
            "FURNITURE",
            "VEHICLE",
            "GENERATOR",
            "UPS",
            "NETWORK EQUIPMENT",
            "SOFTWARE",
            "LICENSE",
            "BOOKS",
            "LABORATORY EQUIPMENT",
            "MEDICAL EQUIPMENT",
            "SPORTS EQUIPMENT",
            "MUSICAL INSTRUMENTS",
            "OFFICE SUPPLIES"
        ]
        
        # Manufacturers
        manufacturers = [
            "DELL",
            "HP",
            "LENOVO",
            "APPLE",
            "SAMSUNG",
            "LG",
            "CANON",
            "EPSON",
            "BROTHER",
            "MICROSOFT",
            "CISCO",
            "INTEL",
            "AMD",
            "WESTERN DIGITAL",
            "SEAGATE",
            "KINGSTON",
            "LOGITECH",
            "PHILIPS",
            "PANASONIC",
            "SONY"
        ]
        
        # Locations
        locations = [
            "MAIN CAMPUS",
            "NORTH CAMPUS",
            "SOUTH CAMPUS",
            "EAST CAMPUS",
            "WEST CAMPUS",
            "ADMINISTRATION BLOCK",
            "LIBRARY",
            "LABORATORY",
            "COMPUTER LAB",
            "LECTURE HALL",
            "STAFF ROOM",
            "STUDENT CENTER",
            "CAFETERIA",
            "GYMNASIUM",
            "AUDITORIUM",
            "RESEARCH CENTER",
            "INNOVATION HUB",
            "BUSINESS SCHOOL",
            "ENGINEERING BLOCK",
            "SCIENCE BUILDING"
        ]
        
        # Generate random asset data
        school = random.choice(kenyan_schools)
        department = random.choice(departments)
        asset_type = random.choice(asset_types)
        manufacturer = random.choice(manufacturers)
        location = random.choice(locations)
        
        # Generate asset name
        asset_name = f"{asset_type} - {manufacturer} - {school[:10]}"
        
        # Generate serial number
        serial_number = f"{manufacturer[:3]}{random.randint(100000, 999999)}"
        
        # Generate model number
        model_number = f"{manufacturer[:3]}-{random.randint(1000, 9999)}"
        
        # Generate purchase date (within last 2 years)
        days_ago = random.randint(0, 730)
        purchase_date = (datetime.now() - timedelta(days=days_ago)).strftime("%Y-%m-%d")
        
        # Generate purchase price (between 10,000 and 500,000 KES)
        purchase_price = random.randint(10000, 500000)
        
        # Asset data
        asset_data = {
            "assetName": asset_name,
            "assetType": asset_type,
            "institutionName": school,
            "department": department,
            "functionalArea": f"{department} Department",
            "manufacturer": manufacturer,
            "modelNumber": model_number,
            "serialNumber": serial_number,
            "location": location,
            "status": random.choice(["Active", "Inactive", "Under Maintenance", "Retired"]),
            "purchaseDate": purchase_date,
            "purchasePrice": purchase_price
        }
        
        # Create asset
        response = self.client.post(
            "/addAsset",
            json=asset_data,
            headers=self.headers
        )
        
        if response.status_code == 200:
            data = response.json()
            if data.get("success"):
                print(f"✅ Asset created: {asset_name}")
            else:
                print(f"❌ Asset creation failed: {data.get('error', 'Unknown error')}")
        else:
            print(f"❌ Asset creation request failed with status {response.status_code}")
            print(f"Response: {response.text}")

    @task(1)
    def get_dashboard_stats(self):
        """Get dashboard statistics"""
        if not self.token:
            return
            
        response = self.client.get("/api/dashboard/stats", headers=self.headers)
        if response.status_code == 200:
            data = response.json()
            if data.get("success"):
                stats = data.get("data", {})
                total_assets = stats.get("total_assets", 0)
                print(f"📊 Dashboard stats - Total assets: {total_assets}")
            else:
                print(f"❌ Dashboard stats failed: {data.get('error', 'Unknown error')}")
        else:
            print(f"❌ Dashboard stats request failed with status {response.status_code}")

    @task(1)
    def get_trial_status(self):
        """Get trial status"""
        if not self.token:
            return
            
        response = self.client.get("/api/trial/status", headers=self.headers)
        if response.status_code == 200:
            data = response.json()
            if data.get("success"):
                trial_data = data.get("data", {})
                days_remaining = trial_data.get("days_remaining", 0)
                is_expired = trial_data.get("is_expired", False)
                print(f"⏰ Trial status - Days remaining: {days_remaining}, Expired: {is_expired}")
            else:
                print(f"❌ Trial status failed: {data.get('error', 'Unknown error')}")
        else:
            print(f"❌ Trial status request failed with status {response.status_code}")

    @task(1)
    def get_assets_list(self):
        """Get list of assets"""
        if not self.token:
            return
            
        response = self.client.get("/api/assets", headers=self.headers)
        if response.status_code == 200:
            data = response.json()
            if data.get("success"):
                assets = data.get("data", [])
                print(f"📋 Assets list - Count: {len(assets)}")
            else:
                print(f"❌ Assets list failed: {data.get('error', 'Unknown error')}")
        else:
            print(f"❌ Assets list request failed with status {response.status_code}")

class AssetBulkUser(HttpUser):
    """User class for bulk asset creation"""
    wait_time = between(0.5, 1.5)  # Faster requests for bulk operations
    
    def on_start(self):
        """Login and get authentication token"""
        login_data = {
            "username": "Terminal Reality Admin",  # Updated username
            "password": "H@rri50n"  # Updated password
        }
        
        response = self.client.post("/api/login", json=login_data)
        if response.status_code == 200:
            data = response.json()
            if data.get("success"):
                self.token = data["data"]["token"]
                self.headers = {"Authorization": f"Bearer {self.token}"}
                print(f"✅ Bulk user login successful")
            else:
                self.token = None
        else:
            self.token = None

    @task(1)
    def create_multiple_assets(self):
        """Create multiple assets at once"""
        if not self.token:
            return
            
        # Generate 5-10 assets at once
        num_assets = random.randint(5, 10)
        assets = []
        
        for i in range(num_assets):
            asset = {
                "assetName": f"Bulk Asset {i+1} - {random.choice(['LAPTOP', 'DESKTOP', 'PRINTER'])}",
                "assetType": random.choice(["LAPTOP", "DESKTOP", "PRINTER"]),
                "institutionName": random.choice([
                    "KENYATTA UNIVERSITY",
                    "UNIVERSITY OF NAIROBI",
                    "JOMO KENYATTA UNIVERSITY"
                ]),
                "department": random.choice(["ICT", "ENGINEERING", "BUSINESS"]),
                "functionalArea": "Bulk Import",
                "manufacturer": random.choice(["DELL", "HP", "LENOVO"]),
                "modelNumber": f"BULK-{random.randint(1000, 9999)}",
                "serialNumber": f"BULK{random.randint(100000, 999999)}",
                "location": "BULK IMPORT",
                "status": "Active",
                "purchaseDate": datetime.now().strftime("%Y-%m-%d"),
                "purchasePrice": random.randint(50000, 200000)
            }
            assets.append(asset)
        
        bulk_data = {
            "assets": assets
        }
        
        response = self.client.post(
            "/api/assets/multiple",
            json=bulk_data,
            headers=self.headers
        )
        
        if response.status_code == 200:
            data = response.json()
            if data.get("success"):
                print(f"✅ Bulk assets created: {len(assets)} assets")
            else:
                print(f"❌ Bulk assets creation failed: {data.get('error', 'Unknown error')}")
        else:
            print(f"❌ Bulk assets request failed with status {response.status_code}")

# Configuration for running the test
if __name__ == "__main__":
    print("🚀 Asset Tagging Load Test Script")
    print("=" * 50)
    print("This script will:")
    print("1. Login to company ID 8")
    print("2. Create assets with Kenyan schools")
    print("3. Use Gmail-style emails")
    print("4. Test dashboard and trial functionality")
    print("=" * 50)
    print("To run: locust -f locustfile.py --host=https://graf.moowigroup.com")
    print("Then open http://localhost:8089 to start the test") 