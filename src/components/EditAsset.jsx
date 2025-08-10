import React, { useState, useEffect } from 'react';
import axios from 'axios';
import { Form, Button } from 'react-bootstrap';

const EditAsset = ({ assetId }) => {
  const [assetData, setAssetData] = useState({ assetName: '', assetType: '', location: '' });

  useEffect(() => {
    const fetchAssetDetails = async () => {
      try {
        const response = await axios.get(`http://localhost:5000/api/assets/${assetId}`);
        setAssetData(response.data);
      } catch (error) {
        console.error('Error fetching asset details:', error);
      }
    };

    fetchAssetDetails();
  }, [assetId]);

  const handleSave = async () => {
    try {
      await axios.put(`http://localhost:5000/api/assets/${assetId}`, assetData);
      // Show success message or redirect to asset list
    } catch (error) {
      console.error('Error updating asset:', error);
      // Show error message
    }
  };

  return (
    <div>
      <h2>Edit Asset</h2>
      <Form>
        <Form.Group className="mb-3" controlId="formAssetName">
          <Form.Label>Asset Name</Form.Label>
          <Form.Control
            type="text"
            placeholder="Enter asset name"
            value={assetData.assetName}
            onChange={(e) => setAssetData({ ...assetData, assetName: e.target.value })}
          />
        </Form.Group>

        <Form.Group className="mb-3" controlId="formAssetType">
          <Form.Label>Asset Type</Form.Label>
          <Form.Control
            type="text"
            placeholder="Enter asset type"
            value={assetData.assetType}
            onChange={(e) => setAssetData({ ...assetData, assetType: e.target.value })}
          />
        </Form.Group>

        <Form.Group className="mb-3" controlId="formLocation">
          <Form.Label>Location</Form.Label>
          <Form.Control
            type="text"
            placeholder="Enter location"
            value={assetData.location}
            onChange={(e) => setAssetData({ ...assetData, location: e.target.value })}
          />
        </Form.Group>

        <Button variant="primary" onClick={handleSave}>
          Save
        </Button>
      </Form>
    </div>
  );
};

export default EditAsset;
