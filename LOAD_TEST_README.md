# 🚀 Asset Tagging Load Test Script

This Locust script generates realistic assets for company ID 8 with Kenyan schools and Gmail-style data.

## 📋 Features

### ✅ **Realistic Data Generation**
- **Kenyan Schools**: 20+ major Kenyan universities in uppercase
- **Departments**: ICT, Engineering, Business, etc.
- **Asset Types**: Laptops, Desktops, Printers, etc.
- **Manufacturers**: Dell, HP, Lenovo, Apple, etc.
- **Locations**: Campus buildings, labs, offices
- **Pricing**: Realistic Kenyan pricing (10K-500K KES)

### ✅ **Test Scenarios**
- **Single Asset Creation**: Creates individual assets
- **Bulk Asset Creation**: Creates 5-10 assets at once
- **Dashboard Testing**: Checks dashboard statistics
- **Trial Status**: Monitors trial period status
- **Asset Listing**: Retrieves asset lists

## 🛠️ Installation

### 1. Install Python Dependencies
```bash
pip install -r requirements.txt
```

### 2. Verify Installation
```bash
locust --version
```

## 🎯 Usage

### Quick Start
```bash
# Run the load test
locust -f locustfile.py --host=https://graf.moowigroup.com

# Open browser and go to: http://localhost:8089
```

### Command Line Options
```bash
# Run with specific number of users
locust -f locustfile.py --host=https://graf.moowigroup.com --users 10 --spawn-rate 2

# Run headless (no web UI)
locust -f locustfile.py --host=https://graf.moowigroup.com --users 50 --spawn-rate 5 --run-time 5m --headless

# Save results to CSV
locust -f locustfile.py --host=https://graf.moowigroup.com --users 20 --spawn-rate 2 --run-time 10m --headless --csv=results
```

## 📊 Test Configuration

### User Classes
1. **AssetTaggingUser**: Main user class for single asset creation
2. **AssetBulkUser**: Bulk asset creation user

### Task Weights
- **Asset Creation**: 60% (3 out of 5 tasks)
- **Dashboard Stats**: 20% (1 out of 5 tasks)
- **Trial Status**: 20% (1 out of 5 tasks)
- **Asset Listing**: 20% (1 out of 5 tasks)

### Wait Times
- **Single Asset User**: 1-3 seconds between requests
- **Bulk Asset User**: 0.5-1.5 seconds between requests

## 🏫 Kenyan Schools Included

1. KENYATTA UNIVERSITY
2. UNIVERSITY OF NAIROBI
3. JOMO KENYATTA UNIVERSITY OF AGRICULTURE AND TECHNOLOGY
4. MOI UNIVERSITY
5. EGERTON UNIVERSITY
6. MASENO UNIVERSITY
7. UNIVERSITY OF ELDORET
8. MULTIMEDIA UNIVERSITY OF KENYA
9. TECHNICAL UNIVERSITY OF KENYA
10. KENYA METHODIST UNIVERSITY
11. CATHOLIC UNIVERSITY OF EASTERN AFRICA
12. UNITED STATES INTERNATIONAL UNIVERSITY
13. STRATHMORE UNIVERSITY
14. DAYSTAR UNIVERSITY
15. AFRICA NAZARENE UNIVERSITY
16. SCOTT THEOLOGICAL COLLEGE
17. GREAT LAKES UNIVERSITY OF KISUMU
18. KABARAK UNIVERSITY
19. MOUNT KENYA UNIVERSITY

## 🔧 Customization

### Update Login Credentials
Edit the `on_start()` method in both user classes:
```python
login_data = {
    "username": "your_username",  # Change this
    "password": "your_password"   # Change this
}
```

### Add More Schools
Add to the `kenyan_schools` list:
```python
kenyan_schools = [
    "YOUR NEW SCHOOL",
    # ... existing schools
]
```

### Modify Asset Types
Update the `asset_types` list:
```python
asset_types = [
    "YOUR NEW ASSET TYPE",
    # ... existing types
]
```

## 📈 Expected Results

### Asset Generation
- **Single Assets**: Realistic asset names with manufacturer and school
- **Serial Numbers**: Manufacturer prefix + 6 digits
- **Model Numbers**: Manufacturer prefix + 4 digits
- **Purchase Dates**: Random dates within last 2 years
- **Prices**: 10,000 - 500,000 KES

### Example Generated Assets
```
LAPTOP - DELL - KENYATTA U
DESKTOP - HP - UNIVERSITY O
PRINTER - CANON - JOMO KENY
```

## 🚨 Important Notes

### Authentication
- Script logs in as admin user for company ID 8
- Uses JWT token authentication
- Token expires after 24 hours

### Trial System
- Script tests trial status functionality
- Will show trial expiration if applicable
- Payment prompts may appear if trial expired

### Rate Limiting
- Respects server rate limits
- Includes realistic wait times
- Avoids overwhelming the server

## 🔍 Monitoring

### Console Output
```
✅ Login successful for company ID 8
✅ Asset created: LAPTOP - DELL - KENYATTA U
📊 Dashboard stats - Total assets: 15
⏰ Trial status - Days remaining: 25, Expired: False
📋 Assets list - Count: 15
```

### Locust Web UI
- Real-time metrics
- Response time charts
- Error rate monitoring
- User count tracking

## 🐛 Troubleshooting

### Common Issues
1. **Login Failed**: Check username/password
2. **500 Errors**: Check server status
3. **403 Forbidden**: Check trial status
4. **Network Errors**: Check host URL

### Debug Mode
```bash
# Run with verbose output
locust -f locustfile.py --host=https://graf.moowigroup.com --loglevel=DEBUG
```

## 📞 Support

For issues with the load test script:
1. Check console output for error messages
2. Verify login credentials
3. Ensure server is accessible
4. Check trial status and payment requirements

---

**Happy Load Testing! 🎉** 