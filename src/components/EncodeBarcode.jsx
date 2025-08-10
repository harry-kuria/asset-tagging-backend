import React, { useState, useEffect, useRef } from 'react';
import { Form, Button, Col, Row, ListGroup } from 'react-bootstrap';
import AssetDetails from './AssetDetails'; // Create AssetDetails component separately
import Barcode from 'react-barcode';
import QRCode from 'qrcode.react';
import html2canvas from 'html2canvas';
import { saveAs } from 'file-saver';
import axios from 'axios';

const EncodeBarcode = () => {
  const [searchTerm, setSearchTerm] = useState('');
  const [searchResults, setSearchResults] = useState([]);
  const [assetDetails, setAssetDetails] = useState(null);
  const [barcodeType, setBarcodeType] = useState(null);
  const [barcodeData, setBarcodeData] = useState(null);
  // const [logo, setLogo] = useState(null);

  useEffect(() => {
    const fetchSearchResults = async () => {
      try {
        const response = await axios.get(`http://localhost:5000/api/searchAssets?term=${searchTerm}`);
        setSearchResults(response.data);
      } catch (error) {
        console.error('Error fetching search results:', error);
      }
    };

    if (searchTerm.trim() !== '') {
      fetchSearchResults();
    } else {
      setSearchResults([]);
    }
  }, [searchTerm]);

  const handleSearch = () => {
    // No need to implement anything here as the search is triggered by useEffect
  };

  const handleEncode = async (selectedBarcodeType) => {
    // Set the selected barcode type
    setBarcodeType(selectedBarcodeType);
  
    // Fetch asset details from the database based on assetDetails.id
    if (assetDetails) {
      try {
        // Make an API request to get detailed information based on asset ID
        const response = await axios.get(`http://localhost:5000/api/getAssetDetails?id=${assetDetails.id}`);
        const detailedInfo = response.data;
        const institutionName = response.data.institutionName.toUpperCase();
        // Extract relevant details for encoding
        const institutionShort = detailedInfo.institutionName.substring(0, 2).toUpperCase();
      const departmentShort = detailedInfo.department.substring(0, 2).toUpperCase();
      const functionalAreaShort = detailedInfo.functionalArea.substring(0, 2).toUpperCase();
      const assetNameShort = detailedInfo.assetName.replace(/\s/g, '').toUpperCase();
      const id = detailedInfo.id;
      // setLogo(detailedInfo.logo);

      // Create the formatted data for the barcode
      const dataForBarcode = `${institutionShort}/${departmentShort}/${functionalAreaShort}/${assetNameShort}/${id}`;
      setBarcodeData(dataForBarcode);
      } catch (error) {
        console.error('Error fetching asset details:', error);
      }
    }
  };

  const barcodeRef = useRef(null);

  const handleDownload = async () => {
    try {
      const canvas = await html2canvas(barcodeRef.current);
      const dataUrl = canvas.toDataURL('image/png');

      // Create a temporary link element to trigger the download
      const a = document.createElement('a');
      a.href = dataUrl;
      a.download = 'barcode.png';
      document.body.appendChild(a);
      a.click();
      document.body.removeChild(a);
    } catch (error) {
      console.error('Error capturing barcode:', error);
    }
  };
  useEffect(() => {
    // Implement barcode generation logic here when barcodeType changes
  }, [barcodeType]);

  const handleSelectAsset = (selectedAsset) => {
    setAssetDetails(selectedAsset);
    setSearchTerm('');
    setSearchResults([]);
  };

  return (
    <div className="container">
      <h2 className="mb-4">Encode Barcode</h2>

      <Form>
        <Form.Group controlId="searchTerm">
          <Form.Label>Search Asset</Form.Label>
          <Form.Control
            type="text"
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            placeholder="Enter asset name or ID"
          />
        </Form.Group>
      </Form>

      {searchResults.length > 0 && (
        <ListGroup>
          {searchResults.map((result) => (
            <ListGroup.Item
              key={result.id}
              action
              onClick={() => handleSelectAsset(result)}
              className="cursor-pointer"
            >
              {result.assetName}
            </ListGroup.Item>
          ))}
        </ListGroup>
      )}

{assetDetails && <AssetDetails details={assetDetails} />}
      {barcodeType && (
        <div>
          <h3 className="mb-3">Generated Barcode</h3>
          <div>
      {/* <p>{assetDetails.institutionName}</p> */}
    </div>
          <div ref={barcodeRef}>

          
    <img
      src={assetDetails.logo}
      alt="Logo"
      style={{ width: '50px', height: '50px', marginRight: '10px' }}
    />
  
          <p style={{ textTransform: 'uppercase' }}>{assetDetails.institutionName}</p>
            {barcodeType === 'CODE128' && barcodeData && <Barcode value={barcodeData} format={barcodeType} />}
            {barcodeType === 'QR-Code' && barcodeData && <QRCode value={barcodeData} />}
          </div>
          <Button onClick={handleDownload} className="btn-primary mt-3">
            Download Barcode
          </Button>
        </div>
      )}

    {assetDetails && (
      <div>
        <h3 className="mb-3">Encode Options</h3>
        <Button onClick={() => handleEncode('CODE128')} className="btn-primary me-3">
          Encode as EAN-13
        </Button>
        <Button onClick={() => handleEncode('QR-Code')} className="btn-primary">
          Encode as QR-Code
        </Button>
      </div>
    )}
    </div>
  );
};

export default EncodeBarcode;
