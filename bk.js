import React, { useState, useEffect } from 'react';
import { Form, Button, Col, Row } from 'react-bootstrap';
import axios from 'axios';
import DatePicker from 'react-datepicker';
import 'react-datepicker/dist/react-datepicker.css';
import * as XLSX from 'xlsx';

const AddAsset = () => {
  const [assetData, setAssetData] = useState({
    assetName: '',
    assetType: '',
    serialNumber: '',
    description: '',
    purchaseDate: '',
    purchasePrice: '',
    marketValue: '',
    manufacturer: '',
    modelNumber: '',
    location: '',
    status: '',
    barcode: '',
    institutionName: '', 
    department: '', 
    functionalArea: '',
    logo: null,
  });

  const [assetCategories, setAssetCategories] = useState([]);
  const [assetsArray, setAssetsArray] = useState([]);

  useEffect(() => {
    // Fetch asset categories when the component mounts
    const fetchAssetCategories = async () => {
      try {
        const response = await axios.get('http://localhost:5000/api/categories');
        setAssetCategories(response.data);
      } catch (error) {
        console.error('Error fetching asset categories:', error);
      }
    };

    fetchAssetCategories();
  }, []);

  

  const handleChange = (e) => {
    const { name, value } = e.target;
    setAssetData({ ...assetData, [name]: value });
  };
  const handleFileChange = (e) => {
    const file = e.target.files[0];
    if (file) {
      const reader = new FileReader();
      reader.onload = (e) => {
        const data = new Uint8Array(e.target.result);
        const workbook = XLSX.read(data, { type: 'array' });
        const sheetName = workbook.SheetNames[0];
        const sheet = workbook.Sheets[sheetName];
        const parsedData = XLSX.utils.sheet_to_json(sheet, { header: 1 });
        const headerMapping = {
          'NAME': 'assetName',
          'TYPE': 'assetType',
          'SERIAL NUMBER':'serialNumber',
          'DESCRIPTION': 'description',
          'PRICE': 'purchasePrice',
          'MARKET VALUE':'marketValue',
          'MANUFACTURER': 'manufacturer',	
          'MODEL NUMBER': 'modelNumber',
          'LOCATION': 'location',	
          'STATUS': 'status',	
          'BARCODE': 'barcode',
          'INSTITUTION': 'institutionName',
          'DEPARTMENT':'department',
          'FUNCTIONAL AREA': 'functionalArea',
          'REG. NO': 'vehicleregno',
          'SOURCE OF FUNDS': 'sourceoffunds',
          'ENGINE NO.': 'enginenumber',
          'CHASSIS NO': 'chassisnumber',
          'MAKE': 'make',
          'PURCHASE YEAR': 'purchaseyear',
          'PV NUMBER': 'pvnumber',
          'ORIGINAL LOCATION': 'originallocation',
          'CURRENT LOCATION': 'currentlocation',
          'REPLACEMENT DATE': 'replacementdate',
          'AMOUNT': 'amount',
          'DEPRECIATION RATE': 'depreciationrate',
          'ANNUAL DEPRECIATION': 'annualdepreciation',
          'ACCUMULATED DEPRECIATION': 'accumulateddepreciation',
          'NETBOOK VALUE': 'netbookvalue',
          'DISPOSAL DATE': 'disposaldate',
          'RESPONSIBLE OFFICER': 'responsibleofficer',
          'CONDITION': 'assetcondition',
        };
  
        // Assuming your Excel sheet has headers and follows a specific structure
        const [excelHeaders, ...rows] = parsedData;
        const assetsArray = rows.map((row) => {
          const assetObject = {};
          excelHeaders.forEach((header, index) => {
            const columnName = headerMapping[header] || header; // Use the mapping or the original header
            assetObject[columnName] = row[index];
          });
          return assetObject;
        });
  
        setAssetsArray(assetsArray);
      };
      reader.readAsArrayBuffer(file);
    }
  };

  const handleAddToDatabase = async () => {
    try {
      // Make a request to your backend to add the imported assets to the database
      const response = await axios.post('http://localhost:5000/api/addMultipleAssets', assetsArray);

      // Assuming your backend responds with a success message
      if (response.data.success) {
        // Reset the state or perform any other necessary actions
        setAssetsArray([]);

        // Display a success alert
        alert('Assets added to the database successfully!');
      } else {
        // Display an error alert if the backend indicates a failure
        alert('Failed to add assets. Please try again.');
      }
    } catch (error) {
      console.error('Error adding assets:', error);
      // Display an error alert for any unexpected errors
      alert('An unexpected error occurred. Please try again.');
    }
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
  
    try {
      if (assetsArray.length > 0) {
        // Importing multiple assets from Excel
  
        for (const asset of assetsArray) {
          const formattedDate = new Date(asset.purchaseDate)
            .toISOString()
            .slice(0, 19)
            .replace('T', ' ');
  
          const formData = new FormData();
          Object.entries(asset).forEach(([key, value]) => {
            if (key === 'logo') {
              formData.append('logo', value);
            } else {
              formData.append(key, value);
            }
          });
  
          const response = await axios.post('http://localhost:5000/api/addAsset', formData, {
            headers: {
              'Content-Type': 'multipart/form-data',
            },
          });
  
          if (response.data.success) {
            // Handle success as needed
          } else {
            console.error('Failed to add asset. Please try again.');
          }
        }
  
        // Display a success alert
        alert('Assets added successfully!');
      } else {
        // Adding a single asset via the form
  
        const dateObject = new Date(assetData.purchaseDate);
        const formattedDate = dateObject.toISOString().slice(0, 19).replace('T', ' ');
  
        const formData = new FormData();
        Object.entries(assetData).forEach(([key, value]) => {
          if (key === 'logo') {
            formData.append('logo', value);
          } else {
            formData.append(key, value);
          }
        });
  
        const response = await axios.post('http://localhost:5000/api/addAsset', formData, {
          headers: {
            'Content-Type': 'multipart/form-data',
          },
        });
  
        if (response.data.success) {
          // Reset the form or perform any other necessary actions
          setAssetData({
            assetName: '',
            assetType: '',
            serialNumber: '',
            description: '',
            purchaseDate: null,
            purchasePrice: '',
            marketValue: '',
            manufacturer: '',
            modelNumber: '',
            location: '',
            status: '',
            barcode: '',
            institutionName: '',
            department: '',
            functionalArea: '',
            logo: null,
          });
  
          // Display a success alert
          alert('Asset added successfully!');
        } else {
          // Display an error alert if the backend indicates a failure
          alert('Failed to add asset. Please try again.');
        }
      }
    } catch (error) {
      console.error('Error:', error);
      alert('An unexpected error occurred. Please try again.');
    }
  };
  
  

  return (
    <Form onSubmit={handleSubmit}>
      <h2 className="mb-4">Add New Asset</h2>
      <Row className="mb-3">
      <Form.Group as={Col} md="6" controlId="excelFile">
        <Form.Label>Import Excel File</Form.Label>
        <Form.Control type="file" accept=".xlsx" onChange={handleFileChange} />
      </Form.Group>

      {/* Button to add imported assets to the database */}
      {assetsArray.length > 0 && (
        <Button type="button" className="btn-primary" onClick={handleAddToDatabase}>
          Add Imported Assets to Database
        </Button>
      )}
      </Row>

      <Row className="mb-3">
        <Form.Group as={Col} md="6" controlId="assetName">
          <Form.Label>Asset Name</Form.Label>
          <Form.Control
            type="text"
            name="assetName"
            value={assetData.assetName}
            onChange={handleChange}
            placeholder="Enter asset name"
            required
          />
        </Form.Group>

        <Form.Group as={Col} md="6" controlId="assetType">
          <Form.Label>Asset Type</Form.Label>
          <Form.Control
            as="select"
            name="assetType"
            value={assetData.assetType}
            onChange={handleChange}
            required
          >
            <option value="">Select Asset Type</option>
            {assetCategories.map((category) => (
              <option key={category.id} value={category.category_name}>
                {category.category_name}
              </option>
            ))}
          </Form.Control>
        </Form.Group>
      </Row>

      <Row className="mb-3">
        {/* Add more Form.Group components for other fields */}
        {/* For example: */}
        <Form.Group as={Col} md="6" controlId="serialNumber">
          <Form.Label>Serial Number</Form.Label>
          <Form.Control
            type="text"
            name="serialNumber"
            value={assetData.serialNumber}
            onChange={handleChange}
            placeholder="Enter serial number"
            required
          />
        </Form.Group>

        <Form.Group as={Col} md="6" controlId="description">
          <Form.Label>Description</Form.Label>
          <Form.Control
            as="textarea"
            name="description"
            value={assetData.description}
            onChange={handleChange}
            placeholder="Enter asset description"
            required
          />
        </Form.Group>

        <Form.Group as={Col} md="6" controlId="purchaseDate">
          <Form.Label>Purchase Date</Form.Label>
          <DatePicker
            selected={assetData.purchaseDate}
            onChange={(date) => setAssetData({ ...assetData, purchaseDate: date })}
            dateFormat="MM/dd/yyyy"
            placeholderText="Select a date"
            className="form-control"
            required
          />
        </Form.Group>
        <Form.Group as={Col} md="6" controlId="purchasePrice">
          <Form.Label>Purchase Price</Form.Label>
          <Form.Control
            type="text"
            name="purchasePrice"
            value={assetData.purchasePrice}
            onChange={handleChange}
            placeholder="Enter Purchase Price"
            required
          />
        </Form.Group>
        <Form.Group as={Col} md="6" controlId="marketValue">
          <Form.Label>Market Value</Form.Label>
          <Form.Control
            type="text"
            name="marketValue"
            value={assetData.marketValue}
            onChange={handleChange}
            placeholder="Enter Market Value"
            required
          />
        </Form.Group>
        <Form.Group as={Col} md="6" controlId="manufacturer">
          <Form.Label>Manufacturer</Form.Label>
          <Form.Control
            type="text"
            name="manufacturer"
            value={assetData.manufacturer}
            onChange={handleChange}
            placeholder="Enter Manufacturer"
            required
          />
        </Form.Group>
        <Form.Group as={Col} md="6" controlId="modelNumber">
          <Form.Label>Model Number</Form.Label>
          <Form.Control
            type="text"
            name="modelNumber"
            value={assetData.modelNumber}
            onChange={handleChange}
            placeholder="Enter Model Number"
            required
          />
        </Form.Group>
        <Form.Group as={Col} md="6" controlId="institutionName">
          <Form.Label>Institution Name</Form.Label>
          <Form.Control
            type="text"
            name="institutionName"
            value={assetData.institutionName}
            onChange={handleChange}
            placeholder="Enter name of Institution"
            required
          />
        </Form.Group>
        <Form.Group as={Col} md="6" controlId="department">
          <Form.Label>Department</Form.Label>
          <Form.Control
            type="text"
            name="department"
            value={assetData.department}
            onChange={handleChange}
            placeholder="Enter Department"
            required
          />
        </Form.Group>
        <Form.Group as={Col} md="6" controlId="functionalArea">
          <Form.Label>Functional Area</Form.Label>
          <Form.Control
            type="text"
            name="functionalArea"
            value={assetData.functionalArea}
            onChange={handleChange}
            placeholder="Enter Functional Area"
            required
          />
        </Form.Group>
        <Form.Group as={Col} md="6" controlId="location">
          <Form.Label>Location</Form.Label>
          <Form.Control
            type="text"
            name="location"
            value={assetData.location}
            onChange={handleChange}
            placeholder="Enter Location"
            required
          />
        </Form.Group>
        <Form.Group as={Col} md="6" controlId="logo">
  <Form.Label>Logo</Form.Label>
  <Form.Control
    type="file"
    accept="image/*"
    onChange={(e) => setAssetData({ ...assetData, logo: e.target.files[0] })}
  />
</Form.Group>
      </Row>

      {/* Add more Form.Group components for other fields */}
      
      <Button type="submit" className="btn-primary" onClick={handleSubmit}>
        Add Asset
      </Button>
    </Form>
  );
};

export default AddAsset;

